# Pressluft

Self-hosted WordPress hosting panel for web agencies.

> Experimental project. The platform contract below is now the source of truth for what Pressluft actually supports today.

![Screenshot](screenshot.png)

## What Pressluft is aiming at

Professional WordPress hosting platforms like WP Engine, Cloudways, Kinsta, and Raidboxes give agencies a complete workflow: staging, push-to-live, server management, monitoring, backups, updates, and more.

Pressluft is an attempt to build that experience as open-source software you run yourself, on infrastructure you own.

## Where the repository is today

- Hetzner Cloud is the only provider with real server lifecycle integration.
- The control plane can queue jobs, run Ansible playbooks, and talk to agents over WebSockets.
- The agent can report heartbeats and execute remote service restarts.
- `nginx-stack` is the only profile that currently converges a verified stack baseline end-to-end.
- `openlitespeed-stack` and `woocommerce-optimized` remain listed as internal identifiers, but they are intentionally unavailable until they gain real convergence and verification.
- Production bootstrap now has one supported path: HTTPS registration followed by `wss` + mTLS reconnect.
- Runtime logs now use `server_id`, `job_id`, and `command_id` consistently so command dispatch, job events, and activity entries can be correlated during incident review.

## Platform Contract

This section is Milestone 0's platform-layer contract. Later runtime work should reference this section, `docs/internal/lifecycle-state-semantics.md`, and `docs/glossary.md` instead of inventing new meanings.

### Execution modes

- `dev`: Go binaries built with the `dev` build tag. Agent trust uses a dev WebSocket token. TLS is not required. This is the default local contributor workflow.
- In `dev`, `make dev` will start a Cloudflare quick tunnel when `PRESSLUFT_CONTROL_PLANE_URL` is unset. That path is intentionally ephemeral: remote agents configured against a `*.trycloudflare.com` callback URL will not reconnect after control-plane restart.
- Durable dev reconnect requires setting `PRESSLUFT_CONTROL_PLANE_URL` to a stable public hostname. The recommended approach is a named Cloudflare Tunnel bound to fixed DNS.
- `single-node-local-control-plane`: non-`dev` control-plane binary running locally or on one node for platform work. Hetzner workflows can run, but agent bootstrap is intentionally disabled in this mode.
- `production-bootstrap`: non-`dev` agent and control-plane contract for production bootstrap. The control plane must start with HTTPS enabled, agents register once over HTTPS with a registration token, then reconnect over `wss` with mTLS.

### Canonical server lifecycle states

- `pending`: server record exists and a provisioning job has been accepted, but infrastructure has not been proven yet.
- `provisioning`: provider-side create is actively running or has returned enough information to continue.
- `configuring`: Pressluft is converging the server after provider-side provisioning or rebuild.
- `rebuilding`: a destructive rebuild job is queued or running.
- `resizing`: a destructive resize job is queued or running.
- `deleting`: a delete job is queued or running; the record remains until provider-side removal succeeds or fails.
- `deleted`: provider-side deletion completed and the record is retained as a tombstone.
- `ready`: the latest platform job completed successfully and the control plane believes the server is available for the currently implemented contract. For the supported `nginx-stack` profile this now means provider provisioning, baseline convergence, profile role execution, and post-config verification all succeeded. For unavailable profiles, configure is blocked before readiness can be implied.
- `failed`: the latest platform job failed and human attention is required.

Server records also expose setup truth separately from lifecycle truth:

- `setup_state=not_started`: no setup phase has run yet
- `setup_state=running`: setup/bootstrap is in progress
- `setup_state=degraded`: provider-backed machine exists, but setup failed and needs attention
- `setup_state=ready`: the latest setup phase completed successfully

`configured` is not a separate durable server status. In the current contract, configure success is folded into `ready`.

`deleted` is now a durable tombstone status. Pressluft keeps the record for audit history instead of silently deleting the database row.

### Canonical node status states

