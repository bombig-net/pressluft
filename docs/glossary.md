# Glossary

- `server status`: durable lifecycle status for the managed server record, such as `pending`, `provisioning`, `ready`, or `failed`.
- `node status`: durable health status for the agent connection, such as `online`, `unhealthy`, `offline`, or `unknown`; it is persisted on connect, heartbeat, disconnect degradation, and stale-node expiry.
- `job status`: durable orchestration status for one job: `queued`, `running`, `succeeded`, or `failed`.
- `activity event`: account-facing audit entry that summarizes something meaningful that happened in the platform.
- `profile convergence`: proof that a selected profile actually brought a machine to the expected baseline and verification passed.
- `registration token`: one-time bootstrap secret delivered during configure for the production-bootstrap path; it is consumed only after certificate issuance is persisted.
- `dev ws token`: development-only bearer token that lets an agent connect to the control plane over WebSockets without certificates.
- `certificate`: X.509 client identity issued by the Pressluft CA for the mTLS agent transport path.
