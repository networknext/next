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
DEPLOY_DIR = ./deploy
DIST_DIR = ./dist
ARTIFACT_BUCKET = gs://development_artifacts
ARTIFACT_BUCKET_PROD = gs://prod_artifacts
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

ifndef FIRESTORE_EMULATOR_HOST
export FIRESTORE_EMULATOR_HOST = 127.0.0.1:9000
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

.PHONY: build-portal
build-portal: ## builds the portal binary
	@printf "Building portal... \n"
	@printf "TIMESTAMP: ${TIMESTAMP}\n"
	@printf "SHA: ${SHA}\n"
	@printf "RELEASE: ${RELEASE}\n"
	@printf "COMMITMESSAGE: ${COMMITMESSAGE}\n"
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/portal ./cmd/portal/portal.go
	@printf "done\n"

.PHONY: build-portal-artifact
build-portal-artifact: build-portal ## builds the portal with the right env vars and creates a .tar.gz
	@printf "Building portal dev artifact... "
	@mkdir -p $(DIST_DIR)/artifact/portal
	@cp $(DIST_DIR)/portal $(DIST_DIR)/artifact/portal/app
	@cp -r ./cmd/portal/public $(DIST_DIR)/artifact/portal
	@cp ./cmd/portal/dev.env $(DIST_DIR)/artifact/portal/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/portal/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/portal && tar -zcf ../../portal.dev.tar.gz public app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/portal.dev.tar.gz\n"

.PHONY: build-portal-prod-artifact
build-portal-prod-artifact: build-portal ## builds the portal with the right env vars and creates a .tar.gz
	@printf "Building portal prod artifact... "
	@mkdir -p $(DIST_DIR)/artifact/portal
	@cp $(DIST_DIR)/portal $(DIST_DIR)/artifact/portal/app
	@cp -r ./cmd/portal/public $(DIST_DIR)/artifact/portal
	@cp ./cmd/portal/prod.env $(DIST_DIR)/artifact/portal/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/portal/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/portal && tar -zcf ../../portal.prod.tar.gz public app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/portal.prod.tar.gz\n"

.PHONY: publish-portal-artifact
publish-portal-artifact: ## publishes the portal artifact to GCP Storage with gsutil
	@printf "Publishing portal dev artifact... \n\n"
	@gsutil cp $(DIST_DIR)/portal.dev.tar.gz $(ARTIFACT_BUCKET)/portal.dev.tar.gz
	@gsutil setmeta -h "x-goog-meta-build-time:$(TIMESTAMP)" -h "x-goog-meta-sha:$(SHA)" -h "x-goog-meta-release:$(RELEASE)" -h "x-goog-meta-commitMessage:$(COMMITMESSAGE)" $(ARTIFACT_BUCKET)/portal.dev.tar.gz
	@printf "done\n"

.PHONY: publish-portal-prod-artifact
publish-portal-prod-artifact: ## publishes the portal artifact to GCP Storage with gsutil
	@printf "Publishing portal prod artifact... \n\n"
	@gsutil cp $(DIST_DIR)/portal.prod.tar.gz $(ARTIFACT_BUCKET_PROD)/portal.prod.tar.gz
	@gsutil setmeta -h "x-goog-meta-build-time:$(TIMESTAMP)" -h "x-goog-meta-sha:$(SHA)" -h "x-goog-meta-release:$(RELEASE)" -h "x-goog-meta-commitMessage:$(COMMITMESSAGE)" $(ARTIFACT_BUCKET_PROD)/portal.prod.tar.gz
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

.PHONY: build-relay-backend-artifact
build-relay-backend-artifact: build-relay-backend ## builds the relay backend and creates the dev artifact
	@printf "Building relay backend dev artifact... "
	@mkdir -p $(DIST_DIR)/artifact/relay_backend
	@cp $(DIST_DIR)/relay_backend $(DIST_DIR)/artifact/relay_backend/app
	@cp ./cmd/relay_backend/dev.env $(DIST_DIR)/artifact/relay_backend/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/relay_backend/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/relay_backend && tar -zcf ../../relay_backend.dev.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/relay_backend.dev.tar.gz\n"

