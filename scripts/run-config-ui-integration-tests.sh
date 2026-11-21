#!/usr/bin/env bash
set -euo pipefail
exec bash "$(dirname -- "${BASH_SOURCE[0]}")/run-integration-tests.sh" "${1:-2}" '^TestConfigUI_'
