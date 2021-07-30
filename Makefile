.DEFAULT_GOAL = build

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

setup: ## Get all dependencies
# Only install if missing
ifeq (,$(wildcard bin/golangci-lint))
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
endif

	go mod tidy
.PHONY: setup

build: ## Build tb
	go build
.PHONY: build

clean: ## Clean all build artifacts
	rm -rf artifacts
	rm -rf coverage
	rm -rf dist
	rm -f tb
.PHONY: clean

lint: ## Run the linter
	./bin/golangci-lint run ./...
.PHONY: lint

go-uninstall: ## Remove version of tb installed with go install
	rm $(shell go env GOPATH)/bin/tb
.PHONY: go-uninstall

test: ## Run tests and collect coverage data
	mkdir -p coverage
	go test -coverprofile=coverage/coverage.txt ./...
	go tool cover -html=coverage/coverage.txt -o coverage/coverage.html
.PHONY: test

test-ci: ## Run tests and print coverage data to stdout
	mkdir -p coverage
	go test -coverprofile=coverage/coverage.txt ./...
	go tool cover -func=coverage/coverage.txt
.PHONY: test-ci
