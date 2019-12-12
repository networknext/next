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
test: clean lint build-relay build-sdk-test ## runs linters and all tests with coverage
	@$(DIST_DIR)/$(SDKNAME)_test
	@$(DIST_DIR)/relay test
	@$(GO) test -race -v ./core/...

.PHONY: build-sdk-test
build-sdk-test: build-sdk ## builds the sdk test binary
	@printf "Building sdk test... "
	@$(CXX) -Isdk -o $(DIST_DIR)/$(SDKNAME)_test ./sdk/next_test.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@echo "done"

PHONY: build-tools
build-tools: ## builds all the tools
	@./cmd/tools/build.sh

#####################
## MAIN COMPONENTS ##
#####################

.PHONY: dev-optimizer
dev-optimizer: ## runs a local optimizer
	$(GO) run cmd/optimizer/optimizer.go

.PHONY: dev-relay-backend
dev-relay-backend: ## runs a local relay_backend
	$(GO) run cmd/relay_backend/relay_backend.go

.PHONY: dev-server-backend
dev-server-backend: ## runs a local server_backend
	$(GO) run cmd/server_backend/server_backend.go

.PHONY: build-optimizer
build-optimizer: ## builds the optimizer binary
	@printf "Building optimizer... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/optimizer ./cmd/optimizer/optimizer.go
	@echo "done"

.PHONY: build-relay
build-relay: ## builds the relay
	@printf "Building relay... "
	@$(CXX) -o $(DIST_DIR)/relay ./cmd/relay/*.cpp $(LDFLAGS)
	@echo "done"

.PHONY: build-sdk
build-sdk: clean ## builds the sdk into a shared object for linking
	@printf "Building sdk... "
	@$(CXX) -fPIC -shared -o $(DIST_DIR)/$(SDKNAME).so ./sdk/next.cpp ./sdk/next_ios.cpp ./sdk/next_linux.cpp ./sdk/next_mac.cpp ./sdk/next_ps4.cpp ./sdk/next_switch.cpp ./sdk/next_windows.cpp ./sdk/next_xboxone.cpp $(LDFLAGS)
	@echo "done"

.PHONY: build-relay-backend
build-relay-backend: ## builds the relay_backend binary
	@printf "Building relay backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/relay_backend ./cmd/relay_backend/relay_backend.go
	@echo "done"

.PHONY: build-server-backend
build-server-backend: ## builds the server_backend binary
	@printf "Building server backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/server_backend ./cmd/server_backend/server_backend.go
	@echo "done"

.PHONY: build-server
build-server: build-sdk ## builds the game server linking in the sdk shared library
	@printf "Building server... "
	@$(CXX) -Isdk -o $(DIST_DIR)/server ./cmd/server/server.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@echo "done"

.PHONY: build-client
build-client: build-sdk ## builds the game client linking in the sdk shared library
	@printf "Building client... "
	@$(CXX) -Isdk -o $(DIST_DIR)/client ./cmd/client/client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@echo "done"

.PHONY: build-all
build-all: build-optimizer build-relay-backend build-server-backend build-relay build-sdk-test build-tools ## builds everything
