#!/bin/sh

set -eu

. "$(dirname "$0")/common.sh"
maybe_print_help "${1:-}"
announce_step "04 agent register/connect"

server_id=$(state_get server_id)
if [ -z "$server_id" ]; then
  printf '%s\n' 'server_id missing; run 02_server_provision.sh first' >&2
  exit 1
fi

response=$(wait_for_agent_online "$server_id" 120 5)
status=$(json_get "$response" status)
connected=$(json_get "$response" connected)

state_set agent_status "$status"

printf 'agent_connected=%s\nagent_status=%s\n' "$connected" "$status"
