# Ops Validation

Milestone 7 makes the contributor verification flow explicit and split by cost.

## Test pyramid for the current platform

- Unit tests: package-level Go tests for validation, executor behavior, websocket/session handling, agent registration, token stores, and startup helpers.
- Integration and store tests: database-backed handler, store, migration, and orchestration tests that exercise real SQLite schemas and cross-package contracts.
- Smoke path: disposable-environment scripts that drive the supported happy path end to end for provider setup, provision, configure verification, agent connectivity, restart service, and delete.

## Fast local checks

Run these before pushing most platform changes:

```bash
make check
```

`make check` aggregates the fast gates:

- `fmt-check`
- `lint`
- `test` (`test-unit` plus uncached `test-integration`)
- `ansible-validate` (`ansible-syntax`, schema validation with the named `santhosh-tekuri/jsonschema` Draft 2020-12 validator, and profile consistency validation)
- `build`

Useful targeted commands:

```bash
make test-unit
make test-integration
make validate-profile-schema
make validate-profile-consistency
make ansible-syntax
make ansible-check
```

`make ansible-check` runs `ansible-playbook --check --diff` only as advisory coverage. It is not proof of convergence and does not replace a real smoke run.

## Smoke workflow

The heavier disposable-environment path lives under `ops/tests/smoke/`.

Run the full sequence with:

```bash
make smoke
```

Or run steps individually:

```bash
./ops/tests/smoke/01_provider_setup.sh
./ops/tests/smoke/02_server_provision.sh
./ops/tests/smoke/03_configure.sh
./ops/tests/smoke/04_agent_register_connect.sh
./ops/tests/smoke/05_restart_service.sh
./ops/tests/smoke/06_delete_server.sh
```

Required environment for the smoke scripts:

- `PRESSLUFT_API_BASE` - control-plane base URL, for example `http://127.0.0.1:8080`
- `PRESSLUFT_HETZNER_API_TOKEN` - disposable Hetzner token for the smoke environment

Common optional overrides:

- `PRESSLUFT_PROVIDER_NAME` - defaults to `hetzner-smoke`
- `PRESSLUFT_SERVER_NAME` - defaults to `pressluft-smoke`
- `PRESSLUFT_SERVER_LOCATION` - defaults to `nbg1`
- `PRESSLUFT_SERVER_TYPE` - defaults to `cx22`
- `PRESSLUFT_PROFILE_KEY` - defaults to `nginx-stack`
- `PRESSLUFT_RESTART_SERVICE` - defaults to `nginx`
- `PRESSLUFT_SMOKE_STATE_DIR` - defaults to `ops/tests/.smoke-state`

Smoke scripts intentionally run env-sensitive validation fresh each time. They poll live API state instead of relying on cached Go test results.
