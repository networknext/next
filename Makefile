CXX = g++
GO = go

OS := $(shell uname -s | tr A-Z a-z)
ifeq ($(OS),darwin)
	LDFLAGS = -lsodium -lcurl -lpthread -lm -framework CoreFoundation -framework SystemConfiguration
else
	LDFLAGS = -lsodium -lcurl -lpthread -lm
endif

SDKNAME = libnext

TIMESTAMP ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA ?= $(shell git rev-parse --short HEAD)
TAG ?= $(shell git describe --tags 2> /dev/null)

CURRENT_DIR = $(shell pwd -P)
DIST_DIR = "./dist"

#####################
##    RELAY ENV    ##
#####################

export RELAY_ID = local
export RELAY_ADDRESS = 127.0.0.1
export RELAY_PUBLIC_KEY = BrBNnqb1fAs8ai2dvzQytmYAoDsrW10AkUoy7vI2wpw=
export RELAY_PRIVATE_KEY = sQXQuzR9HixMjhL+mJ2FCFg76cBcrS+MTN2H+qTxS+wGsE2epvV8CzxqLZ2/NDK2ZgCgOytbXQCRSjLu8jbCnA==
export RELAY_BACKEND_HOSTNAME = http://localhost:30000
export RELAY_ROUTER_PUBLIC_KEY = SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=
export RELAY_DEBUG = 0

.PHONY: help
help: ## this list
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## cleans the dist directory of all builds
	@rm -fr $(DIST_DIR)
	@mkdir $(DIST_DIR)

.PHONY: lint
lint: ## runs go vet
	@$(GO) vet ./core/...

#####################
## TESTS AND TOOLS ##
#####################

.PHONY: test
test: clean lint test-unit test-soak test-func ## runs linters and all tests with coverage

.PHONY: test-unit
test-unit: clean build-sdk-test build-relay ## runts unit tests for sdk, relay, and core
	@$(DIST_DIR)/$(SDKNAME)_test
	@$(DIST_DIR)/relay test
	@$(GO) test -race -v ./core/...
	@printf "\n"

.PHONY: test-soak
test-soak: clean build-sdk-test build-soak-test ## runs soak test
	@$(DIST_DIR)/$(SDKNAME)_soak_test
	@printf "\n"

.PHONY: test-func
test-func: clean build-sdk build-relay build-functional-server build-functional-client ## runs functional tests
	@printf "Building functional backend... "
	@go build -o ./dist/func_backend ./cmd/tools/functional/backend/*.go
	@printf "done\n"

	@printf "\nRunning functional tests...\n\n"
	@$(GO) run ./cmd/tools/functional/tests/func_tests.go
	@printf "done\n"

.PHONY: build-sdk-test 
build-sdk-test: build-sdk ## builds the sdk test binary
	@printf "Building sdk test... "
	@$(CXX) -Isdk -o $(DIST_DIR)/$(SDKNAME)_test ./sdk/next_test.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-soak-test
build-soak-test: build-sdk ## builds the sdk test binary
	@printf "Building soak test... "
	@$(CXX) -Isdk -o $(DIST_DIR)/$(SDKNAME)_soak_test ./sdk/next_soak.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

PHONY: build-tools
build-tools: ## builds all the tools
	@./cmd/tools/build.sh

#####################
## MAIN COMPONENTS ##
#####################

.PHONY: dev-relay
dev-relay: build-relay
	@./dist/relay

.PHONY: dev-optimizer
dev-optimizer: ## runs a local optimizer
	$(GO) run cmd/optimizer/optimizer.go

.PHONY: dev-relay-backend
dev-relay-backend: ## runs a local relay_backend
	$(GO) run cmd/relay_backend/relay_backend.go

.PHONY: dev-server-backend
dev-server-backend: ## runs a local server_backend
	$(GO) run cmd/server_backend/server_backend.go

.PHONY: dev-backend
dev-backend: ## runs a local mock backend that encompasses the relay backend and server backend
	$(GO) run cmd/tools/functional/backend/*.go

.PHONY: dev-server
dev-server: build-functional-server  ## runs a local mock backend that encompasses the relay backend and server backend
	@./dist/functional_server

.PHONY: dev-client
dev-client: build-functional-client  ## runs a local mock backend that encompasses the relay backend and client backend
	@./dist/functional_client

.PHONY: build-optimizer
build-optimizer: ## builds the optimizer binary
	@printf "Building optimizer... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/optimizer ./cmd/optimizer/optimizer.go
	@printf "done\n"

.PHONY: build-relay
build-relay: ## builds the relay
	@printf "Building relay... "
	@$(CXX) -o $(DIST_DIR)/relay ./cmd/relay/*.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-sdk
build-sdk: clean ## builds the sdk into a shared object for linking
	@printf "Building sdk... "
	@$(CXX) -fPIC -shared -o $(DIST_DIR)/$(SDKNAME).so ./sdk/next.cpp ./sdk/next_ios.cpp ./sdk/next_linux.cpp ./sdk/next_mac.cpp ./sdk/next_ps4.cpp ./sdk/next_switch.cpp ./sdk/next_windows.cpp ./sdk/next_xboxone.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-relay-backend
build-relay-backend: ## builds the relay_backend binary
	@printf "Building relay backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/relay_backend ./cmd/relay_backend/relay_backend.go
	@printf "done\n"

.PHONY: build-server-backend
build-server-backend: ## builds the server_backend binary
	@printf "Building server backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/server_backend ./cmd/server_backend/server_backend.go
	@printf "done\n"

.PHONY: build-server
build-server: build-sdk ## builds the game server linking in the sdk shared library
	@printf "Building server... "
	@$(CXX) -Isdk -o $(DIST_DIR)/server ./cmd/server/server.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-server
build-functional-server:
	@printf "Building functional server... "
	@$(CXX) -Isdk -o $(DIST_DIR)/func_server ./cmd/tools/functional/server/func_server.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-client
build-functional-client:
	@printf "Building functional client... "
	@$(CXX) -Isdk -o $(DIST_DIR)/func_client ./cmd/tools/functional/client/func_client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-client
build-client: build-sdk ## builds the game client linking in the sdk shared library
	@printf "Building client... "
	@$(CXX) -Isdk -o $(DIST_DIR)/client ./cmd/client/client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-all
build-all: build-optimizer build-relay-backend build-server-backend build-relay build-sdk-test build-tools ## builds everything
