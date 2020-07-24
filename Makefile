CXX = g++
CXX_FLAGS := -Wall -Wextra -std=c++17
GO = go
GOFMT = gofmt

OS := $(shell uname -s | tr A-Z a-z)
ifeq ($(OS),darwin)
	LDFLAGS = -lsodium -lcurl -lpthread -lm -framework CoreFoundation -framework SystemConfiguration -DNEXT_DEVELOPMENT
else
	LDFLAGS = -lsodium -lcurl -lpthread -lm -DNEXT_DEVELOPMENT
endif

SDKNAME = libnext

TIMESTAMP ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA ?= $(shell git rev-parse --short HEAD)
RELEASE ?= $(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE ?= $(shell git log -1 --pretty=%B | tr '\n' ' ')

CURRENT_DIR = $(shell pwd -P)
DIST_DIR = ./dist
ARTIFACT_BUCKET = gs://development_artifacts
SYSTEMD_SERVICE_FILE = app.service

COST_FILE = $(DIST_DIR)/cost.bin
OPTIMIZE_FILE = $(DIST_DIR)/optimize.bin

export ENV = local

##################
##    SDK ENV   ##
##################

export NEXT_LOG_LEVEL = 4
export NEXT_DATACENTER = local
export NEXT_CUSTOMER_PUBLIC_KEY = leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==
export NEXT_CUSTOMER_PRIVATE_KEY = leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn
export NEXT_HOSTNAME = 127.0.0.1
export NEXT_PORT = 40000

####################
##    RELAY ENV   ##
####################

ifndef RELAY_BACKEND_HOSTNAME
export RELAY_BACKEND_HOSTNAME = http://127.0.0.1:30000
endif

ifndef RELAY_ADDRESS
export RELAY_ADDRESS = 127.0.0.1:0
endif

ifndef RELAY_DEBUG
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

## Relay routing keys are used to ENCRYPT and SIGN route tokens sent to a relay
ifndef RELAY_ROUTER_PUBLIC_KEY
export RELAY_ROUTER_PUBLIC_KEY = SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=
endif

ifndef RELAY_ROUTER_PRIVATE_KEY
export RELAY_ROUTER_PRIVATE_KEY = ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=
endif

## Maximum allowed jitter and packet loss during the relay backend's cost matrix generation process
ifndef RELAY_ROUTER_MAX_JITTER
export RELAY_ROUTER_MAX_JITTER = 10.0
endif

ifndef RELAY_ROUTER_MAX_PACKET_LOSS
export RELAY_ROUTER_MAX_PACKET_LOSS = 0.1
endif

## By default we set only error and warning logs for server_backend and relay_backend
ifndef BACKEND_LOG_LEVEL
export BACKEND_LOG_LEVEL = warn
endif

ifndef ROUTE_MATRIX_URI
export ROUTE_MATRIX_URI = http://127.0.0.1:30000/route_matrix
endif

ifndef ROUTE_MATRIX_SYNC_INTERVAL
export ROUTE_MATRIX_SYNC_INTERVAL = 1s
endif

ifndef COST_MATRIX_INTERVAL
export COST_MATRIX_INTERVAL = 1s
endif

ifndef MAXMIND_CITY_DB_URI
export MAXMIND_CITY_DB_URI = ./testdata/GeoIP2-City-Test.mmdb
endif

ifndef MAXMIND_ISP_DB_URI
export MAXMIND_ISP_DB_URI = ./testdata/GeoIP2-ISP-Test.mmdb
endif

ifndef SESSION_MAP_INTERVAL
export SESSION_MAP_INTERVAL = 1s
endif

ifndef REDIS_HOST_PORTAL
export REDIS_HOST_PORTAL = 127.0.0.1:6379
endif

ifndef REDIS_HOST_PORTAL_EXPIRATION
export REDIS_HOST_PORTAL_EXPIRATION = 30s
endif

ifndef REDIS_HOST_RELAYS
export REDIS_HOST_RELAYS = 127.0.0.1:6379
endif

ifndef REDIS_HOST_CACHE
export REDIS_HOST_CACHE = 127.0.0.1:6379
endif

ifndef AUTH_DOMAIN
export AUTH_DOMAIN = networknext.auth0.com
endif
ifndef AUTH_CLIENTID
export AUTH_CLIENTID = KxEiJeUh5tE1cZrI64GXHs455XcxRDKX
endif
ifndef AUTH_CLIENTSECRET
export AUTH_CLIENTSECRET = d6w4zWBUT07UQlpDIA52pBMDukeuhvWJjCEnHWkkkZypd453qRn4e18Nz84GkfkO
endif

ifndef GOOGLE_FIRESTORE_SYNC_INTERVAL
export GOOGLE_FIRESTORE_SYNC_INTERVAL = 1s
endif

ifndef PORTAL_CRUNCHER_HOST
export PORTAL_CRUNCHER_HOST = tcp://127.0.0.1:5555
endif

.PHONY: help
help: ## this list
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\033[36m\1\\033[m:\2/' | column -c2 -t -s :)"

.PHONY: clean
clean: ## cleans the dist directory of all builds
	@rm -fr $(DIST_DIR)
	@mkdir $(DIST_DIR)

.PHONY: format
format: ## runs gofmt on all go source code
	@$(GOFMT) -s -w .
	@printf "\n"

#####################
## TESTS AND TOOLS ##
#####################

.PHONY: test
test: test-unit

ifeq ($(OS),linux)
.PHONY: test-unit-relay
test-unit-relay: build-relay-tests ## runs relay unit tests
	@$(RELAY_DIR)/bin/relay.test
endif

.PHONY: test-unit-backend
test-unit-backend: ## runs backend unit tests
	@./scripts/test-unit-backend.sh

.PHONY: test-unit
test-unit: clean test-unit-backend ## runs backend unit tests

ifeq ($(OS),linux)
.PHONY: test-soak-valgrind
test-soak-valgrind: clean build-soak-test ## runs sdk soak test under valgrind (linux only)
	@valgrind --tool=memcheck --leak-check=yes --show-reachable=yes --num-callers=20 --track-fds=yes --track-origins=yes $(DIST_DIR)/$(SDKNAME)_soak_test
	@printf "\n"
endif

.PHONY: build-functional-backend
build-functional-backend: ## builds the functional backend
	@printf "Building functional backend... " ; \
	$(GO) build -o ./dist/func_backend ./cmd/func_backend/*.go ; \
	printf "done\n" ; \

.PHONY: build-functional-tests
build-functional-tests: ## builds functional tests
	@printf "Building functional tests... " ; \
	$(GO) build -o ./dist/func_tests ./cmd/func_tests/*.go ; \
	printf "done\n" ; \

.PHONY: build-test-func
build-test-func: clean build-sdk build-relay-ref build-functional-server build-functional-client build-functional-backend build-functional-tests ## builds the functional tests

.PHONY: run-test-func
run-test-func:
	@printf "\nRunning functional tests...\n\n" ; \
	$(GO) run ./cmd/func_tests/func_tests.go $(tests) ; \
	printf "\ndone\n\n"

.PHONY: test-func
test-func: build-test-func run-test-func ## runs functional tests

.PHONY: build-test-func-parallel
build-test-func-parallel:
	@docker build -t func_tests -f ./cmd/func_tests/Dockerfile .

.PHONY: run-test-func-parallel
run-test-func-parallel:
	@./scripts/test-func-parallel.sh

.PHONY: test-func-parallel
test-func-parallel: build-test-func-parallel run-test-func-parallel ## runs functional tests in parallel

#######################
# Relay Build Process #
#######################

RELAY_DIR := ./cmd/relay
RELAY_MAKEFILE := Makefile
RELAY_EXE := relay

.PHONY: build-relay-ref
build-relay-ref: ## builds the reference relay
	@printf "Building reference relay... "
	@$(CXX) $(CXX_FLAGS) -o $(DIST_DIR)/reference_relay reference/relay/*.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-relay
build-relay: ## builds the relay
	@printf "Building relay... "
	@mkdir -p $(DIST_DIR)
	@cd $(RELAY_DIR) && $(MAKE) release
	@echo "done"

.PHONY: build-relay-tests
build-relay-tests: ## builds the relay unit tests (linux only)
	@printf "Building relay with tests enabled... "
	@cd $(RELAY_DIR) && $(MAKE) test
	@echo "done"

.PHONY: dev-relay
dev-relay: build-relay ## runs a local relay
	@$(DIST_DIR)/$(RELAY_EXE)

.PHONY: dev-multi-relays
dev-multi-relays: build-relay ## runs 10 local relays
	./scripts/relay-spawner.sh -n 10 -p 10000

#######################

.PHONY: dev-optimizer
dev-optimizer: ## runs a local optimizer
	$(GO) run cmd/optimizer/optimizer.go

.PHONY: dev-portal
dev-portal: build-portal ## runs a local portal web server
	@PORT=20000 BASIC_AUTH_USERNAME=local BASIC_AUTH_PASSWORD=local UI_DIR=./cmd/portal/public ./dist/portal

.PHONY: dev-relay-backend
dev-relay-backend: build-relay-backend ## runs a local relay backend
	@PORT=30000 ./dist/relay_backend

.PHONY: dev-server-backend
dev-server-backend: build-server-backend ## runs a local server backend
	@HTTP_PORT=40000 UDP_PORT=40000 ./dist/server_backend

.PHONY: dev-billing
dev-billing: build-billing ## runs a local billing service
	@PORT=41000 ./dist/billing

.PHONY: dev-portal-cruncher
dev-portal-cruncher: build-portal-cruncher ## runs a local portal cruncher
	@HTTP_PORT=42000 CRUNCHER_PORT=5555 ./dist/portal_cruncher

.PHONY: dev-reference-backend
dev-reference-backend: ## runs a local reference backend
	$(GO) run reference/backend/*.go

.PHONY: dev-reference-relay
dev-reference-relay: build-relay-ref ## runs a local reference relay
	@$(DIST_DIR)/reference_relay

.PHONY: dev-server
dev-server: build-sdk build-server  ## runs a local server
	@./dist/server

.PHONY: dev-client
dev-client: build-client  ## runs a local client
	@./dist/client

.PHONY: dev-multi-clients
dev-multi-clients: build-client ## runs 20 local clients
	./scripts/client-spawner.sh -n 20

$(DIST_DIR)/$(SDKNAME).so:
	@printf "Building sdk... "
	@$(CXX) -fPIC -Isdk/include -shared -o $(DIST_DIR)/$(SDKNAME).so ./sdk/source/next.cpp ./sdk/source/next_ios.cpp ./sdk/source/next_linux.cpp ./sdk/source/next_mac.cpp ./sdk/source/next_ps4.cpp ./sdk/source/next_switch.cpp ./sdk/source/next_windows.cpp ./sdk/source/next_xboxone.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-sdk
build-sdk: $(DIST_DIR)/$(SDKNAME).so ## builds the sdk


PHONY: build-portal-cruncher
build-portal-cruncher: ## builds the portal_cruncher binary
	@printf "Building portal_cruncher... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/portal_cruncher ./cmd/portal_cruncher/portal_cruncher.go
	@printf "done\n"

.PHONY: build-portal
build-portal: ## builds the portal binary
	@printf "Building portal... \n"
	@printf "TIMESTAMP: ${TIMESTAMP}\n"
	@printf "SHA: ${SHA}\n"
	@printf "RELEASE: ${RELEASE}\n"
	@printf "COMMITMESSAGE: ${COMMITMESSAGE}\n"
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/portal ./cmd/portal/portal.go
	@printf "done\n"

.PHONY: deploy-portal
deploy-portal: ## builds and deploys the portal to dev
	@printf "Deploying portal to dev... \n\n"
	gcloud compute --project "network-next-v3-dev" ssh portal-dev-1 -- 'cd /app && sudo ./bootstrap.sh -b $(ARTIFACT_BUCKET) -a portal.dev.tar.gz'

.PHONY: build-relay-backend
build-relay-backend: ## builds the relay backend binary
	@printf "Building relay backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/relay_backend ./cmd/relay_backend/relay_backend.go
	@printf "done\n"

.PHONY: deploy-relay-backend
deploy-relay-backend: ## builds and deploys the relay backend to dev
	@printf "Deploying relay backend to dev... \n\n"
	gcloud compute --project "network-next-v3-dev" ssh relay-backend-dev-1 -- 'cd /app && sudo ./bootstrap.sh -b $(ARTIFACT_BUCKET) -a relay_backend.dev.tar.gz'

.PHONY: build-server-backend
build-server-backend: ## builds the server backend binary
	@printf "Building server backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/server_backend ./cmd/server_backend/server_backend.go
	@printf "done\n"

.PHONY: build-billing
build-billing: ## builds the billing binary
	@printf "Building billing... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/billing ./cmd/billing/billing.go
	@printf "done\n"

.PHONY: deploy-server-backend
deploy-server-backend: ## builds and deploys the server backend to dev
	@printf "Deploying server backend to dev... \n\n"
	gcloud compute --project "network-next-v3-dev" ssh server-backend-dev-1 -- 'cd /app && sudo ./bootstrap.sh -b $(ARTIFACT_BUCKET) -a server_backend.dev.tar.gz'

.PHONY: build-backend-artifacts-dev
build-backend-artifacts-dev: ## builds the backend artifacts dev
	./deploy/build-artifacts.sh -e dev

.PHONY: build-backend-artifacts-staging
build-backend-artifacts-staging: ## builds the backend artifacts staging
	./deploy/build-artifacts.sh -e staging

.PHONY: build-backend-artifacts-prod
build-backend-artifacts-prod: ## builds the backend artifacts prod
	./deploy/build-artifacts.sh -e prod

.PHONY: publish-backend-artifacts-dev
publish-backend-artifacts-dev: ## publishes the backend artifacts to GCP Storage with gsutil dev
	./deploy/publish.sh -e dev -b gs://development_artifacts

.PHONY: publish-backend-artifacts-staging
publish-backend-artifacts-staging: ## publishes the backend artifacts to GCP Storage with gsutil staging
	./deploy/publish.sh -e staging -b gs://staging_artifacts

.PHONY: publish-backend-artifacts-prod
publish-backend-artifacts-prod: ## publishes the backend artifacts to GCP Storage with gsutil staging
	./deploy/publish.sh -e prod -b gs://prod_artifacts

.PHONY: build-server
build-server: build-sdk ## builds the server
	@printf "Building server... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/server ./cmd/server/server.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-server
build-functional-server:
	@printf "Building functional server... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/func_server ./cmd/func_server/func_server.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-client
build-functional-client:
	@printf "Building functional client... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/func_client ./cmd/func_client/func_client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional
build-functional: build-functional-client build-functional-server build-functional-backend build-functional-tests

.PHONY: build-client
build-client: build-sdk ## builds the client
	@printf "Building client... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/client ./cmd/client/client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-next
build-next: ## builds the operator tool
	@printf "Building operator tool... "
	@$(GO) build -o ./dist/next ./cmd/next/*.go
	@printf "done\n"

.PHONY: build-all
build-all: build-portal-cruncher build-billing build-relay-backend build-server-backend build-relay-ref build-client build-server build-functional build-next ## builds everything

.PHONY: rebuild-all
rebuild-all: clean build-all

.PHONY: update-sdk
update-sdk: ## updates the sdk submodule
	git submodule update --remote --merge

.PHONY: sync
sync: ## syncs to latest code including submodules
	git pull && git submodule update --recursive
