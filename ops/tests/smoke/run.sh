#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
. "$SCRIPT_DIR/common.sh"

maybe_print_help "${1:-}"
announce_step "smoke: provider setup"

"$SCRIPT_DIR/01_provider_setup.sh"
announce_step "smoke: server provision"
"$SCRIPT_DIR/02_server_provision.sh"
announce_step "smoke: configure verification"
"$SCRIPT_DIR/03_configure.sh"
announce_step "smoke: agent connectivity"
"$SCRIPT_DIR/04_agent_register_connect.sh"
announce_step "smoke: restart service"
"$SCRIPT_DIR/05_restart_service.sh"
announce_step "smoke: delete server"
"$SCRIPT_DIR/06_delete_server.sh"
