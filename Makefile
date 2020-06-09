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
RELEASE ?= $(shell git describe --tags --exact-match 2> /dev/null || git rev-parse --short HEAD)

CURRENT_DIR = $(shell pwd -P)
DEPLOY_DIR = ./deploy
DIST_DIR = ./dist
ARTIFACT_BUCKET = gs://artifacts.network-next-v3-dev.appspot.com
ARTIFACT_BUCKET_PROD = gs://us.artifacts.network-next-v3-prod.appspot.com
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

## These variables are only used for old backend compatability

ifndef RELAY_V3_ENABLED
export RELAY_V3_ENABLED = 0
endif

ifndef RELAY_V3_BACKEND_HOSTNAME
export RELAY_V3_BACKEND_HOSTNAME = localhost
endif

ifndef RELAY_V3_BACKEND_PORT
export RELAY_V3_BACKEND_PORT = 40002
endif

ifndef RELAY_V3_UPDATE_KEY
export RELAY_V3_UPDATE_KEY = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==
endif

ifndef RELAY_V3_SPEED
export RELAY_V3_SPEED = 10000
endif

ifndef RELAY_V3_NAME
export RELAY_V3_NAME = some name
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

ifndef REDIS_HOST_PORTAL
export REDIS_HOST_PORTAL = 127.0.0.1:6379
endif

ifndef REDIS_HOST_PORTAL_EXPIRATION
export REDIS_HOST_PORTAL_EXPIRATION = 1m
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

