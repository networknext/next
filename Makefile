CXX = g++
CXX_FLAGS := -Wall -Wextra
GO = go
GOFMT = gofmt

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
DIST_DIR = ./dist

COST_FILE = $(DIST_DIR)/cost.bin
OPTIMIZE_FILE = $(DIST_DIR)/optimize.bin

##################
##    SDK ENV   ##
##################

export NEXT_LOG_LEVEL = 4
export NEXT_DATACENTER = local
export NEXT_CUSTOMER_PUBLIC_KEY = leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==
export NEXT_CUSTOMER_PRIVATE_KEY = leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn
export NEXT_HOSTNAME = 127.0.0.1
export NEXT_PORT = 40000    # Do not change. This must stay at 40000. The shipped SDK relies on this!

####################
##    RELAY ENV   ##
####################

ifndef RELAY_ADDRESS
export RELAY_BACKEND_HOSTNAME = http://127.0.0.1:30000
endif

ifndef RELAY_ADDRESS
export RELAY_ADDRESS = 127.0.0.1
endif

ifndef RELAY_ADDRESS
export RELAY_DEBUG = 0
endif

## Relay keys are unique to each relay and used to DECRYPT only the segment in the route token indended for itself
## For local dev purposes ALL relays we run will have the same keys, but in production they are all different 
ifndef RELAY_PUBLIC_KEY
export RELAY_PUBLIC_KEY = 9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=
endif

ifndef RELAY_PRIVATE_KEY
export RELAY_PRIVATE_KEY = lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=
endif

######################
##    BACKEND ENV   ##
######################

## Server backend keys are used for SIGNING data so game servers can verify response data's authenticity 
ifndef SERVER_BACKEND_PUBLIC_KEY
export SERVER_BACKEND_PUBLIC_KEY = TGHKjEeHPtSgtZfDyuDPcQgtJTyRDtRvGSKvuiWWo0A=
endif

ifndef SERVER_BACKEND_PRIVATE_KEY
export SERVER_BACKEND_PRIVATE_KEY = FXwFqzjGlIwUDwiq1N5Um5VUesdr4fP2hVV2cnJ+yARMYcqMR4c+1KC1l8PK4M9xCC0lPJEO1G8ZIq+6JZajQA==
endif

## Relay routing keys are used to ENCRYPT route tokens for the client, server, and all relays in between
ifndef RELAY_ROUTER_PUBLIC_KEY
export RELAY_ROUTER_PUBLIC_KEY = SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=
endif

ifndef RELAY_ROUTER_PRIVATE_KEY
export RELAY_ROUTER_PRIVATE_KEY = ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=
endif

## By default we set only error and warning logs for server_backend and relay_backend 
ifndef BACKEND_LOG_LEVEL
export BACKEND_LOG_LEVEL = warn
endif

ifndef ROUTE_MATRIX_URI
export ROUTE_MATRIX_URI = http://127.0.0.1:30000/route_matrix
endif

ifndef MAXMIND_DB_URI
export MAXMIND_DB_URI = ./testdata/GeoIP2-City-Test.mmdb
endif

ifndef REDIS_HOST
export REDIS_HOST = 127.0.0.1:6379
endif

.PHONY: help
help: ## this list
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## cleans the dist directory of all builds
	@rm -fr $(DIST_DIR)
	@mkdir $(DIST_DIR)

.PHONY: lint
lint: ## runs go vet
	@printf "Skipping vet/staticcheck for now...\n"

.PHONY: format
format: ## runs gofmt on all go source code
	@$(GOFMT) -s -w .
	@printf "\n"

#####################
## TESTS AND TOOLS ##
#####################

.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit: clean lint build-sdk-test build-relay ## runs unit tests
	@$(DIST_DIR)/$(SDKNAME)_test
	@$(DIST_DIR)/relay test
	@printf "Running go tests:\n\n"
	@$(GO) test  ./... -coverprofile ./cover.out
	@printf "\n\nCoverage results of go tests:\n\n"
	@$(GO) tool cover -func ./cover.out
	@printf "\n"

.PHONY: test-soak
test-soak: clean build-sdk-test build-soak-test ## runs soak test
	@$(DIST_DIR)/$(SDKNAME)_soak_test
	@printf "\n"

ifeq ($(OS),linux)
.PHONY: test-soak-valgrind
test-soak-valgrind: clean build-sdk-test build-soak-test
	@valgrind --tool=memcheck --leak-check=yes --show-reachable=yes --num-callers=20 --track-fds=yes --track-origins=yes $(DIST_DIR)/$(SDKNAME)_soak_test
	@printf "\n"
endif

