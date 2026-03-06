# Platform Stabilization Plan

## Goal

Stabilize the Pressluft platform layer so the system is operationally truthful, testable, and safe enough to become the foundation for future WordPress and site-level features.

This plan assumes the current near-term target is still:

- prove the control-plane -> job -> agent loop
- make provisioning, configuration, and deletion trustworthy
- turn profile scaffolding into a real convergence layer
- establish enough automated validation to safely build on top

## What "stabilized" means

Before any WordPress or site product work starts, the platform must satisfy all of these:

- A provisioned server can reliably bootstrap an agent and maintain a trusted connection.
- Server lifecycle actions are truthful: create, configure, rebuild, resize, and delete all reflect reality.
- Job state, retries, recovery, and timeline events are durable and consistent.
- Selected profiles actually converge machines into known-good baselines.
- Agent commands are validated, observable, and failure-safe.
- Node health and status in the UI reflect durable backend truth.
- The highest-risk platform paths have tests and smoke validation.
- Docs and runtime contracts match actual behavior.

## Non-goals

These are explicitly out of scope until this plan is done:

- WordPress install and site creation flows
- staging and push-to-live workflows
- backup, migration, and update orchestration
- billing, tenancy, RBAC, and multi-user features
- advanced provider abstraction beyond what is needed to stabilize current flows
- WordPress and WooCommerce compatibility validation, tuning, and server-specific rewrite behavior

## Planning Principles

These rules should guide implementation decisions while this plan is in flight:

- Prefer honesty over sophistication. Remove misleading capability before adding new abstraction.
- Stabilize one real path end-to-end before broadening support.
- Do not preserve fake completeness. Experimental or partial behavior should be clearly marked or blocked.
- Preserve dev velocity, but not by hiding broken production assumptions.
- Platform truth comes before UI polish.
- Every milestone should leave the system easier to test, reason about, and recover.

## Workstream Overview

1. Agent bootstrap, trust, and transport hardening
2. Lifecycle truthfulness for server actions
3. Job and orchestration contract simplification and recovery
4. Profile execution and post-config verification
5. Agent command and runtime hardening
6. Observability, docs, and dead-code cleanup
7. Automated validation and release gates

---

## Program Rules

These are the working rules for executing this plan:

- No new WordPress or site feature work lands until the exit gate is green.
- Changes that narrow scope or remove misleading behavior are allowed and encouraged.
- If a profile is not real, it should be downgraded, hidden, or marked experimental.
- If production mode is not truly supported, it must be hard-gated rather than implied.
- Each milestone should produce updated docs and tests alongside code changes.
- Each milestone should end with an explicit "what is now true" note in the plan or changelog.

---

## Milestone 0 - Freeze the Platform Contract

### Outcome

Establish one explicit platform-layer contract so implementation work stops drifting across "prototype", "dev-only", and "production" assumptions.

### Files and areas

- `README.md`
- `cmd/main.go`
- `cmd/pressluft-agent/main.go`
- `internal/orchestrator/types.go`
- `internal/server/profiles/registry.go`
- `ops/ansible/playbooks/`
- `web/README.md`

### Problems to solve

- Core terms like `ready`, `online`, `configured`, and `deleted` are not defined tightly enough.
- Dev and production assumptions are mixed in code and docs.
- Supported platform guarantees are broader in language than they are in runtime behavior.

### Tasks

- Define supported execution modes:
  - dev mode
  - single-node local control plane
  - production bootstrap path
- Define canonical server lifecycle states and meanings.
- Define canonical job kinds and the states each kind may enter.
- Define what `server ready` means.
- Define what a profile guarantees after successful configure.
- Define the supported agent trust model in dev vs production.
- Define destructive-action semantics for delete, rebuild, and resize.
- Document current provider support truthfully.
- Decide which currently implemented behavior is experimental and must be labeled as such.

### Deliverables

- Platform contract section in `README.md`
- Internal design note for lifecycle and state semantics
- Short glossary for:
  - server status
  - node status
  - job status
  - activity event
  - profile convergence
  - registration token
  - dev ws token
  - certificate

### Acceptance criteria

- No core runtime behavior depends on undocumented assumptions.
- `ready`, `online`, `configured`, and `deleted` all have one meaning.
- Future milestones can reference a stable contract instead of inventing semantics ad hoc.

### Exit check

- One written contract exists and is referenced by later implementation work.

---

## Milestone 1 - Fix Agent Bootstrap and Trust

