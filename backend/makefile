# Энэ backend нь platform-core-ын НИМГЭН хэрэглэгч болсон: бүх суурь код
# github.com/gerege-systems/platform-core модульд байрлана. Тиймээс swag
# (OpenAPI) болон mockery target-ууд ЭНД БАЙХГҮЙ — тэдгээр нь core repo-д
# амьдарна. Энд зөвхөн энэ аппын build/test/lint үлдэнэ.

.PHONY: help build tidy test test-cover lint fmt core-update \
        pre-push ci-lint ci-test ci-build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the API binary (local sanity build; production = Dockerfile)
	go build -o bin/api ./cmd/api

tidy: ## go mod tidy
	go mod tidy

test: ## Run this app's unit tests (core-ийн тестүүд core repo-д ажиллана)
	go test ./...

test-cover: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format all Go files
	gofmt -w .

core-update: ## Bump platform-core to its latest tagged release
	go get github.com/gerege-systems/platform-core@latest
	go mod tidy

pre-push: ci-lint ci-test ci-build ## Mirror CI checks locally before pushing
	@echo "All CI checks passed."

ci-lint: ## Run golangci-lint the way CI does
	@command -v golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	golangci-lint run ./...

ci-test: ## Run unit tests with race + coverage matching CI
	go test -race -coverprofile=coverage.out ./...

ci-build: ## Build API binary matching CI build job
	go build -o bin/api ./cmd/api
