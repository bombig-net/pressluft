# Pressluft

Self-hosted WordPress hosting control plane for agencies.

![Screenshot](screenshot.png)

## Current support

- Hetzner Cloud is the only real provider integration.
- `nginx-stack` is the only supported server profile.
- The control plane can provision/configure/delete servers, run Ansible playbooks, queue jobs, and communicate with agents over WebSockets.
- Production bootstrap is the HTTPS registration plus `wss` + mTLS path. Dev mode uses the dev WebSocket token path.

## Quickstart

Use a Unix-like environment. Local development is tested on Ubuntu 24.04 in WSL2.

Required tools:

- Go 1.24+
- Node.js 20+
- `pnpm`
- `cloudflared`
- Python 3 with `venv`

Install the Ansible dependency set from the repository root:

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install ansible
ansible-galaxy collection install -r ops/ansible/requirements.yml
```

Common entrypoints:

```bash
make help
make dev
make check
make build
make agent
```

## Machine-readable contracts

Runtime and environment contracts are generated from code:

```bash
make contract-json
make generate-contract
```

- `make contract-json` prints the runtime contract and environment variable catalog as JSON.
- `make generate-contract` refreshes [platform-contract.generated.ts](/home/deniz/projects/pressluft/web/app/lib/platform-contract.generated.ts) from the Go contract package.

## Development notes

- `make dev` starts the Go backend, the Nuxt app, and a Cloudflare quick tunnel when `PRESSLUFT_CONTROL_PLANE_URL` is unset.
- Quick tunnels are intentionally ephemeral; use a stable public URL if you need durable agent reconnect behavior across control-plane restarts.
- `make smoke` runs the disposable-environment smoke flow. Each smoke script supports `--help`.