.PHONY: build-relay-backend-prod-artifact
build-relay-backend-prod-artifact: build-relay-backend ## builds the relay backend and creates the prod artifact
	@printf "Building relay backend prod artifact... "
	@mkdir -p $(DIST_DIR)/artifact/relay_backend
	@cp $(DIST_DIR)/relay_backend $(DIST_DIR)/artifact/relay_backend/app
	@cp ./cmd/relay_backend/prod.env $(DIST_DIR)/artifact/relay_backend/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/relay_backend/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/relay_backend && tar -zcf ../../relay_backend.prod.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/relay_backend.prod.tar.gz\n"

.PHONY: publish-relay-backend-artifact
publish-relay-backend-artifact: ## publishes the relay backend dev artifact
	@printf "Publishing relay backend dev artifact... \n\n"
	@gsutil cp $(DIST_DIR)/relay_backend.dev.tar.gz $(ARTIFACT_BUCKET)/relay_backend.dev.tar.gz
	@gsutil setmeta -h "x-goog-meta-build-time:$(TIMESTAMP)" -h "x-goog-meta-sha:$(SHA)" -h "x-goog-meta-release:$(RELEASE)" -h "x-goog-meta-commitMessage:$(COMMITMESSAGE)" $(ARTIFACT_BUCKET)/relay_backend.dev.tar.gz
	@printf "done\n"

.PHONY: publish-relay-backend-prod-artifact
publish-relay-backend-prod-artifact: ## publishes the relay backend prod artifact
	@printf "Publishing relay backend prod artifact... \n\n"
	@gsutil cp $(DIST_DIR)/relay_backend.prod.tar.gz $(ARTIFACT_BUCKET_PROD)/relay_backend.prod.tar.gz
	@gsutil setmeta -h "x-goog-meta-build-time:$(TIMESTAMP)" -h "x-goog-meta-sha:$(SHA)" -h "x-goog-meta-release:$(RELEASE)" -h "x-goog-meta-commitMessage:$(COMMITMESSAGE)" $(ARTIFACT_BUCKET_PROD)/relay_backend.prod.tar.gz
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

.PHONY: build-billing-artifact
build-billing-artifact: build-billing ## builds the billing service and creates the dev artifact
	@printf "Building billing dev artifact..."
	@mkdir -p $(DIST_DIR)/artifact/billing
	@cp $(DIST_DIR)/billing $(DIST_DIR)/artifact/billing/app
	@cp ./cmd/billing/dev.env $(DIST_DIR)/artifact/billing/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/billing/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/billing && tar -zcf ../../billing.dev.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/billing.dev.tar.gz\n"

.PHONY: build-billing-prod-artifact
build-billing-prod-artifact: build-billing ## builds the belling service and creates the prod artifact
	@printf "Building billing prod artifact... "
	@mkdir -p $(DIST_DIR)/artifact/billing
	@cp $(DIST_DIR)/billing $(DIST_DIR)/artifact/billing/app
	@cp ./cmd/billing/prod.env $(DIST_DIR)/artifact/billing/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/billing/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/billing && tar -zcf ../../billing.prod.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/billing.prod.tar.gz\n"

.PHONY: publish-billing-artifact
publish-billing-artifact: ## publishes the billing dev artifact
	@printf "Publishing billing dev artifact... \n\n"
	@gsutil cp $(DIST_DIR)/billing.dev.tar.gz $(ARTIFACT_BUCKET)/billing.dev.tar.gz
	@gsutil setmeta -h "x-goog-meta-build-time:$(TIMESTAMP)" -h "x-goog-meta-sha:$(SHA)" -h "x-goog-meta-release:$(RELEASE)" -h "x-goog-meta-commitMessage:$(COMMITMESSAGE)" $(ARTIFACT_BUCKET)/billing.dev.tar.gz
	@printf "done\n"

