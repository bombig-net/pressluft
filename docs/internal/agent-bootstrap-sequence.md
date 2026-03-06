# Agent Bootstrap Sequence

This is the Milestone 1 bootstrap contract for the supported production path.

## Production bootstrap

```text
Control plane            Worker/configure             Node agent              PKI / DB
     |                         |                          |                      |
     | provision server        |                          |                      |
     |------------------------>|                          |                      |
     |                         | create registration token|                      |
     |                         |------------------------->| registration_tokens   |
     |                         | render /etc/pressluft/agent.yaml                |
     |                         | with HTTPS control_plane and registration token  |
     |                         |------------------------->|                      |
     |                         | start agent service      |                      |
     |                         |------------------------->|                      |
     |                         |                          | build keypair + CSR   |
     |                         |                          | POST /api/nodes/:id/register over HTTPS
     |                         |                          |--------------------->|
     |                         |                          | validate token + CSR |
     |                         |                          | sign cert            |
     |                         |                          | save cert, then consume token in one tx
     |                         |                          |<---------------------|
     |                         |                          | persist cert/key/CA  |
     |                         |                          | clear registration token from real config path
     |                         |                          | connect /ws/agent via wss + mTLS
     |<--------------------------------------------------|                      |
```

## Dev bootstrap

- `dev` builds skip certificate registration.
- Configure renders `dev_ws_token` into the agent config.
- Agent connects over `ws` with the dev token.

## Reissue and reconnect policy

- Node certificates are issued for 90 days.
- A still-valid certificate blocks duplicate registration unless it is inside the 14-day reissue window and a fresh registration token is provided.
- Reconnect retries use exponential backoff with jitter, capped at 30 seconds between attempts, and continue until shutdown.
- If a certificate is expired or invalid and no fresh registration token is present, the agent fails closed instead of silently downgrading transport.
