# Web Contract Notes

The frontend should treat the platform contract in `README.md#platform-contract` as the source of truth for status labels, support messaging, and experimental badges.

## Status semantics the UI should preserve

- `server status` and `node status` are different concepts and must not be merged into one badge.
- `ready` means the latest supported lifecycle workflow completed successfully. For `nginx-stack`, that now includes post-config verification; unavailable profiles should never imply readiness.
- `configuring`, `rebuilding`, `resizing`, and `deleting` are explicit in-progress server states and should be shown honestly.
- `deleted` is a tombstone record, not a manageable server.
- `online`, `unhealthy`, `offline`, and `unknown` describe only the agent channel.
- `connected` is live session state; durable node `status` remains the source of truth for badges and historical refreshes.
- Treat `unhealthy` as degraded-but-recently-seen, and `offline` as stale beyond the backend offline threshold.
- `delete`, `rebuild`, and `resize` are asynchronous destructive jobs and should be shown as such.

## Support messaging

- Hetzner is the only truthful provider integration today.
- `nginx-stack` is the only supported profile. `openlitespeed-stack` and `woocommerce-optimized` should be shown as unavailable internal identifiers, not as app-support claims.
- Production bootstrap copy should describe the narrow supported path truthfully: HTTPS registration, then `wss` + mTLS reconnect. Do not generalize that into broad production-readiness claims.

## Cross-referencing expectations

- Job detail views, command result surfaces, and activity-linked UI should preserve `job_id`, `server_id`, and `command_id` where the backend exposes them.
- Treat runtime logs, activity entries, and job event timelines as related troubleshooting surfaces rather than separate stories.

## Local web development

```bash
pnpm install
pnpm dev
```

For the full local stack, run `make dev` from the repository root.