.PHONY: publish-billing-prod-artifact
publish-billing-prod-artifact: ## publishes the billing prod artifact
	@printf "Publishing billing prod artifact... \n\n"
	@gsutil cp $(DIST_DIR)/billing.prod.tar.gz $(ARTIFACT_BUCKET_PROD)/billing.prod.tar.gz
	@gsutil setmeta -h "x-goog-meta-build-time:$(TIMESTAMP)" -h "x-goog-meta-sha:$(SHA)" -h "x-goog-meta-release:$(RELEASE)" -h "x-goog-meta-commitMessage:$(COMMITMESSAGE)" $(ARTIFACT_BUCKET_PROD)/billing.prod.tar.gz
	@printf "done\n"

.PHONY: build-server-backend-artifact
build-server-backend-artifact: build-server-backend ## builds the server backend and creates the dev artifact
	@printf "Building server backend dev artifact..."
	@mkdir -p $(DIST_DIR)/artifact/server_backend
	@cp $(DIST_DIR)/server_backend $(DIST_DIR)/artifact/server_backend/app
	@cp ./cmd/server_backend/dev.env $(DIST_DIR)/artifact/server_backend/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/server_backend/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/server_backend && tar -zcf ../../server_backend.dev.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/server_backend.dev.tar.gz\n"

.PHONY: build-server-backend-prod-artifact
build-server-backend-prod-artifact: build-server-backend ## builds the server backend and creates the prod artifact
	@printf "Building server backend prod artifact... "
	@mkdir -p $(DIST_DIR)/artifact/server_backend
	@cp $(DIST_DIR)/server_backend $(DIST_DIR)/artifact/server_backend/app
	@cp ./cmd/server_backend/prod.env $(DIST_DIR)/artifact/server_backend/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/server_backend/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/server_backend && tar -zcf ../../server_backend.prod.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/server_backend.prod.tar.gz\n"

.PHONY: publish-server-backend-artifact
publish-server-backend-artifact: ## publishes the server backend dev artifact
	@printf "Publishing server backend dev artifact... \n\n"
	@gsutil cp $(DIST_DIR)/server_backend.dev.tar.gz $(ARTIFACT_BUCKET)/server_backend.dev.tar.gz
	@gsutil setmeta -h "x-goog-meta-build-time:$(TIMESTAMP)" -h "x-goog-meta-sha:$(SHA)" -h "x-goog-meta-release:$(RELEASE)" -h "x-goog-meta-commitMessage:$(COMMITMESSAGE)" $(ARTIFACT_BUCKET)/server_backend.dev.tar.gz
	@printf "done\n"

.PHONY: publish-server-backend-prod-artifact
publish-server-backend-prod-artifact: ## publishes the server backend prod artifact
	@printf "Publishing server backend prod artifact... \n\n"
	@gsutil cp $(DIST_DIR)/server_backend.prod.tar.gz $(ARTIFACT_BUCKET_PROD)/server_backend.prod.tar.gz
	@gsutil setmeta -h "x-goog-meta-build-time:$(TIMESTAMP)" -h "x-goog-meta-sha:$(SHA)" -h "x-goog-meta-release:$(RELEASE)" -h "x-goog-meta-commitMessage:$(COMMITMESSAGE)" $(ARTIFACT_BUCKET_PROD)/server_backend.prod.tar.gz
	@printf "done\n"

.PHONY: deploy-server-backend
deploy-server-backend: ## builds and deploys the server backend to dev
	@printf "Deploying server backend to dev... \n\n"
	gcloud compute --project "network-next-v3-dev" ssh server-backend-dev-1 -- 'cd /app && sudo ./bootstrap.sh -b $(ARTIFACT_BUCKET) -a server_backend.dev.tar.gz'

.PHONY: build-backend-artifacts
build-backend-artifacts: build-portal-artifact build-relay-backend-artifact build-server-backend-artifact ## builds the backend artifacts

.PHONY: build-relay-artifact
build-relay-artifact: build-relay ## builds the relay and creates the dev artifact
	@printf "Building relay artifact..."
	@mkdir -p $(DIST_DIR)/artifact/relay
	@cp $(DIST_DIR)/relay $(DIST_DIR)/artifact/relay/relay
	@cp $(DEPLOY_DIR)/relay/relay.service $(DIST_DIR)/artifact/relay/relay.service
	@cp $(DEPLOY_DIR)/relay/install.sh $(DIST_DIR)/artifact/relay/install.sh
	@cd $(DIST_DIR)/artifact/relay && tar -zcf ../../relay.dev.tar.gz relay relay.service install.sh && cd ../..
	@printf "$(DIST_DIR)/relay.dev.tar.gz\n"

