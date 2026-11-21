#!/usr/bin/env bash
# Requires Docker Compose v2, curl and Go 1.26+. The optional argument is a timeout in minutes.
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
minutes=${1:-10}
[[ $minutes =~ ^[1-9][0-9]*$ ]] || { echo "Expected a positive timeout in minutes" >&2; exit 2; }
compose=(docker compose --project-name "reverseproxy-tests-$$" -f integration/docker-compose.test.yml)
docker compose version
cleanup() {
  local status=$?
  if (( status != 0 )); then "${compose[@]}" logs --no-color >&2 || true; fi
  "${compose[@]}" down --volumes --remove-orphans || true
}
trap cleanup EXIT
"${compose[@]}" up --detach --build
for url in https://localhost/ http://localhost:8081/api/config; do
  ready=false
  for ((i=0; i<40; i++)); do
    if curl --noproxy '*' --insecure --fail --silent --max-time 2 "$url" >/dev/null; then ready=true; break; fi
    sleep 1
  done
  if [[ $ready != true ]]; then echo "Service did not become ready at $url" >&2; exit 1; fi
done
INTEGRATION_TEST=true go test -run "${2:-.}" -count=1 -timeout="${minutes}m" -v ./integration/...
