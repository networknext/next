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

.PHONY: dev-relay-ingress
dev-relay-ingress: ## runs a local relay_ingress
	go run cmd/relay_ingress/relay_ingress.go

.PHONY: dev-server-ingress
dev-server-ingress: ## runs a local server_ingress
	go run cmd/server_ingress/server.go

.PHONY: build-relay-ingress
build-relay-ingress: ## builds the relay_ingress binary.
	go build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/relay_ingress ./cmd/relay_ingress/relay_ingress.go

.PHONY: build-server-ingress
build-server-ingress: ## builds the server_ingress binary.
	go build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/server_ingress ./cmd/server_ingress/server_ingress.go

.PHONY: build-all
build-all: build-relay-ingress build-server-ingress ## builds everything