.PHONY: build-relay-prod-artifact
build-relay-prod-artifact: build-relay ## builds the relay and creates the prod artifact
	@printf "Building relay artifact..."
	@mkdir -p $(DIST_DIR)/artifact/relay
	@cp $(DIST_DIR)/relay $(DIST_DIR)/artifact/relay/relay
	@cp $(DEPLOY_DIR)/relay/relay.service $(DIST_DIR)/artifact/relay/relay.service
	@cp $(DEPLOY_DIR)/relay/install.sh $(DIST_DIR)/artifact/relay/install.sh
	@cd $(DIST_DIR)/artifact/relay && tar -zcf ../../relay.prod.tar.gz relay relay.service install.sh && cd ../..
	@printf "$(DIST_DIR)/relay.prod.tar.gz\n"

.PHONY: publish-relay-artifact
publish-relay-artifact: ## publishes the dev relay artifact
	@printf "Publishing relay artifact... \n\n"
	@gsutil cp $(DIST_DIR)/relay.dev.tar.gz $(ARTIFACT_BUCKET)/relay.dev.tar.gz
	@gsutil acl set public-read $(ARTIFACT_BUCKET)/relay.dev.tar.gz
	@gsutil setmeta \
	-h 'Content-Type:application/xtar' \
	-h 'Cache-Control:no-cache, max-age=0' \
	$(ARTIFACT_BUCKET)/relay.dev.tar.gz
	@printf "done\n"

.PHONY: publish-relay-prod-artifact
publish-relay-prod-artifact: ## publishes the prod relay artifact
	@printf "Publishing relay artifact... \n\n"
	@gsutil cp $(DIST_DIR)/relay.prod.tar.gz $(ARTIFACT_BUCKET_PROD)/relay.prod.tar.gz
	@gsutil acl set public-read $(ARTIFACT_BUCKET_PROD)/relay.prod.tar.gz
	@gsutil setmeta \
	-h 'Content-Type:application/xtar' \
	-h 'Cache-Control:no-cache, max-age=0' \
	$(ARTIFACT_BUCKET_PROD)/relay.prod.tar.gz
	@printf "done\n"

.PHONY: publish-bootstrap-script
publish-bootstrap-script:
	@printf "Publishing bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/bootstrap.sh $(ARTIFACT_BUCKET)/bootstrap.sh
	@printf "done\n"

.PHONY: publish-bootstrap-script-prod
publish-bootstrap-script-prod:
	@printf "Publishing bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/bootstrap.sh $(ARTIFACT_BUCKET_PROD)/bootstrap.sh
	@printf "done\n"

.PHONY: build-backend-prod-artifacts
build-backend-prod-artifacts: build-portal-prod-artifact build-relay-backend-prod-artifact build-server-backend-prod-artifact ## builds the backend artifacts

.PHONY: publish-backend-artifacts
publish-backend-artifacts: publish-portal-artifact publish-relay-backend-artifact publish-server-backend-artifact ## publishes the backend artifacts to GCP Storage with gsutil

.PHONY: publish-backend-prod-artifacts
publish-backend-prod-artifacts: publish-portal-prod-artifact publish-relay-backend-prod-artifact publish-server-backend-prod-artifact ## publishes the backend artifacts to GCP Storage with gsutil

.PHONY: deploy-backend
deploy-backend: deploy-portal deploy-relay-backend deploy-server-backend

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
build-all: build-billing build-relay-backend build-server-backend build-relay-ref build-client build-server build-functional build-next ## builds everything

.PHONY: rebuild-all
rebuild-all: clean build-all

.PHONY: update-sdk
update-sdk: ## updates the sdk submodule
	git submodule update --remote --merge

.PHONY: sync
sync: ## syncs to latest code including submodules
	git pull && git submodule update --recursive
