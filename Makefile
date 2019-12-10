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
	rm -fr ${DIST_DIR}
	mkdir ${DIST_DIR}

.PHONY: lint
lint: ## runs go vet
	go vet ./core/...

.PHONY: test
test: clean lint build-relay build-sdk ## runs linters and all tests with coverage
	./dist/relay test
	./dist/sdk
	go test ./core/...

.PHONY: dev-optimizer
dev-optimizer: ## runs a local optimizer
	go run cmd/optimizer/optimizer.go

.PHONY: dev-relay-backend
dev-relay-backend: ## runs a local relay_backend
	go run cmd/relay_backend/relay_backend.go

.PHONY: dev-server-backend
dev-server-backend: ## runs a local server_backend
	go run cmd/server_backend/server_backend.go

.PHONY: build-optimizer
build-optimizer: ## builds the optimizer binary
	go build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/optimizer ./cmd/optimizer/optimizer.go

.PHONY: build-relay
build-relay:
	g++ -o ./dist/relay ./cmd/relay/*.cpp -lsodium -lcurl -lpthread -lm

.PHONY: build-sdk
build-sdk:
	g++ -o ./dist/sdk ./sdk/*.cpp -lsodium -lcurl -lpthread -lm

PHONY: build-tools
build-tools:
	go build -o ./dist/analyze ./cmd/tools/analyze/analyze.go
	go build -o ./dist/cost ./cmd/tools/cost/cost.go
	go build -o ./dist/debug ./cmd/tools/debug/debug.go
	go build -o ./dist/optimize ./cmd/tools/optimize/optimize.go
	go build -o ./dist/route ./cmd/tools/route/route.go
	go build -o ./dist/backend ./cmd/tools/functional/backend/*.go
	go build -o ./dist/tests ./cmd/tools/functional/tests/functional_tests.go
	cd cmd/tools/keygen && ./build.sh

.PHONY: build-relay-backend
build-relay-backend: ## builds the relay_backend binary
	go build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/relay_backend ./cmd/relay_backend/relay_backend.go

.PHONY: build-server-backend
build-server-backend: ## builds the server_backend binary
	go build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/server_backend ./cmd/server_backend/server_backend.go

.PHONY: build-all
build-all: build-optimizer build-relay-backend build-server-backend ## builds everything
