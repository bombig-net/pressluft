Read `README.md` first.

The project uses a unified CLI (`pressluft`) for all development and build tasks.
Bootstrap it with `make`, then use `pressluft <command>` for everything.
Run `pressluft help` to see all available commands.

Do not hand-edit generated files:
- `web/app/lib/api-contract.ts`
- `web/app/lib/platform-contract.generated.ts`

These are regenerated automatically by `pressluft dev` and `pressluft build`.

Ignore generated/local directories during search unless the task explicitly needs them:
- `web/.output`
- `web/.nuxt`
- `.venv`

The three binaries are:
- `cmd/pressluft/` — the CLI (dev tools, build, doctor)
- `cmd/pressluft-server/` — the control-plane server (runtime)
- `cmd/pressluft-agent/` — the server agent (runtime)

Local dev state lives in `.pressluft/` at the repo root (SQLite DB, keys).
To reset: `rm -rf .pressluft`

To check system health: `pressluft doctor`
