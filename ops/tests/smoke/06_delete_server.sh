#!/bin/sh

set -eu

. "$(dirname "$0")/common.sh"

server_id=$(state_get server_id)
if [ -z "$server_id" ]; then
  printf '%s\n' 'server_id missing; run 02_server_provision.sh first' >&2
  exit 1
fi

response=$(api_request DELETE "/api/servers/$server_id")
job_id=$(json_get "$response" job_id)

state_set delete_job_id "$job_id"
wait_for_job_status "$job_id" succeeded 120 10 >/dev/null
wait_for_server_status "$server_id" deleted 30 5 >/dev/null

printf 'delete_job_id=%s\n' "$job_id"