### Outcome

A newly provisioned node can always bootstrap correctly, register once, persist usable credentials, and reconnect securely.

### Files and areas

- `internal/agent/agent.go`
- `internal/agent/register.go`
- `internal/agent/config.go`
- `internal/server/handler_nodes.go`
- `internal/server/handler_ws_mtls.go`
- `internal/server/handler_ws_dev.go`
- `internal/registration/store.go`
- `internal/pki/ca.go`
- `internal/pki/store.go`
- `ops/ansible/playbooks/configure.yml`
- `ops/ansible/playbooks/templates/agent-config.yaml.j2`
- `cmd/main.go`

### Problems to solve

- Production registration and bootstrap path is incomplete.
- Registration token flow is not fully wired into provisioning and configure.
- Registration token is consumed too early in node registration.
- Agent registration currently passes an empty config path when clearing token.
- Private key persistence format is fragile for later TLS loading.
- mTLS websocket handler requires TLS, but the main server starts plain HTTP only.

### Tasks

- Define one complete production bootstrap sequence:
  - server created
  - registration token generated
  - token delivered during configure
  - agent starts
  - agent registers with CSR
  - cert issued and stored
  - token invalidated only after successful certificate issuance
  - agent reconnects via mTLS
- Fix registration transaction semantics:
  - validate token and CSR before token consumption
  - guarantee single-use behavior without accidental lockout
- Fix agent local persistence:
  - store private key as PEM-encoded key material
  - prefer PKCS#8 in a `PRIVATE KEY` PEM block
  - ensure `tls.LoadX509KeyPair` works reliably
  - clear registration token from the real config path
- Make the control-plane transport contract real:
  - either add proper TLS server startup for production
  - or explicitly gate production mode until TLS is configured
  - require `wss` plus mTLS for the production agent transport path
  - define client and server trust configuration explicitly
- Make dev and production bootstrap code paths intentionally separate, not half-shared.
- Add certificate rotation and reissue policy, even if minimal.
- Define reconnect behavior after abnormal websocket closure:
  - jittered backoff
  - bounded retry behavior
- Add clean failure behavior when:
  - token is expired
  - CSR CN mismatches
  - server already has a valid cert
  - CA signing fails
  - cert persistence fails

### Deliverables

- Reliable dev bootstrap flow
- Reliable production bootstrap flow
- Documented startup requirements for production TLS
- Bootstrap sequence diagram in docs

### Acceptance criteria

- Fresh provisioned server registers successfully end-to-end.
- Rebooted agent reconnects without re-registration.
- Invalid or replayed registration token is rejected.
- Existing valid cert blocks duplicate registration safely.
- Production mode cannot silently run in an impossible transport configuration.

### Tests

- Unit tests for `internal/agent/register.go`
- Unit tests for `internal/server/handler_nodes.go`
- Tests for token consume, replay, and expiry behavior
- Tests for persisted keypair and cert loadability
- Smoke test: provision -> configure -> register -> reconnect

### Exit check

- A fresh node can bootstrap and reconnect with no manual credential surgery.

---

## Milestone 2 - Make Lifecycle Actions Truthful

### Outcome

Every server action shown in the UI and API corresponds to the real infrastructure state.

### Files and areas

- `internal/server/handler_servers.go`
- `internal/worker/executor.go`
- `internal/server/store_servers.go`
- `internal/provider/`
- `ops/ansible/playbooks/delete_server.yml`
- `ops/ansible/playbooks/rebuild_server.yml`
- `ops/ansible/playbooks/resize_server.yml`

### Problems to solve

- API delete currently removes DB records directly instead of orchestrating real deletion.
- Real deletion logic exists elsewhere, creating semantic mismatch.
- Server status may be set back to `ready` too simplistically after workflows.
- Lifecycle APIs do not consistently represent asynchronous orchestration.

### Tasks

- Replace direct delete behavior with orchestration-backed deletion.
- Decide and implement one deletion contract:
  - soft-delete requested -> delete job queued -> provider deletion -> DB tombstone or final removal
- Add guardrails for destructive actions:
  - reject duplicate delete, rebuild, or resize jobs for the same server
  - reject actions on already deleting or deleted servers
- Normalize status transitions across:
  - provision
  - configure
  - rebuild
  - resize
  - delete
- Introduce explicit in-progress server statuses where needed.
- Ensure provider-side failure leaves truthful status and a clear recovery path.
- Ensure activity log titles match actual action semantics.
- Ensure UI and API do not imply synchronous completion for async operations.

