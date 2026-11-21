# Integration tests

From the repository root run `bash scripts/run-integration-tests.sh` on Linux or macOS. On Windows run `.\scripts\run-integration-tests.ps1`. Both commands require Docker Compose v2, curl and the Go version declared by the module. Leave loopback ports 443 and 8081 available. Use `scripts/run-config-ui-integration-tests.sh` or the corresponding PowerShell script to run only the configuration UI cases.

The runner builds the proxy and the gRPC backend from the checkout. Four nginx services supply HTTP responses. A fixture container copies the tracked configuration into a fresh volume and generates a one-day localhost certificate in a second volume. The proxy can update its configuration during the tests. The runner enables `INTEGRATION_TEST=true`, disables Go's test-result cache and removes the project's containers and volumes on exit. [Compose stack](docker-compose.test.yml) · [Configuration used by the proxy](testdata/config-docker.json)

The suite exercises HTTP routing, transparent and selective gRPC-Web routes and the configuration web interface. A plain `go test ./integration/...` skips the network cases unless the integration environment variable is set.
