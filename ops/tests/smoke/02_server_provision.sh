#!/bin/sh

set -eu

. "$(dirname "$0")/common.sh"
maybe_print_help "${1:-}"
announce_step "02 server provision"

provider_id=$(state_get provider_id)
if [ -z "$provider_id" ]; then
  printf '%s\n' 'provider_id missing; run 01_provider_setup.sh first' >&2
  exit 1
fi

server_name=${PRESSLUFT_SERVER_NAME:-pressluft-smoke}
server_location=${PRESSLUFT_SERVER_LOCATION:-nbg1}
server_type=${PRESSLUFT_SERVER_TYPE:-cx22}
profile_key=${PRESSLUFT_PROFILE_KEY:-nginx-stack}

response=$(api_request POST /api/servers "$(printf '{"provider_id":%s,"name":"%s","location":"%s","server_type":"%s","profile_key":"%s"}' "$provider_id" "$server_name" "$server_location" "$server_type" "$profile_key")")
server_id=$(json_get "$response" server_id)
job_id=$(json_get "$response" job_id)

state_set server_id "$server_id"
state_set provision_job_id "$job_id"

wait_for_server_status "$server_id" configuring 60 5 >/dev/null || wait_for_server_status "$server_id" ready 120 5 >/dev/null

printf 'server_id=%s\nprovision_job_id=%s\n' "$server_id" "$job_id"
