#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
STATE_DIR=${PRESSLUFT_SMOKE_STATE_DIR:-$REPO_ROOT/ops/tests/.smoke-state}
STATE_FILE=$STATE_DIR/state.json
API_BASE=${PRESSLUFT_API_BASE:-}

mkdir -p "$STATE_DIR"
if [ ! -f "$STATE_FILE" ]; then
  printf '{}\n' >"$STATE_FILE"
fi

require_env() {
  var_name=$1
  eval "var_value=\${$var_name:-}"
  if [ -z "$var_value" ]; then
    printf '%s\n' "$var_name is required" >&2
    print_smoke_env_help >&2
    exit 1
  fi
}

print_smoke_env_help() {
  printf '%s\n' 'Smoke environment contract:'
  printf '%s\n' '  PRESSLUFT_API_BASE            Required control-plane base URL'
  printf '%s\n' '  PRESSLUFT_HETZNER_API_TOKEN   Required disposable Hetzner token'
  printf '%s\n' '  PRESSLUFT_PROVIDER_NAME       Optional, defaults to hetzner-smoke'
  printf '%s\n' '  PRESSLUFT_SERVER_NAME         Optional, defaults to pressluft-smoke'
  printf '%s\n' '  PRESSLUFT_SERVER_LOCATION     Optional, defaults to nbg1'
  printf '%s\n' '  PRESSLUFT_SERVER_TYPE         Optional, defaults to cx22'
  printf '%s\n' '  PRESSLUFT_PROFILE_KEY         Optional, defaults to nginx-stack'
  printf '%s\n' '  PRESSLUFT_RESTART_SERVICE     Optional, defaults to nginx'
  printf '%s\n' '  PRESSLUFT_SMOKE_STATE_DIR     Optional state directory'
}

maybe_print_help() {
  case "${1:-}" in
    -h|--help|help)
      print_smoke_env_help
      exit 0
      ;;
  esac
}

announce_step() {
  printf '==> %s\n' "$1"
}

require_api_base() {
  require_env PRESSLUFT_API_BASE
}

state_set() {
  python3 - "$STATE_FILE" "$1" "$2" <<'PY'
import json
import pathlib
import sys

state_path = pathlib.Path(sys.argv[1])
key = sys.argv[2]
value = sys.argv[3]
state = json.loads(state_path.read_text())
state[key] = value
state_path.write_text(json.dumps(state, indent=2, sort_keys=True) + "\n")
PY
}

state_get() {
  python3 - "$STATE_FILE" "$1" <<'PY'
import json
import pathlib
import sys

state_path = pathlib.Path(sys.argv[1])
key = sys.argv[2]
state = json.loads(state_path.read_text())
value = state.get(key, "")
if isinstance(value, bool):
    print("true" if value else "false")
else:
    print(value)
PY
}

json_get() {
  python3 - "$1" "$2" <<'PY'
import json
import sys

payload = json.loads(sys.argv[1])
value = payload
for part in sys.argv[2].split('.'):
    if not part:
        continue
    if isinstance(value, list):
        value = value[int(part)]
    else:
        value = value.get(part, "")
if isinstance(value, bool):
    print("true" if value else "false")
elif value is None:
    print("")
else:
    print(value)
PY
}

api_request() {
  method=$1
  path=$2
  body=${3:-}
  require_api_base
  url=${API_BASE%/}$path
  if [ -n "$body" ]; then
    curl -fsS -X "$method" "$url" -H 'Content-Type: application/json' -d "$body"
  else
    curl -fsS -X "$method" "$url"
  fi
}

wait_for_job_status() {
  job_id=$1
  expected=$2
  attempts=${3:-120}
  sleep_seconds=${4:-5}
  i=0
  while [ "$i" -lt "$attempts" ]; do
    response=$(api_request GET "/api/jobs/$job_id")
    status=$(json_get "$response" status)
    if [ "$status" = "$expected" ]; then
      printf '%s\n' "$response"
      return 0
    fi
    if [ "$status" = "failed" ]; then
      printf '%s\n' "$response" >&2
      return 1
    fi
    i=$((i + 1))
    sleep "$sleep_seconds"
  done
  printf '%s\n' "timed out waiting for job $job_id to reach $expected" >&2
  return 1
}

wait_for_server_status() {
  server_id=$1
  expected=$2
  attempts=${3:-120}
  sleep_seconds=${4:-5}
  i=0
  while [ "$i" -lt "$attempts" ]; do
    response=$(api_request GET "/api/servers/$server_id")
    status=$(json_get "$response" status)
    if [ "$status" = "$expected" ]; then
      printf '%s\n' "$response"
      return 0
    fi
    if [ "$status" = "failed" ]; then
      printf '%s\n' "$response" >&2
      return 1
    fi
    i=$((i + 1))
    sleep "$sleep_seconds"
  done
  printf '%s\n' "timed out waiting for server $server_id to reach $expected" >&2
  return 1
}

wait_for_agent_online() {
  server_id=$1
  attempts=${2:-120}
  sleep_seconds=${3:-5}
  i=0
  while [ "$i" -lt "$attempts" ]; do
    response=$(api_request GET "/api/servers/$server_id/agent-status")
    connected=$(json_get "$response" connected)
    status=$(json_get "$response" status)
    if [ "$connected" = "true" ] && [ "$status" = "online" ]; then
      printf '%s\n' "$response"
      return 0
    fi
    i=$((i + 1))
    sleep "$sleep_seconds"
  done
  printf '%s\n' "timed out waiting for agent on server $server_id to become online" >&2
  return 1
}
