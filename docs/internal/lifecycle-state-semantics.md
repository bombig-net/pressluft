# Lifecycle and State Semantics

This note is the internal design reference for Milestone 0. Future platform milestones should reuse these meanings instead of introducing alternate interpretations.

## Why this exists

Pressluft had drift between prototype behavior, dev-only assumptions, and production language. This document freezes the current contract so later lifecycle, trust, and convergence work can tighten behavior without changing terminology every milestone.

## Execution modes

- `dev`: local contributor mode, compiled with the `dev` build tag. Agent trust uses a dev WebSocket token.
- `single-node-local-control-plane`: non-`dev` control-plane runtime for local or one-node platform work. Useful for real provider interactions, but agent bootstrap is intentionally disabled here.
- `production-bootstrap`: supported bootstrap path for non-`dev` agents. Registration is HTTPS with a bootstrap token and CSR; steady-state transport is `wss` plus mTLS.

## Server lifecycle

- `pending`: record created, job accepted, no durable infrastructure proof yet.
- `provisioning`: provider create is actively running or has returned enough information to continue.
- `configuring`: Pressluft is running post-provider convergence such as agent deployment after provision or rebuild.
- `rebuilding`: a destructive rebuild workflow is queued or running.
- `resizing`: a destructive resize workflow is queued or running.
- `deleting`: provider-side deletion has been requested and is still in flight.
- `deleted`: provider-side deletion completed and the record is retained as a tombstone.
- `ready`: latest implemented workflow completed successfully for the current contract.
- `failed`: latest workflow failed and requires attention.

Design rule: `ready` is a capability claim, not just an absence of errors. For Milestone 0 that capability claim is intentionally narrow: provider provisioning plus agent deployment. Later milestones may strengthen the proof required before setting `ready`, but must not silently change the name.

`configured` is intentionally not its own durable server status. It is a workflow phase whose successful completion currently rolls up into `ready`.

Setup truth is persisted separately from lifecycle truth:

- `not_started`: no setup workflow has run yet
- `running`: setup/bootstrap is in progress
- `degraded`: provider-backed machine exists, but setup failed and needs attention
- `ready`: latest setup workflow completed successfully

`deleted` is the canonical terminal meaning for a removed server and is now a durable tombstone status.

## Node lifecycle

- `online`: connected and heartbeating.
- `unhealthy`: the control plane recently saw the node, but the session is degraded or disconnected.
- `offline`: disconnected and stale beyond the offline threshold.
- `unknown`: no durable observation yet.

Node status is about the agent channel only. It does not prove anything about server convergence or application health.

Current transition rules:

- a successful websocket connect or heartbeat writes `online`
- websocket/session failure writes `unhealthy`
- monitor passes degrade stale connections to `unhealthy` after 45 seconds without heartbeat
- monitor marks stale durable node records `offline` after 150 seconds without heartbeat

## Job lifecycle

Durable job statuses currently in contract:

- `queued`: accepted and waiting for worker claim.
- `running`: workflow logic or command dispatch is actively executing.
- `succeeded`: terminal success.
- `failed`: terminal failure.

The worker claims jobs by moving them straight from `queued` to `running` and attaching a kind-specific timeout deadline.

Automatic retries are intentionally disabled for all currently supported job kinds. If a worker restarts mid-job or a timeout deadline is reached, Pressluft records an explicit `failed` result and a recovery event instead of silently replaying the workflow.

States like `verifying`, `retrying`, `waiting_reboot`, and `timed_out` are intentionally outside the current runtime contract. Timeouts are recorded as terminal failure events, not separate durable statuses.

## Activity events vs job events

- Activity events are account-facing audit entries.
- Job events are ordered timeline entries attached to one job.

They can describe the same workflow, but they are not interchangeable. Activity events summarize platform truth for operators; job events provide step-by-step execution history. Both are emitted from the same job transition points so terminal activity cannot claim success or failure without a matching job timeline event.

## Ready and convergence

Current `ready` semantics depend on profile support level.

- Ready proves the latest implemented platform workflow completed successfully.
- For the supported `nginx-stack` profile, ready now includes baseline convergence plus post-config verification for services, files, listeners, and health checks.
- Unavailable profiles are blocked before configure can imply readiness at all.

This distinction matters because the UI and docs must not treat `ready` as if it already means `fully converged hosting baseline`.

## Destructive actions

- `delete_server`: intended to mean provider-side removal.
- `configure_server`: intended to mean post-provider setup/bootstrap/profile convergence on an existing machine.
- `rebuild_server`: intended to mean destructive replacement of the machine image.
- `resize_server`: intended to mean disruptive provider-side capacity change.

All three are asynchronous jobs. None should be described as synchronous API actions.

Current lifecycle contract for destructive actions:

- request accepted -> server moves into an in-progress durable status (`deleting`, `rebuilding`, or `resizing`)
- worker attempts the provider-side action through a job
- duplicate destructive actions are rejected while one is queued or running
- success moves the server to `deleted` (for delete) or back to `ready` after any required follow-up configure work
- failure moves the server to `failed`, preserving a recovery path through operator retry or follow-up action

## Timeout and recovery policy

| Job kind | Timeout | Automatic retries | Recovery behavior |
| --- | --- | --- | --- |
| `provision_server` | 45 minutes | none | mark failed; inspect provider state before manual retry |
| `delete_server` | 20 minutes | none | mark failed; verify provider deletion state before manual retry |
| `rebuild_server` | 45 minutes | none | mark failed; inspect machine state before manual retry |
| `resize_server` | 20 minutes | none | mark failed; inspect resize state before manual retry |
| `update_firewalls` | 15 minutes | none | mark failed; retry manually after inspection |
| `manage_volume` | 20 minutes | none | mark failed; retry manually after inspection |
| `restart_service` | 2 minutes | none | mark failed; late agent results are ignored |

## Trust semantics

- Dev trust is token-based and intentionally local/developer-friendly.
- Production trust is certificate-based, single-registration by bootstrap token, then mTLS reconnect.
- A registration token is bootstrap-only; a certificate is the durable identity.
- Token validation happens before consumption, and token consumption is committed only with certificate persistence so bootstrap failures do not lock the node out.

## Support labels

Use these labels consistently:

- `supported`: runtime behavior is intentionally implemented and should be relied on.
- `experimental`: surfaced for testing or development, but not yet operationally trustworthy.
- `unavailable`: intentionally not supported and should not be implied by docs or UI.
