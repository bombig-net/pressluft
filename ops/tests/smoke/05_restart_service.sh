#!/bin/sh

set -eu

. "$(dirname "$0")/common.sh"

server_id=$(state_get server_id)
if [ -z "$server_id" ]; then
  printf '%s\n' 'server_id missing; run 02_server_provision.sh first' >&2
  exit 1
fi

service_name=${PRESSLUFT_RESTART_SERVICE:-nginx}
response=$(api_request POST /api/jobs "$(printf '{"kind":"restart_service","server_id":%s,"payload":{"service_name":"%s"}}' "$server_id" "$service_name")")
job_id=$(json_get "$response" id)

state_set restart_job_id "$job_id"
wait_for_job_status "$job_id" succeeded 60 5 >/dev/null

printf 'restart_job_id=%s\n' "$job_id"
