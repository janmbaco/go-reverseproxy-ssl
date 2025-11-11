# Go Reverse Proxy SSL - Build and Test

# Build the application
build:
	go build -o bin/reverseproxy ./cmd/reverseproxy

# Run unit tests
test-unit:
	go test ./internal/...

# Run integration tests with Docker (auto-detect OS)
test-integration-docker:
	@echo "Running integration tests with Docker..."
	@if [ "$$(uname)" = "Linux" ] || [ "$$(uname)" = "Darwin" ]; then \
		./scripts/run-integration-tests.sh; \
	elif [ "$$(uname)" = "Windows" ] || [ "$$(uname -o)" = "Msys" ]; then \
		.\scripts\run-integration-tests.ps1; \
	else \
		echo "Unsupported OS. Please run manually."; \
		exit 1; \
	fi

# Test HTTP virtual hosts
test-http:
	@echo "Running HTTP integration tests..."
	INTEGRATION_TEST=true go test -v ./integration -run TestHttpThroughProxy

# Test gRPC virtual hosts
test-grpc:
	@echo "Running gRPC integration tests..."
	INTEGRATION_TEST=true go test -v ./integration -run TestGrpcWebThroughProxy

# Run all tests
test-all: test-unit test-integration

# Build and start services for testing
start-services:
	docker-compose -f build/package/docker-compose.dev.yml up -d

# Stop services
stop-services:
	docker-compose -f build/package/docker-compose.dev.yml down

# Clean up
clean:
	docker-compose -f build/package/docker-compose.dev.yml down -v

# Development helpers
fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

# Docker build
docker-build:
	docker build -f build/package/Dockerfile -t go-reverseproxy-ssl .

# Run locally
run:
	go run ./cmd/reverseproxy --config configs/config.json