### Deliverables

- Truthful deletion workflow
- Consistent server lifecycle state model
- Clear action precondition rules

### Acceptance criteria

- Deleting a server always means provider-side deprovision is attempted through a job.
- DB records are not silently removed while cloud resources still exist.
- Rebuild, resize, and delete actions are serialized safely.
- UI state remains coherent during and after failures.

### Tests

- Handler tests for delete endpoint behavior
- Executor tests for delete, rebuild, and resize success and failure paths
- Store tests for lifecycle status updates and duplicate-action blocking

### Exit check

- No destructive server action lies about what happened.

### What is now true

- Server deletion is orchestration-backed instead of directly deleting the database row.
- Delete, rebuild, and resize now use explicit in-progress server statuses and block duplicate destructive workflows.
- A successful delete leaves a durable `deleted` tombstone; failures leave an operator-visible `failed` state instead of implying success.
- UI and API responses describe delete/rebuild/resize as asynchronous jobs rather than synchronous completion.

---

## Milestone 3 - Simplify and Harden the Job Model

### Outcome

The orchestrator exposes only states and recovery behavior that the worker actually supports.

### Files and areas

- `internal/orchestrator/types.go`
- `internal/orchestrator/state_machine.go`
- `internal/orchestrator/store.go`
- `internal/worker/worker.go`
- `internal/worker/executor.go`
- `internal/database/migrations/00003_create_jobs.sql`
- `internal/dispatch/completer.go`

### Problems to solve

- The orchestrator surface is richer than real execution.
- The state machine includes statuses not truly exercised.
- `job_steps` and `job_checkpoints` exist but are not meaningfully used.
- Recovery logic is thinner than the modeled lifecycle suggests.

### Tasks

- Audit all current job kinds and actual transitions.
- Choose one of these paths and commit to it:
  - reduce the model to the states the platform really uses now
  - or fully implement checkpoints, retries, and intermediate states
- Recommended now: simplify first.
- Align:
  - durable statuses
  - emitted events
  - current step updates
  - worker recovery logic
- Make stuck-job recovery explicit and deterministic.
- Define timeout strategy per job kind.
- Define retry policy per job kind.
- Decide whether `job_steps` and `job_checkpoints` are:
  - implemented now
  - deferred and removed
- Ensure activity log and job timeline are derived from the same truth, not parallel assumptions.

### Deliverables

- Reduced and honest job state model
- Recovery strategy per job kind
- Timeout and retry policy table

### Acceptance criteria

- Every defined job status is either used or removed.
- Recovery logic covers every non-terminal status still in use.
- Timeline events and final statuses cannot contradict each other.
- Worker restart does not leave ambiguous job outcomes.

### Tests

- State machine tests updated to match actual supported transitions
- Worker recovery tests
- Store tests for transitions, event sequencing, and timeout handling
- Executor tests for fail, partial-fail, and resume behavior

### Exit check

- The job model stops pretending to support states the runtime cannot honor.

### What is now true

- Jobs now use only `queued`, `running`, `succeeded`, and `failed` as durable states.
- Worker interruption and timeout handling are explicit: in-flight jobs fail with a durable recovery reason instead of being silently requeued.
- Automatic retries are disabled for all current job kinds; recovery is manual and documented per kind.
- `job_steps` and `job_checkpoints` are removed from the live schema, and the job timeline is the only step-history source of truth.

---

## Milestone 4 - Make Profiles Real

### Outcome

Profiles stop being metadata and become actual, verified convergence targets.

### Files and areas

- `internal/server/profiles/registry.go`
- `ops/profiles/*/profile.yaml`
- `ops/schemas/profile.schema.json`
- `ops/ansible/playbooks/configure.yml`
- `ops/ansible/roles/common/tasks/main.yml`
- `ops/ansible/roles/nginx-stack/tasks/main.yml`
- `ops/ansible/roles/openlitespeed-stack/tasks/main.yml`
- `ops/ansible/roles/woocommerce-optimized/tasks/main.yml`
- `internal/worker/executor.go`

### Problems to solve

- `profile_path_resolved` is computed but not used.
- Role tasks are mostly stubs.
- A server can appear ready without meaningful stack convergence.
- Profile registry and ops artifacts are not strongly validated against each other.

### Tasks

- Define the minimal platform baseline all profiles must apply:
  - package updates policy
  - firewall defaults
  - agent and service prerequisites
  - system users and directories
  - common hardening baseline