- `online`: the agent has an active WebSocket session and recent heartbeats.
- `unhealthy`: the node was connected recently, but the session closed or heartbeat freshness degraded past 45 seconds.
- `offline`: the last durable heartbeat is older than 150 seconds and the node is not currently reachable through the agent channel.
- `unknown`: no durable node health has been established yet.

Exact node-status transitions in the current runtime contract:

- connect or heartbeat -> `online`
- websocket/session loss -> `unhealthy`
- no heartbeat for more than 150 seconds -> `offline`
- reconnect from `unhealthy` or `offline` -> `online`

### Canonical job kinds and states

- Supported job kinds: `provision_server`, `configure_server`, `delete_server`, `rebuild_server`, `resize_server`, `update_firewalls`, `manage_volume`, `restart_service`.
- Supported durable job statuses: `queued`, `running`, `succeeded`, `failed`.
- Every currently supported job kind may only enter those four states.
- Jobs move directly from `queued` to `running` when claimed. If the worker restarts or the kind-specific timeout expires before a terminal result is recorded, the job is marked `failed` with an explicit recovery reason instead of being silently retried.

### Job timeout and retry policy

| Job kind | Timeout | Automatic retries | Recovery behavior |
| --- | --- | --- | --- |
| `provision_server` | 30 minutes | none | mark failed; inspect provider state before manual retry |
| `configure_server` | 30 minutes | none | mark failed; retry setup manually after inspection |
| `delete_server` | 20 minutes | none | mark failed; verify provider deletion state before manual retry |
| `rebuild_server` | 45 minutes | none | mark failed; inspect machine state before manual retry |
| `resize_server` | 20 minutes | none | mark failed; inspect resize state before manual retry |
| `update_firewalls` | 15 minutes | none | mark failed; retry manually after inspection |
| `manage_volume` | 20 minutes | none | mark failed; retry manually after inspection |
| `restart_service` | 2 minutes | none | mark failed; late agent results are ignored |

### Ready semantics

`server ready` means all currently implemented steps for the latest server workflow completed successfully:

- provider provisioning returned a usable machine for provision jobs
- configure deployed the Pressluft agent and recorded the selected profile contract
- no job step in the current workflow ended in `failed`

If provider provisioning succeeds but setup fails, the server remains provider-backed and actionable while `setup_state` becomes `degraded`.

`ready` is now stronger for the supported baseline profile: Pressluft sets it only after the selected configure pipeline verifies required services, listeners, config files, and health checks. Profile breadth remains intentionally narrow.

### Profile guarantees

- Profile names such as `nginx-stack`, `openlitespeed-stack`, and `woocommerce-optimized` are internal platform identifiers, not proof of application support.
- `nginx-stack` is `supported`: configure applies the common Ubuntu 24.04 baseline, deploys the agent, installs and enables NGINX/PHP-FPM/Redis, manages the Pressluft landing site config, flushes handlers, and verifies services, files, listeners, and health checks before marking the server `ready`.
- `openlitespeed-stack` is `unavailable`: selection is blocked until Pressluft has a real OpenLiteSpeed role and verification contract.
- `woocommerce-optimized` is `unavailable`: selection is blocked until Pressluft has a real commerce-specific convergence and verification contract.

### Trust model

- `dev`: agent connects over `ws` using a dev WebSocket token delivered during local configure.
- `production-bootstrap`: agent registers once with a registration token, receives a certificate signed by the Pressluft CA, stores the keypair locally, clears the bootstrap token from disk, and reconnects over `wss` with mTLS.
- Registration tokens are validated before use and consumed only inside the certificate persistence transaction, so CA or storage failures do not burn the token.
- Existing valid certificates block duplicate registration. Reissue is only allowed with a fresh token when the current certificate is within the reissue window.

### Destructive-action semantics

