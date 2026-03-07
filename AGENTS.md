# Pressluft Agent Notes

- Keep repository-level Markdown limited to `README.md` and `AGENTS.md`.
- Treat [platform-contract.generated.ts](/home/deniz/projects/pressluft/web/app/lib/platform-contract.generated.ts) as generated output. Refresh it with `make generate-contract` after changing the Go runtime contract.
- When changing runtime semantics, update the Go contract source, generated TypeScript contract, and contract tests together.
- Prefer executable contracts: schemas, tests, types, assertions, and command help over new prose files.
