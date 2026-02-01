.PHONY: help build test lint docker-build docker-push docker-run docker-stop clean

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the Go application"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linters"
	@echo "  docker-build  - Build Docker image locally"
	@echo "  docker-push   - Push Docker image to GHCR"
	@echo "  docker-run    - Start Docker Compose services"
	@echo "  docker-stop   - Stop Docker Compose services"
	@echo "  clean         - Clean build artifacts"

# Build the Go application
build:
	go build -o bin/vibe-cv ./cmd/main.go

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run linters
lint:
	golangci-lint run ./...

# Build Docker image locally
docker-build:
	docker build -f docker/Dockerfile -t ghcr.io/sammyoina/vibe-cv:local .

# Push Docker image to GHCR (requires authentication)
docker-push:
	docker push ghcr.io/sammyoina/vibe-cv:local

# Start Docker Compose services
docker-run:
	cd docker && docker-compose up -d

# Stop Docker Compose services
docker-stop:
	cd docker && docker-compose down

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out
	go clean