- Define per-profile guarantees:
  - installed services
  - enabled services
  - config files and templates
  - health checks
- Wire `configure.yml` to actually load and execute profile artifact data.
- Choose one explicit profile-loading model for runtime-selected artifacts:
  - prefer dynamic includes
  - avoid mixing dynamic and static include/import behavior without need
- Validate `registry.go` keys against `ops/profiles/*`.
- Validate profile artifact schema before runtime with a named Draft 2020-12-compatible validator.
- Add post-config verification steps:
  - services running
  - ports and listeners as expected
  - config files present
  - optional command-based health checks
- Flush or otherwise sequence handlers before verification when verification depends on restarted services.
- Ensure `ready` is set only after verification passes.
- If full profile implementation is too broad now, reduce to one fully supported baseline profile first and downgrade others explicitly.
- Treat profile names like `nginx-stack`, `openlitespeed-stack`, and `woocommerce-optimized` as internal profile identifiers, not app-support claims.

### Deliverables

- Real configure pipeline using selected profile
- Verified baseline convergence contract
- Honest support matrix for profiles

### Acceptance criteria

- Selecting a profile changes actual server convergence behavior.
- A successful configure job proves more than `agent installed`.
- Profile failures leave actionable diagnostics.
- Unsupported or incomplete profiles are clearly marked and blocked or labeled experimental.

### Tests

- Schema validation tests for profiles
- Ansible syntax validation and advisory check-mode coverage
- Smoke tests for at least one supported profile
- Tests that registry and profile mismatches fail fast

### Exit check

- At least one profile is real enough to trust as a future site-hosting base.

### What is now true

- `configure.yml` now loads profile artifacts, applies the common baseline, dynamically includes the selected profile role, flushes handlers, and runs post-config verification before success.
- `nginx-stack` is the one fully supported baseline profile; `openlitespeed-stack` and `woocommerce-optimized` are explicitly unavailable instead of implying unsupported breadth.
- Registry keys, profile artifacts, and the profile schema are validated together, and profile schema checks use the named Draft 2020-12-compatible `santhosh-tekuri/jsonschema` validator.
- `ready` is now gated on verification passing for supported profile convergence rather than agent deployment alone.

---

## Milestone 5 - Harden Agent Commands and Node Status

### Outcome

The agent command channel is safe, validated, and accurately reflected in durable node status.

### Files and areas

- `internal/agent/executor.go`
- `internal/agent/commands/restart_service.go`
- `internal/agent/commands/list_services.go`
- `internal/dispatch/agent_runner.go`
- `internal/ws/handler.go`
- `internal/ws/command_wait.go`
- `internal/ws/monitor.go`
- `internal/server/store_servers.go`
- `web/app/composables/useAgentStatus.ts`

### Problems to solve

- Restart service path does not consistently validate service names.
- Some helper code validates but is not used in the live path.
- Node unhealthy and offline is persisted, but durable online updates are incomplete.
- Panic-based JSON helpers exist in long-lived runtime paths.

### Tasks

- Centralize command validation rules.
- Use the same validator in:
  - API initiation
  - dispatch
  - agent execution
- Add allowlist and format validation for service names.
- Normalize command result payloads and error reporting.
- Persist durable online status on connect and heartbeat.
- Define exact transitions among:
  - online
  - unhealthy
  - offline
- Replace all `mustMarshal` panic helpers in runtime paths with error-returning serialization.
- Ensure websocket and session failures cannot panic the process.
- Add command timeout semantics and stale-command handling.
- Decide whether `internal/dispatch/dispatcher.go` remains or is removed.

### Deliverables

- Safe command execution contract
- Durable node status model
- Panic-free runtime serialization path

### Acceptance criteria

- Invalid restart targets are rejected before execution.
- Healthy connected nodes become durably `online`.
- Connection loss degrades to `unhealthy` and then `offline` predictably.
- Serialization errors never crash long-lived services.

### Tests

- Agent executor tests
- WS handler and monitor tests
- Dispatch tests for timeout and result handling
- Node status persistence tests

### Exit check

- A connected node is durably represented as healthy, and bad commands fail safely.

### What is now true

- Agent commands now share one validation contract across API submission, dispatch, and on-node execution.
- Restart targets are constrained by both service-name format rules and an explicit allowlist for the supported baseline services.
- Websocket connect, heartbeat, disconnect degradation, and stale-node expiry now persist durable `online`, `unhealthy`, and `offline` status transitions.
- Runtime websocket and command serialization paths no longer rely on panic-based `mustMarshal` helpers.

