#!/bin/sh

set -eu

. "$(dirname "$0")/common.sh"

server_id=$(state_get server_id)
job_id=$(state_get provision_job_id)
if [ -z "$server_id" ] || [ -z "$job_id" ]; then
  printf '%s\n' 'server_id or provision_job_id missing; run 02_server_provision.sh first' >&2
  exit 1
fi

wait_for_job_status "$job_id" succeeded 180 10 >/dev/null
wait_for_server_status "$server_id" ready 60 5 >/dev/null

history=$(api_request GET "/api/jobs/$job_id/events/history")
configure_seen=$(python3 - "$history" <<'PY'
import json
import sys

events = json.loads(sys.argv[1])
matched = any(event.get("step_key") == "configure" and event.get("event_type") == "step_complete" for event in events)
print("true" if matched else "false")
PY
)
if [ "$configure_seen" != "true" ]; then
  printf '%s\n' 'configure step completion not found in job history' >&2
  exit 1
fi

printf 'configure_job_id=%s\n' "$job_id"
