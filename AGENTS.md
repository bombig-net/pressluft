Read `README.md` first.

The project uses a unified CLI (`pressluft`) for all development and build tasks.
Bootstrap it with `make`, then use `pressluft <command>` for everything.
Run `pressluft help` to see all available commands.

Do not hand-edit generated files:
- `web/app/lib/api-contract.ts`
- `web/app/lib/platform-contract.generated.ts`

Regenerate them with `pressluft generate`.

Ignore generated/local directories during search unless the task explicitly needs them:
- `web/.output`
- `web/.nuxt`
- `.venv`

The three binaries are:
- `cmd/pressluft/` — the CLI (dev tools, build, generate, diagnostics)
- `cmd/pressluft-server/` — the control-plane server (runtime)
- `cmd/pressluft-agent/` — the server agent (runtime)
