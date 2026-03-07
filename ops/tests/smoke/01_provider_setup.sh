#!/bin/sh

set -eu

. "$(dirname "$0")/common.sh"
maybe_print_help "${1:-}"
announce_step "01 provider setup"

require_env PRESSLUFT_HETZNER_API_TOKEN

provider_name=${PRESSLUFT_PROVIDER_NAME:-hetzner-smoke}
response=$(api_request POST /api/providers "$(printf '{"type":"hetzner","name":"%s","api_token":"%s"}' "$provider_name" "$PRESSLUFT_HETZNER_API_TOKEN")")
provider_id=$(json_get "$response" id)

state_set provider_id "$provider_id"

printf 'provider_id=%s\n' "$provider_id"