- `delete_server`: soft-delete requested -> delete job queued -> provider-side deprovision attempted -> durable `deleted` tombstone on success.
- `rebuild_server`: destructive replacement of the machine image while preserving the Pressluft identity; the server returns to `ready` only after reconfigure succeeds.
- `resize_server`: disruptive provider-side resize; destructive actions are serialized so only one delete/rebuild/resize workflow may be active per server.

### Provider support

- `Hetzner Cloud`: the only provider with real server lifecycle support today.
- Other providers: unsupported until they implement truthful provisioning, lifecycle actions, and catalog integration.

### Experimental behavior labels

Treat these as explicitly experimental today:

- profile-specific convergence claims
- firewall and volume workflows beyond their current happy-path job execution
- any UI wording that implies production readiness or provider breadth beyond Hetzner

## Security

- SSH private keys and the CA private key are encrypted at rest with [age](https://age-encryption.org).
- The control plane acts as its own certificate authority.
- Dev mode uses a dev WebSocket token instead of certificates.
- Production transport requires `PRESSLUFT_EXECUTION_MODE=production-bootstrap`, `PRESSLUFT_CONTROL_PLANE_URL=https://...`, `PRESSLUFT_TLS_CERT_FILE`, and `PRESSLUFT_TLS_KEY_FILE`. In this mode the server starts HTTPS in-process and the agent path requires `wss` plus mTLS.

## Configuration and Env Vars

Pressluft currently reads a small set of environment variables directly from the Go binaries. Everything not listed here is either compiled in, passed through the database, or delivered via the agent YAML config during configure.

### Control plane

| Variable | Required | Meaning |
| --- | --- | --- |
| `PRESSLUFT_EXECUTION_MODE` | recommended | Execution mode contract. Use `dev`, `single-node-local-control-plane`, or `production-bootstrap`. Empty resolves to the build-tag default. |
| `PRESSLUFT_CONTROL_PLANE_URL` | yes for real agent bootstrap and configure | Public base URL the agent uses for registration and websocket reconnect. Must be `https://...` in `production-bootstrap`. |
| `PRESSLUFT_TLS_CERT_FILE` | yes in `production-bootstrap` | HTTPS certificate file for the control plane listener. |
| `PRESSLUFT_TLS_KEY_FILE` | yes in `production-bootstrap` | HTTPS private key file for the control plane listener. |
| `PRESSLUFT_DB` | optional | SQLite database path. Defaults under the Pressluft data directory. |
| `PORT` | optional | HTTP listen port. Defaults to `8080`. |
| `XDG_DATA_HOME` | optional | Base data directory when `PRESSLUFT_DB` or `PRESSLUFT_CA_KEY_PATH` are not set. |
| `PRESSLUFT_CA_KEY_PATH` | optional | Path for the encrypted Pressluft CA private key. Defaults under the data directory. |
| `PRESSLUFT_AGE_KEY_PATH` | optional | Path for the age identity used to encrypt SSH keys and the CA key at rest. Defaults to the Pressluft data directory and is auto-generated locally when missing. |
| `PRESSLUFT_ANSIBLE_DIR` | optional | Working directory used to resolve the Ansible virtualenv and playbook paths. Defaults to the current working directory. |
| `PRESSLUFT_ANSIBLE_BIN` | optional | Absolute or `PRESSLUFT_ANSIBLE_DIR`-relative path to `ansible-playbook`. Defaults to `.venv/bin/ansible-playbook`. |

### Agent process

| Variable | Required | Meaning |
| --- | --- | --- |
| `PRESSLUFT_EXECUTION_MODE` | recommended | Agent-side execution mode. `dev` uses the dev websocket token flow; `production-bootstrap` uses HTTPS registration then `wss` + mTLS reconnect. |

### Agent YAML config

The agent runtime contract is driven primarily by the YAML file rendered during configure, not by process environment variables.

| Field | Required | Meaning |
| --- | --- | --- |
| `server_id` | yes | Durable Pressluft server identifier used for registration and log correlation. |
| `control_plane` | yes | Control-plane base URL. `ws://` is allowed only in `dev`; `https://` is required for production bootstrap. |
| `cert_file` | yes in `production-bootstrap` | Client certificate path. |
| `key_file` | yes in `production-bootstrap` | Client private key path. |
| `ca_cert_file` | yes in `production-bootstrap` | Control-plane CA certificate path for HTTPS and `wss` trust. |
| `data_dir` | yes | Agent state directory on disk. |
| `registration_token` | bootstrap-only | One-time token used only until certificate issuance succeeds or reissue is requested. |
| `dev_ws_token` | yes in `dev` | Dev websocket bearer token used instead of certificates. |

Required local dependencies still remain:

- `go` 1.24+
- `node` 20+
- `pnpm`
- `cloudflared`
- `ansible-playbook` available through `.venv` or `PRESSLUFT_ANSIBLE_BIN`

## Contract References

- Lifecycle and state note: `docs/internal/lifecycle-state-semantics.md`
- Shared glossary: `docs/glossary.md`
- Bootstrap sequence note: `docs/internal/agent-bootstrap-sequence.md`
- Ansible playbook notes: `ops/ansible/playbooks/README.md`
- Frontend contract notes: `web/README.md`

## Runtime Hygiene Notes

- `internal/agentproto` was removed because its old heartbeat/checkpoint types were unused and described a job-execution contract the runtime no longer implements.
- The old direct `ServerStore.Delete` helper was removed so the storage layer no longer implies that deleting a server can bypass orchestration.
- `internal/dispatch/dispatcher.go` is no longer part of the live tree; dispatch now flows through `internal/dispatch/agent_runner.go`, `internal/ws/handler.go`, and `internal/dispatch/completer.go`.

## Profile Support Matrix

| Profile key | Level | Current guarantee |
| --- | --- | --- |
| `nginx-stack` | supported | Verified baseline convergence on Ubuntu 24.04 with agent, firewall/hardening baseline, NGINX, PHP-FPM, Redis, managed config files, and post-config verification before `ready` |
| `openlitespeed-stack` | unavailable | Listed only as an internal identifier; selection is blocked |
| `woocommerce-optimized` | unavailable | Listed only as an internal identifier; selection is blocked |

## Get involved

The codebase is split into three main areas:

- `web/` - Nuxt dashboard
- `internal/` - Go backend, jobs, agent communication, trust, and persistence
- `ops/` - Ansible playbooks, profiles, and infrastructure artifacts

Open an issue, start a discussion, send a PR, or write to [deniz@bombig.net](mailto:deniz@bombig.net).

## Get started

Use a Unix-like environment. Development is tested on Ubuntu 24.04 in WSL2 on Windows 11.

Required tools:

- [Go 1.24+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [pnpm](https://pnpm.io/installation)
- [cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/)

Set up dependencies from the repository root:

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install ansible
ansible-galaxy collection install -r ops/ansible/requirements.yml
```

## Development

```bash
make dev
```

This starts the Go backend on `http://localhost:8081`, the Nuxt dev server on `http://localhost:8080`, and a Cloudflare quick tunnel so provisioned servers can reach your local control plane.

## Verification

Contributor verification is now split into fast local gates and heavier smoke coverage.

Fast local validation:

```bash
make check
```

This runs format drift checks, `go vet`, unit tests, uncached integration/store tests, Ansible syntax checks, profile schema validation with the named `santhosh-tekuri/jsonschema` Draft 2020-12 validator, profile consistency validation, and the production build.

Heavier disposable-environment smoke validation:

```bash
make smoke
```

The smoke scripts live in `ops/tests/smoke/` and cover the supported happy path in order:

- provider setup
- server provision
- configure completion
- agent register/connect verification
- restart service
- delete server

`make ansible-check` remains advisory only because `ansible-playbook --check --diff` is not proof of convergence.

## Building

```bash
make build
make agent
make all
```