.PHONY: test-func
test-func: clean build-sdk build-relay build-functional-server build-functional-client ## runs functional tests
	@printf "Building functional backend... " ; \
	go build -o ./dist/func_backend ./cmd/tools/functional/backend/*.go ; \
	printf "done\n" ; \
	printf "overriding RELAY_BACKEND_HOSTNAME\n" ; \
	export RELAY_BACKEND_HOSTNAME='http://localhost:30000' ; \
	printf "\nRunning functional tests...\n\n" ; \
	$(GO) run ./cmd/tools/functional/tests/func_tests.go ; \
	printf "\ndone\n\n"

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

.PHONY: dev-cost
dev-cost: ## gets the cost matrix from the local backend
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/cost ./cmd/tools/cost/cost.go
	$(DIST_DIR)/cost -url=http://localhost:30000/cost_matrix > $(COST_FILE)

.PHONY: dev-optimize
dev-optimize: ## transforms the cost matrix into a route matrix
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/optimize ./cmd/tools/optimize/optimize.go
	test -f $(COST_FILE) && cat $(COST_FILE) | ./dist/optimize -threshold-rtt=1 > $(OPTIMIZE_FILE)

.PHONY: dev-analyze
dev-analyze: ## analyzes the route matrix
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/analyze ./cmd/tools/analyze/analyze.go
	test -f $(OPTIMIZE_FILE) && cat $(OPTIMIZE_FILE) | $(DIST_DIR)/analyze

.PHONY: debug
dev-debug: ## debugs relay in route matrix
	test -f $(OPTIMIZE_FILE) && cat $(OPTIMIZE_FILE) | $(DIST_DIR)/debug -relay=$(relay)

.PHONY: dev-route
dev-route: ## prints routes from relay to datacenter in route matrix
	test -f $(OPTIMIZE_FILE) && cat $(OPTIMIZE_FILE) | $(DIST_DIR)/route -relay=$(relay) -datacenter=$(datacenter)

#######################
# Relay Build Process #
#######################

RELAY_DIR	:= ./cmd/relay
RELAY_EXE	:= relay

.PHONY: $(DIST_DIR)/$(RELAY_EXE)
$(DIST_DIR)/$(RELAY_EXE):

.PHONY: dev-relay
dev-relay: $(DIST_DIR)/$(RELAY_EXE) build-relay ## runs a SINGLE relay
	@$<

.PHONY: dev-multi-relays
dev-multi-relays: $(DIST_DIR)/$(RELAY_EXE) build-relay ## runs 10 relays, use ./relay-spawner.sh directly for more options
	./cmd/tools/scripts/relay-spawner.sh -n 10 -p 10000

#######################

.PHONY: dev-optimizer
dev-optimizer: ## runs a local optimizer
	$(GO) run cmd/optimizer/optimizer.go

.PHONY: dev-relay-backend
dev-relay-backend: ## runs a local relay backend
	@$(GO) run cmd/relay_backend/relay_backend.go

.PHONY: dev-server-backend
dev-server-backend: ## runs a local server backend
	@$(GO) run cmd/server_backend/server_backend.go

.PHONY: dev-backend
dev-backend: ## runs a local mock backend
	$(GO) run cmd/tools/functional/backend/*.go

.PHONY: dev-server
dev-server: build-sdk build-server  ## runs a local server
	@./dist/server

.PHONY: dev-client
dev-client: build-client  ## runs a local client
	@./dist/client

.PHONY: build-relay
build-relay: ## builds the relay
	@printf "Building relay... "
	@$(CXX) $(CXX_FLAGS) -o $(DIST_DIR)/$(RELAY_EXE) cmd/relay/*.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-sdk
build-sdk: ## builds the sdk into a shared object for linking
	@printf "Building sdk... "
	@$(CXX) -fPIC -shared -o $(DIST_DIR)/$(SDKNAME).so ./sdk/next.cpp ./sdk/next_ios.cpp ./sdk/next_linux.cpp ./sdk/next_mac.cpp ./sdk/next_ps4.cpp ./sdk/next_switch.cpp ./sdk/next_windows.cpp ./sdk/next_xboxone.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-relay-backend
build-relay-backend: ## builds the relay backend binary
	@printf "Building relay backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/relay_backend ./cmd/relay_backend/relay_backend.go
	@printf "done\n"

.PHONY: build-server-backend
build-server-backend: ## builds the server backend binary
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

.PHONY: build-functional
build-functional: build-functional-client build-functional-server

.PHONY: build-client
build-client: build-sdk ## builds the game client linking in the sdk shared library
	@printf "Building client... "
	@$(CXX) -Isdk -o $(DIST_DIR)/client ./cmd/client/client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-all
build-all: build-relay-backend build-server-backend build-relay build-client build-server build-functional build-sdk-test build-soak-test build-tools ## builds everything

.PHONY: rebuild-all
rebuild-all: clean build-all
