# Playbook Contract Notes

- `configure.yml` is the profile convergence entrypoint. It now loads the selected profile artifact, applies the common baseline, dynamically includes the selected role, flushes handlers, and verifies services, files, listeners, and health checks before the worker can mark a server `ready`.
- The runtime profile loading model is intentionally one thing: `dynamic-include-role`. Profile artifacts must declare `artifacts.loading_model: dynamic-include-role`, and `configure.yml` asserts that contract before execution.
- In `production-bootstrap` mode, `configure.yml` must receive an HTTPS `control_plane_url` and a fresh `agent_registration_token`. In `dev` mode it must receive `dev_ws_token` instead.
- `nginx-stack` is the only currently supported profile. `openlitespeed-stack` and `woocommerce-optimized` are intentionally unavailable and should fail fast if selected.
- `delete_server.yml`, `rebuild_server.yml`, and `resize_server.yml` are destructive workflows. They are wired into later lifecycle milestones and remain experimental until server statuses become fully truthful.
- The platform contract lives in `README.md#platform-contract`, with deeper lifecycle semantics in `docs/internal/lifecycle-state-semantics.md`.
