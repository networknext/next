TIMESTAMP ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA ?= $(shell git rev-parse --short HEAD)
TAG ?= $(shell git describe --tags 2> /dev/null)

CURRENT_DIR = $(shell pwd -P)
DIST_DIR = "./dist"

.PHONY: help
help: ## this list
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## cleans the dist directory of all builds
	@rm -fr ${DIST_DIR}

.PHONY: lint
lint: ## runs o vet
	go vet ./...

.PHONY: test
test: lint ## runs linters and all tests with coverage
	go test -v ./...

.PHONY: dev-router
dev-router: ## runs a local router-ingress
	go run cmd/router-ingress/router.go

.PHONY: dev-server
dev-server: ## runs a local server-ingress
	go run cmd/server-ingress/server.go

.PHONY: build-router
build-router: ## builds the router-ingress binary.
	go build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/router ./cmd/router-ingress/router.go

.PHONY: build-server
build-server: ## builds the server-ingress binary.
	go build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/server ./cmd/server-ingress/server.go

.PHONY: build-all
build-all: build-router build-server ## builds everything