---

## Milestone 6 - Observability, Cleanup, and Runtime Hygiene

### Outcome

The platform is easier to reason about, support, and extend.

### Files and areas

- `internal/ws/handler.go`
- `internal/dispatch/agent_runner.go`
- `internal/agent/agent.go`
- `internal/agentproto/types.go`
- `internal/dispatch/dispatcher.go`
- `web/README.md`
- `README.md`

### Tasks

- Remove or justify dead and unused packages and files.
- Remove misleading docs and stale references.
- Standardize structured logs for:
  - server actions
  - job lifecycle
  - agent registration
  - command dispatch and results
  - node health transitions
- Add correlation identifiers where useful:
  - job id
  - server id
  - command id
- Ensure activity log, job events, and runtime logs are cross-referenceable.
- Audit config and env var usage and document all required variables.

### Deliverables

- Cleaned runtime and logging surface
- Reduced dead code
- Accurate docs for current capabilities

### Acceptance criteria

- No obviously dead packages remain without explanation.
- Logs are sufficient to debug bootstrap, provision, and delete failures.
- Docs no longer overstate support or hide dev vs production differences.

### Exit check

- A new contributor can understand current support boundaries without reading half the codebase.

### What is now true

- Runtime logs now use consistent `server_id`, `job_id`, and `command_id` correlation fields across websocket sessions, command dispatch/results, agent bootstrap, and key job lifecycle paths.
- Dead runtime surface was reduced: the unused `internal/agentproto` package is gone, the stale `internal/dispatch/dispatcher.go` reference is now explicitly historical, and the direct `ServerStore.Delete` helper no longer implies a bypass around orchestration-backed deletion.
- README and `web/README.md` now document the real support boundaries, the live dispatch flow, and the full set of control-plane and agent-side configuration inputs that the current binaries read.

---

## Milestone 7 - Build Validation Gates Before Feature Work

### Outcome

The stabilized platform has automated gates that must pass before new higher-level features land.

### Files and areas

- `Makefile`
- `ops/tests/README.md`
- `internal/**/**/*_test.go`
- CI configuration if and where added later

### Tasks

- Define the test pyramid for the current platform:
  - unit tests
  - integration and store tests
  - end-to-end smoke path
- Add missing coverage in high-risk packages:
  - `internal/worker`
  - `internal/ws`
  - `internal/dispatch`
  - `internal/agent`
  - `internal/registration`
  - `internal/agentauth`
- Add migration and startup tests.
- Add smoke scenario scripts for:
  - provider setup
  - server provision
  - configure
  - agent register and connect
  - restart service
  - delete server
- Add Ansible validation targets:
  - syntax check
  - schema validation with a named validator
  - profile consistency validation
- Add one top-level verification command for contributors.
- Use `make check` as the top-level aggregate verification target.
- Separate fast local checks from heavier smoke tests.
- Treat `ansible-playbook --check` and `--diff` as advisory only, not proof of convergence.
- Run env-sensitive Go integration or smoke tests uncached when freshness matters.

### Deliverables

- Stable local validation commands
- Baseline smoke suite
- Minimum required checks for future feature branches

### Acceptance criteria

- There is a documented command sequence that validates the platform layer end-to-end.
- High-risk paths are no longer mostly untested.
- WordPress and site work can rely on automated regression protection.

### Suggested command targets

- `make check`
- `make test`
- `make lint`
- new profile validation target
- new smoke or integration target

### Exit check

- A contributor can verify platform health without manual archaeology.

---

## Recommended Execution Order

### Phase 1 - Contract and trust

- Milestone 0
- Milestone 1

### Phase 2 - Truthful operations

- Milestone 2
- Milestone 3

### Phase 3 - Real convergence

- Milestone 4
- Milestone 5

### Phase 4 - Confidence and cleanup

- Milestone 6
- Milestone 7

This order matters:

- agent bootstrap must be trustworthy before platform convergence can depend on agents
- lifecycle semantics must be truthful before more workflows are added
- profiles must be real before WordPress and site workflows can safely target them
- tests and gates must exist before feature velocity increases

---

## First Implementation Slice

If we want the best first execution slice, start here:

1. Write the platform contract from Milestone 0.
2. Fix agent bootstrap and trust from Milestone 1.
3. Replace fake delete semantics from Milestone 2.
4. Simplify the job model from Milestone 3.