.PHONY: help
help: ## this list
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\033[36m\1\\033[m:\2/' | column -c2 -t -s :)"

.PHONY: clean
clean: ## cleans the dist directory of all builds
	@rm -fr $(DIST_DIR)
	@mkdir $(DIST_DIR)

.PHONY: lint
lint: ## runs go vet
	@printf "Skipping vet/staticcheck for now...\n\n"

.PHONY: format
format: ## runs gofmt on all go source code
	@$(GOFMT) -s -w .
	@printf "\n"

#####################
## TESTS AND TOOLS ##
#####################

.PHONY: test
test: test-unit

.PHONY: test-unit-sdk
test-unit-sdk: build-sdk-test ## runs sdk unit tests
	@$(DIST_DIR)/$(SDKNAME)_test

.PHONY: test-unit-relay
test-unit-relay-new: build-relay-tests ## runs relay unit tests
	@$(NEW_RELAY_DIR)/bin/relay.test

.PHONY: test-unit-reference-relay
test-unit-relay-ref: build-relay-ref
	@$(DIST_DIR)/relay test

.PHONY: test-unit-backend
test-unit-backend: lint ## runs backend unit tests
	@./cmd/tools/scripts/test-unit-backend.sh

.PHONY: test-unit
test-unit: clean test-unit-sdk test-unit-relay test-unit-reference-relay test-unit-backend ## runs all unit tests

.PHONY: test-soak
test-soak: clean build-sdk-test build-soak-test ## runs soak test
	@$(DIST_DIR)/$(SDKNAME)_soak_test
	@printf "\n"

ifeq ($(OS),linux)
.PHONY: test-soak-valgrind
test-soak-valgrind: clean build-sdk-test build-soak-test ## runs sdk soak test
	@valgrind --tool=memcheck --leak-check=yes --show-reachable=yes --num-callers=20 --track-fds=yes --track-origins=yes $(DIST_DIR)/$(SDKNAME)_soak_test
	@printf "\n"
endif

.PHONY: build-functional-backend
build-functional-backend: ## builds the functional backend
	@printf "Building functional backend... " ; \
	$(GO) build -o ./dist/func_backend ./cmd/tools/functional/backend/*.go ; \
	printf "done\n" ; \

.PHONY: build-test-func
build-test-func: clean build-sdk build-relay-ref build-functional-server build-functional-client build-functional-backend ## builds the functional tests

.PHONY: run-test-func
run-test-func:
	@printf "\nRunning functional tests...\n\n" ; \
	$(GO) run ./cmd/tools/functional/tests/func_tests.go $(tests) ; \
	printf "\ndone\n\n"

.PHONY: test-func
test-func: build-test-func run-test-func ## runs functional tests

.PHONY: build-test-func-parallel
build-test-func-parallel:
	@docker build -t func_tests -f ./cmd/tools/functional/tests/Dockerfile .

.PHONY: run-test-func-parallel
run-test-func-parallel:
	@./cmd/tools/scripts/test-func-parallel.sh

.PHONY: test-func-parallel
test-func-parallel: build-test-func-parallel run-test-func-parallel ## runs functional tests in parallel

.PHONY: build-sdk-test
build-sdk-test: build-sdk ## builds the sdk test binary
	@printf "Building sdk test... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/$(SDKNAME)_test ./sdk/test.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-soak-test
build-soak-test: build-sdk ## builds the sdk test binary
	@printf "Building soak test... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/$(SDKNAME)_soak_test ./sdk/soak.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
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
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/debug ./cmd/tools/debug/debug.go
	test -f $(OPTIMIZE_FILE) && cat $(OPTIMIZE_FILE) | $(DIST_DIR)/debug -relay=$(relay)

.PHONY: dev-route
dev-route: ## prints routes from relay to datacenter in route matrix
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA)" -o ${DIST_DIR}/route ./cmd/tools/route/route.go
	test -f $(OPTIMIZE_FILE) && cat $(OPTIMIZE_FILE) | $(DIST_DIR)/route -relay=$(relay) -datacenter=$(datacenter)

#######################
# Relay Build Process #
#######################

NEW_RELAY_DIR := ./cmd/relay_new
NEW_RELAY_MAKEFILE := Makefile
RELAY_EXE := relay

.PHONY: build-relay-ref
build-relay-ref: ## builds the reference relay
	@printf "Building reference relay... "
	@$(CXX) $(CXX_FLAGS) -o $(DIST_DIR)/reference_relay reference/relay/*.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-relay-new
build-relay-new: ## builds the new relay
	@printf "Building new relay... "
	@mkdir -p $(DIST_DIR)
	@cd $(NEW_RELAY_DIR) && $(MAKE) release
	@echo "done"

.PHONY: build-relay-tests
build-relay-tests: ## builds the new relay tests
	@printf "Building relay with tests enabled... "
	@cd $(NEW_RELAY_DIR) && $(MAKE) test
	@echo "done"

.PHONY: dev-relay
dev-relay: build-relay-new ## runs a local relay
	@$(DIST_DIR)/$(RELAY_EXE)

.PHONY: dev-multi-relays
dev-multi-relays: build-relay-new ## runs 10 local relays
	./cmd/tools/scripts/relay-spawner.sh -n 10 -p 10000

#######################

.PHONY: dev-optimizer
dev-optimizer: ## runs a local optimizer
	$(GO) run cmd/optimizer/optimizer.go

.PHONY: dev-portal
dev-portal: build-portal ## runs a local portal web server
	@PORT=20000 BASIC_AUTH_USERNAME=local BASIC_AUTH_PASSWORD=local UI_DIR=./cmd/portal/public ./dist/portal

.PHONY: dev-relay-backend
dev-relay-backend: ## runs a local relay backend
	@PORT=30000 $(GO) run cmd/relay_backend/relay_backend.go

.PHONY: dev-server-backend
dev-server-backend: ## runs a local server backend
	@PORT=40000 $(GO) run cmd/server_backend/server_backend.go

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
	./cmd/tools/scripts/client-spawner.sh -n 20

.PHONY: dev-relay-backend-old
dev-relay-backend-old:
	$(GO) run ./cmd/tools/functional/backend_old/*.go


$(DIST_DIR)/$(SDKNAME).so:
	@printf "Building sdk... "
	@$(CXX) -fPIC -Isdk/include -shared -o $(DIST_DIR)/$(SDKNAME).so ./sdk/source/next.cpp ./sdk/source/next_ios.cpp ./sdk/source/next_linux.cpp ./sdk/source/next_mac.cpp ./sdk/source/next_ps4.cpp ./sdk/source/next_switch.cpp ./sdk/source/next_windows.cpp ./sdk/source/next_xboxone.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-sdk
build-sdk: $(DIST_DIR)/$(SDKNAME).so ## builds the sdk

.PHONY: build-portal
build-portal: ## builds the portal binary
	@printf "Building portal... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA) -X main.release=$(RELEASE)" -o ${DIST_DIR}/portal ./cmd/portal/portal.go
	@printf "done\n"

.PHONY: build-portal-artifact
build-portal-artifact: build-portal ## builds the portal with the right env vars and creates a .tar.gz
	@printf "Building portal artifact... "
	@mkdir -p $(DIST_DIR)/artifact/portal
	@cp $(DIST_DIR)/portal $(DIST_DIR)/artifact/portal/app
	@cp -r ./cmd/portal/public $(DIST_DIR)/artifact/portal
	@cp ./cmd/portal/dev.env $(DIST_DIR)/artifact/portal/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/portal/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/portal && tar -zcf ../../portal.dev.tar.gz public app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/portal.dev.tar.gz\n"

.PHONY: build-portal-artifact
build-portal-prod-artifact: build-portal ## builds the portal with the right env vars and creates a .tar.gz
	@printf "Building portal artifact... "
	@mkdir -p $(DIST_DIR)/artifact/portal
	@cp $(DIST_DIR)/portal $(DIST_DIR)/artifact/portal/app
	@cp -r ./cmd/portal/public $(DIST_DIR)/artifact/portal
	@cp ./cmd/portal/prod.env $(DIST_DIR)/artifact/portal/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/portal/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/portal && tar -zcf ../../portal.prod.tar.gz public app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/portal.prod.tar.gz\n"

.PHONY: publish-portal-artifact
publish-portal-artifact: ## publishes the portal artifact to GCP Storage with gsutil
	@printf "Publishing portal artifact... \n\n"
	@gsutil cp $(DIST_DIR)/portal.dev.tar.gz $(ARTIFACT_BUCKET)/portal.dev.tar.gz
	@printf "done\n"

.PHONY: publish-portal-prod-artifact
publish-portal-prod-artifact: ## publishes the portal artifact to GCP Storage with gsutil
	@printf "Publishing portal artifact... \n\n"
	@gsutil cp $(DIST_DIR)/portal.prod.tar.gz $(ARTIFACT_BUCKET_PROD)/portal.prod.tar.gz
	@printf "done\n"

.PHONY: deploy-portal
deploy-portal: ## builds and deploys the portal to the dev VM
	@printf "Deploying portal... \n\n"
	gcloud compute --project "network-next-v3-dev" ssh portal-dev-1 -- 'cd /app && sudo ./vm-update-app.sh -a $(ARTIFACT_BUCKET)/portal.dev.tar.gz'

.PHONY: deploy-portal-prod
deploy-portal-prod: ## builds and deploys the portal to the prod instance group
	@printf "Deploying portal... \n\n"
	@./deploy/prod.sh portal-mig $(ARTIFACT_BUCKET_PROD)/portal.prod.tar.gz

.PHONY: build-relay-backend
build-relay-backend: ## builds the relay backend binary
	@printf "Building relay backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.commitsha=$(SHA) -X main.release=$(RELEASE)" -o ${DIST_DIR}/relay_backend ./cmd/relay_backend/relay_backend.go
	@printf "done\n"

.PHONY: build-relay-backend-artifact
build-relay-backend-artifact: build-relay-backend ## builds the relay backend with the right env vars and creates a .tar.gz
	@printf "Building relay backend artifact... "
	@mkdir -p $(DIST_DIR)/artifact/relay_backend
	@cp $(DIST_DIR)/relay_backend $(DIST_DIR)/artifact/relay_backend/app
	@cp ./cmd/relay_backend/dev.env $(DIST_DIR)/artifact/relay_backend/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/relay_backend/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/relay_backend && tar -zcf ../../relay_backend.dev.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/relay_backend.dev.tar.gz\n"

.PHONY: build-relay-backend-artifact
build-relay-backend-prod-artifact: build-relay-backend ## builds the relay backend with the right env vars and creates a .tar.gz
	@printf "Building relay backend artifact... "
	@mkdir -p $(DIST_DIR)/artifact/relay_backend
	@cp $(DIST_DIR)/relay_backend $(DIST_DIR)/artifact/relay_backend/app
	@cp ./cmd/relay_backend/prod.env $(DIST_DIR)/artifact/relay_backend/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/relay_backend/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/relay_backend && tar -zcf ../../relay_backend.prod.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/relay_backend.prod.tar.gz\n"

.PHONY: publish-relay-backend-artifact
publish-relay-backend-artifact: ## publishes the relay backend artifact to GCP Storage with gsutil
	@printf "Publishing relay backend artifact... \n\n"
	@gsutil cp $(DIST_DIR)/relay_backend.dev.tar.gz $(ARTIFACT_BUCKET)/relay_backend.dev.tar.gz
	@printf "done\n"

.PHONY: publish-relay-backend-artifact
publish-relay-backend-prod-artifact: ## publishes the relay backend artifact to GCP Storage with gsutil
	@printf "Publishing relay backend artifact... \n\n"
	@gsutil cp $(DIST_DIR)/relay_backend.prod.tar.gz $(ARTIFACT_BUCKET_PROD)/relay_backend.prod.tar.gz
	@printf "done\n"

.PHONY: deploy-relay-backend
deploy-relay-backend: ## builds and deploys the relay backend to the dev VM
	@printf "Deploying relay backend... \n\n"
	gcloud compute --project "network-next-v3-dev" ssh relay-backend-dev-1 -- 'cd /app && sudo ./vm-update-app.sh -a $(ARTIFACT_BUCKET)/relay_backend.dev.tar.gz'

.PHONY: deploy-relay-backend-prod
deploy-relay-backend-prod: ## builds and deploys the relay backend to the prod instance group
	@printf "Deploying relay backend... \n\n"
	@./deploy/prod.sh relay-backend-mig $(ARTIFACT_BUCKET_PROD)/relay_backend.prod.tar.gz

.PHONY: build-server-backend
build-server-backend: ## builds the server backend binary
	@printf "Building server backend... "
	@$(GO) build -ldflags "-s -w -X main.release=$(RELEASE)" -o ${DIST_DIR}/server_backend ./cmd/server_backend/server_backend.go
	@printf "done\n"

.PHONY: build-server-backend-artifact
build-server-backend-artifact: build-server-backend ## builds the server backend with the right env vars and creates a .tar.gz
	@printf "Building server backend artifact..."
	@mkdir -p $(DIST_DIR)/artifact/server_backend
	@cp $(DIST_DIR)/server_backend $(DIST_DIR)/artifact/server_backend/app
	@cp ./cmd/server_backend/dev.env $(DIST_DIR)/artifact/server_backend/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/server_backend/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/server_backend && tar -zcf ../../server_backend.dev.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/server_backend.dev.tar.gz\n"

.PHONY: build-server-backend-artifact
build-server-backend-prod-artifact: build-server-backend ## builds the server backend with the right env vars and creates a .tar.gz
	@printf "Building server backend artifact... "
	@mkdir -p $(DIST_DIR)/artifact/server_backend
	@cp $(DIST_DIR)/server_backend $(DIST_DIR)/artifact/server_backend/app
	@cp ./cmd/server_backend/prod.env $(DIST_DIR)/artifact/server_backend/app.env
	@cp $(DEPLOY_DIR)/$(SYSTEMD_SERVICE_FILE) $(DIST_DIR)/artifact/server_backend/$(SYSTEMD_SERVICE_FILE)
	@cd $(DIST_DIR)/artifact/server_backend && tar -zcf ../../server_backend.prod.tar.gz app app.env $(SYSTEMD_SERVICE_FILE) && cd ../..
	@printf "$(DIST_DIR)/server_backend.prod.tar.gz\n"

.PHONY: publish-server-backend-artifact
publish-server-backend-artifact: ## publishes the server backend artifact to GCP Storage with gsutil
	@printf "Publishing server backend artifact... \n\n"
	@gsutil cp $(DIST_DIR)/server_backend.dev.tar.gz $(ARTIFACT_BUCKET)/server_backend.dev.tar.gz
	@printf "done\n"

.PHONY: publish-server-backend-artifact
publish-server-backend-prod-artifact: ## publishes the server backend artifact to GCP Storage with gsutil
	@printf "Publishing server backend artifact... \n\n"
	@gsutil cp $(DIST_DIR)/server_backend.prod.tar.gz $(ARTIFACT_BUCKET_PROD)/server_backend.prod.tar.gz
	@printf "done\n"

.PHONY: deploy-server-backend
deploy-server-backend: ## builds and deploys the server backend to the dev VM
	@printf "Deploying server backend... \n\n"
	gcloud compute --project "network-next-v3-dev" ssh server-backend-dev-1 -- 'cd /app && sudo ./vm-update-app.sh -a $(ARTIFACT_BUCKET)/server_backend.dev.tar.gz'

.PHONY: deploy-server-backend-prod
deploy-server-backend-prod: ## builds and deploys the server backend to the prod instance group
	@printf "Deploying server backend... \n\n"
	@./deploy/prod.sh server-backend-mig $(ARTIFACT_BUCKET_PROD)/server_backend.prod.tar.gz

.PHONY: build-backend-artifacts
build-backend-artifacts: build-portal-artifact build-relay-backend-artifact build-server-backend-artifact ## builds the backend artifacts

.PHONY: build-relay-artifact
build-relay-artifact: build-relay-new ## builds the relay artifact
	@printf "Building relay artifact..."
	@mkdir -p $(DIST_DIR)/artifact/relay
	@cp $(DIST_DIR)/relay_new $(DIST_DIR)/artifact/relay/relay
	@cp $(DEPLOY_DIR)/relay/relay.service $(DIST_DIR)/artifact/relay/relay.service
	@cp $(DEPLOY_DIR)/relay/install.sh $(DIST_DIR)/artifact/relay/install.sh
	@cd $(DIST_DIR)/artifact/relay && tar -zcf ../../relay.dev.tar.gz relay relay.service install.sh && cd ../..
	@printf "$(DIST_DIR)/relay.dev.tar.gz\n"

.PHONY: publish-relay-artifact
publish-relay-artifact: ## publishes the relay artifact
	@printf "Publishing relay artifact... \n\n"
	@gsutil cp $(DIST_DIR)/relay.dev.tar.gz $(ARTIFACT_BUCKET)/relay.dev.tar.gz
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
	@$(CXX) -Isdk/include -o $(DIST_DIR)/func_server ./cmd/tools/functional/server/func_server.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-client
build-functional-client:
	@printf "Building functional client... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/func_client ./cmd/tools/functional/client/func_client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional
build-functional: build-functional-client build-functional-server

.PHONY: build-client
build-client: build-sdk ## builds the client
	@printf "Building client... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/client ./cmd/client/client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-all
build-all: build-relay-backend build-server-backend build-relay-ref build-client build-server build-functional build-sdk-test build-soak-test build-tools ## builds everything

.PHONY: rebuild-all
rebuild-all: clean build-all

.PHONY: update-sdk
update-sdk: ## updates the sdk submodule to point at head revision
	git submodule update --remote --merge

.PHONY: sync
sync: ## sync latest code including the sdk submodule
	git pull && cd sdk && git pull
