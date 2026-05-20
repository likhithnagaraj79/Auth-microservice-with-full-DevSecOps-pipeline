.PHONY: build run test lint gosec docker-build docker-up docker-down clean help

APP_NAME   := auth-service
IMAGE_NAME := ghcr.io/likhithnagaraj79/auth-service
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	CGO_ENABLED=0 go build -ldflags="-w -s -X main.version=$(VERSION)" -o bin/$(APP_NAME) ./cmd/server

run: ## Run locally (requires .env)
	go run ./cmd/server

test: ## Run unit tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint: ## Run golangci-lint
	golangci-lint run ./...

gosec: ## Run gosec SAST
	gosec -fmt sarif -out gosec-results.sarif ./... || true
	gosec ./...

docker-build: ## Build Docker image
	docker build -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .

docker-up: ## Start services with docker-compose
	docker-compose up -d

docker-down: ## Stop services
	docker-compose down

migrate-up: ## Run DB migrations (requires running DB)
	go run ./cmd/server --migrate-only || true

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out gosec-results.sarif

trivy: ## Scan Docker image for vulnerabilities
	trivy image --exit-code 1 --severity HIGH,CRITICAL $(IMAGE_NAME):latest