That is the smallest slice that turns the platform from promising prototype into something operationally honest.

---

## Priority Backlog

### P0 - Must fix before anything else

- Complete the agent registration and bootstrap contract.
- Fix the production transport mismatch around mTLS and TLS.
- Replace direct DB-delete server behavior with real delete orchestration.
- Simplify or implement the job state model so it stops overstating capabilities.
- Persist durable `online` node status.
- Remove runtime panic paths from websocket and agent serialization.

### P1 - Must fix before the platform is considered stable

- Make profiles actually execute and verify convergence.
- Add retry, timeout, and recovery policy for all live job kinds.
- Add command validation consistency.
- Clean dead code and stale docs.
- Add high-risk package tests.

### P2 - Should finish before the WordPress and site milestone starts

- Improve logs and correlation ids.
- Add smoke environment for full provision, configure, and delete loop.
- Tighten profile schema and registry validation.
- Document contributor validation workflow.

---

## Risks and Mitigations

### Risk: production bootstrap expands scope too much

Mitigation:

- support only one production-safe bootstrap path now
- explicitly defer advanced cert rotation and distributed CA concerns

### Risk: profile work becomes a huge infrastructure project

Mitigation:

- support one profile fully first
- mark remaining profiles experimental or blocked until implemented

### Risk: job-state redesign becomes abstract architecture work

Mitigation:

- optimize for honesty, not sophistication
- remove unused states instead of building future-state machinery prematurely

### Risk: smoke tests become brittle

Mitigation:

- keep one critical happy-path smoke scenario first
- add failure scenarios only after the basic path is reliable

---

## Definition of Done

The platform layer is considered stabilized only when all of the following are true:

- Provision -> configure -> agent connect completes reliably.
- Delete means real deprovision, not just DB removal.
- Rebuild, resize, and delete semantics are truthful and guarded.
- Supported job states match actual worker behavior.
- At least one server profile is fully convergent and verified.
- Agent commands are validated and do not use panic-based helpers.
- Node status is durable and accurate.
- Docs reflect actual runtime behavior.
- High-risk platform paths have automated tests.
- A smoke path exists that proves the baseline platform works end-to-end.

---

## Exit Gate Before WordPress Features

Do not start WordPress or site features until this checklist is green:

- [ ] Agent bootstrap works in dev and production-defined modes.
- [ ] TLS and mTLS story is implemented or explicitly hard-gated.
- [ ] Server delete is orchestration-backed and provider-truthful.
- [ ] Job model is reduced or fully implemented consistently.
- [ ] One real profile converges and verifies successfully.
- [ ] Node online, unhealthy, and offline status is durable.
- [ ] Runtime panic helpers are removed from long-lived paths.
- [ ] Dead code and stale docs are cleaned up.
- [ ] Platform smoke checks pass consistently.

---

## Suggested Issue Breakdown

### Epic 1 - Trusted Agent Bootstrap

- wire registration token generation into configure flow
- fix agent config path handling during registration
- persist TLS-compatible key material
- fix registration consume and validate ordering
- implement or hard-gate production TLS startup
- add bootstrap tests and smoke test

### Epic 2 - Truthful Server Lifecycle

- replace direct delete handler behavior
- add delete job orchestration contract
- normalize lifecycle status transitions
- prevent conflicting destructive actions
- test provider failure cases

### Epic 3 - Honest Orchestration Core

- audit current job and status usage
- simplify or implement unused states
- decide fate of checkpoints and steps tables
- add timeout and recovery policies
- add worker and orchestrator tests

### Epic 4 - Real Profile Convergence

- wire profile artifact usage into configure playbook
- implement baseline common role
- fully support one profile end-to-end
- add post-config verification
- validate registry, profile, and schema consistency

### Epic 5 - Safe Agent Command Plane

- centralize service command validation
- remove panic-based JSON paths
- persist durable online state
- harden ws result handling and timeouts
- add command and monitor tests

### Epic 6 - Confidence Layer

- clean dead code
- fix stale docs
- improve structured logs
- add top-level validation commands
- add smoke and integration gates

---

## Final Note

If implementation capacity is limited, the recommended minimum stabilization slice is:

1. fix bootstrap and trust
2. fix truthful delete and lifecycle semantics
3. simplify the job model
4. fully implement one real profile
5. add smoke coverage

That is the smallest credible platform baseline for future WordPress and site work.
