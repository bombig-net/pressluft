#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

"$SCRIPT_DIR/01_provider_setup.sh"
"$SCRIPT_DIR/02_server_provision.sh"
"$SCRIPT_DIR/03_configure.sh"
"$SCRIPT_DIR/04_agent_register_connect.sh"
"$SCRIPT_DIR/05_restart_service.sh"
"$SCRIPT_DIR/06_delete_server.sh"
