CXX_FLAGS := -Wall -Wextra -std=c++17
GO = go
GOFMT = gofmt

OS := $(shell uname -s | tr A-Z a-z)
ifeq ($(OS),darwin)
	LDFLAGS = -lsodium -lcurl -lpthread -lm -framework CoreFoundation -framework SystemConfiguration -DNEXT_DEVELOPMENT
	CXX = g++
else
	LDFLAGS = -lsodium -lcurl -lpthread -lm -DNEXT_DEVELOPMENT
	CXX = g++-8
endif

SDK3NAME = libnext3
SDK4NAME = libnext4

TIMESTAMP ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA ?= $(shell git rev-parse --short HEAD)
RELEASE ?= $(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE ?= $(shell git log -1 --pretty=%B | tr '\n' ' ')

CURRENT_DIR = $(shell pwd -P)
DEPLOY_DIR = ./deploy
DIST_DIR = ./dist
ARTIFACT_BUCKET = gs://development_artifacts
ARTIFACT_BUCKET_STAGING = gs://staging_artifacts
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

ifndef REDIS_HOST_TOP_SESSIONS
export REDIS_HOST_TOP_SESSIONS = 127.0.0.1:6379
endif

ifndef REDIS_HOST_SESSION_META
export REDIS_HOST_SESSION_META = 127.0.0.1:6379
endif

ifndef REDIS_HOST_SESSION_SLICES
export REDIS_HOST_SESSION_SLICES = 127.0.0.1:6379
endif

ifndef REDIS_HOST_SESSION_MAP
export REDIS_HOST_SESSION_MAP = 127.0.0.1:6379
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
export GOOGLE_FIRESTORE_SYNC_INTERVAL = 10s
endif

ifndef PORTAL_CRUNCHER_HOST
export PORTAL_CRUNCHER_HOST = tcp://127.0.0.1:5555
endif

ifndef ALLOWED_ORIGINS
export ALLOWED_ORIGINS = http://127.0.0.1:8080
endif

ifndef BILLING_CLIENT_COUNT
export BILLING_CLIENT_COUNT = 1
endif

ifndef BILLING_BATCHED_MESSAGE_COUNT
export BILLING_BATCHED_MESSAGE_COUNT = 20
endif

ifndef BILLING_BATCHED_MESSAGE_MIN_BYTES
export BILLING_BATCHED_MESSAGE_MIN_BYTES = 1024
endif

ifndef BILLING_ENTRY_VETO
export BILLING_ENTRY_VETO = true
endif

ifndef USE_THREAD_POOL
export USE_THREAD_POOL = true
endif

ifndef POST_SESSION_THREAD_COUNT
export POST_SESSION_THREAD_COUNT = 100
endif

ifndef POST_SESSION_BUFFER_SIZE
export POST_SESSION_BUFFER_SIZE = 100
endif

ifndef RELAY_STATS_URI
export RELAY_STATS_URI = $(RELAY_BACKEND_HOSTNAME)/relay_stats
endif

ifndef POST_SESSION_PORTAL_MAX_RETRIES
export POST_SESSION_PORTAL_MAX_RETRIES = 10
endif

ifndef POST_SESSION_PORTAL_SEND_BUFFER_SIZE
export POST_SESSION_PORTAL_SEND_BUFFER_SIZE = 100
endif

ifndef CRUNCHER_RECEIVE_BUFFER_SIZE
export CRUNCHER_RECEIVE_BUFFER_SIZE = 100
endif

ifndef GHOST_ARMY_BIN
export GHOST_ARMY_BIN = ./dist/ghost_army.bin
endif

ifndef DATACENTERS_CSV
export DATACENTERS_CSV = ./dist/datacenters.csv
endif

.PHONY: help
help:
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\033[36m\1\\033[m:\2/' | column -c2 -t -s :)"

.PHONY: dist
dist:
	mkdir -p $(DIST_DIR)

#####################
## ESSENTIAL TOOLS ##
#####################

.PHONY: test
test: clean ## runs unit tests
	@./scripts/test-unit-backend.sh

.PHONY: build-analytics
build-analytics: dist
	@printf "Building analytics... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/analytics ./cmd/analytics/analytics.go
	@printf "done\n"

ifeq ($(OS),darwin)
.PHONY: build-load-test-server
build-load-test-server: dist build-sdk3
	@printf "Building load test server... "
	@$(CXX) -Isdk3/include -o $(DIST_DIR)/load_test_server ./cmd/load_test_server/load_test_server.cpp  $(DIST_DIR)/$(SDK3NAME).so $(LDFLAGS)
	@printf "done\n"
else
.PHONY: build-load-test-server
build-load-test-server: dist build-sdk3
	@printf "Building load test server... "
	@$(CXX) -Isdk3/include -o $(DIST_DIR)/load_test_server ./cmd/load_test_server/load_test_server.cpp -L./dist -lnext3 $(LDFLAGS)
	@printf "done\n"
endif

ifeq ($(OS),darwin)
.PHONY: build-load-test-client
build-load-test-client: dist build-sdk3
	@printf "Building load test client... "
	@$(CXX) -Isdk3/include -o $(DIST_DIR)/load_test_client ./cmd/load_test_client/load_test_client.cpp  $(DIST_DIR)/$(SDK3NAME).so $(LDFLAGS)
	@printf "done\n"
else
.PHONY: build-load-test-client
build-load-test-client: dist build-sdk3
	@printf "Building load test client... "
	@$(CXX) -Isdk3/include -o $(DIST_DIR)/load_test_client ./cmd/load_test_client/load_test_client.cpp -L./dist -lnext3 $(LDFLAGS)
	@printf "done\n"
endif

.PHONY: build-functional-backend
build-functional-backend: dist
	@printf "Building functional backend... " ; \
	$(GO) build -o ./dist/func_backend ./cmd/func_backend/*.go ; \
	printf "done\n" ; \

.PHONY: build-functional-tests
build-functional-tests: dist
	@printf "Building functional tests... " ; \
	$(GO) build -o ./dist/func_tests ./cmd/func_tests/*.go ; \
	printf "done\n" ; \

.PHONY: build-test-func
build-test-func: clean dist build-sdk3 build-sdk4 build-relay-ref build-functional-server build-functional-client build-functional-backend build-functional-tests

.PHONY: run-test-func
run-test-func:
	@printf "\nRunning functional tests...\n\n" ; \
	$(GO) run ./cmd/func_tests/func_tests.go $(tests) ; \
	printf "\ndone\n\n"

.PHONY: test-func
test-func: build-test-func run-test-func ## runs functional tests

.PHONY: build-test-func-parallel
build-test-func-parallel: dist
	@docker build -t func_tests -f ./cmd/func_tests/Dockerfile .

.PHONY: run-test-func-parallel
run-test-func-parallel:
	@./scripts/test-func-parallel.sh

.PHONY: test-func-parallel
test-func-parallel: dist build-test-func-parallel run-test-func-parallel ## runs functional tests in parallel

.PHONY: test-load
test-load: ## runs load tests
	@printf "\nRunning load tests...\n" ; \
	$(GO) run ./cmd/load_test/load_tests.go

#######################

.PHONY: dev-portal
dev-portal: build-portal ## runs a local portal
	@PORT=20000 BASIC_AUTH_USERNAME=local BASIC_AUTH_PASSWORD=local UI_DIR=./cmd/portal/public ./dist/portal

.PHONY: dev-relay-backend
dev-relay-backend: build-relay-backend ## runs a local relay backend
	@PORT=30000 ./dist/relay_backend

.PHONY: dev-server-backend
dev-server-backend: build-server-backend ## runs a local server backend
	@HTTP_PORT=40000 UDP_PORT=40000 ./dist/server_backend

.PHONY: dev-server-backend4
dev-server-backend4: build-server-backend4 ## runs a local server backend4
	@HTTP_PORT=40000 UDP_PORT=40000 ./dist/server_backend4

.PHONY: dev-server-backend-valve
dev-server-backend-valve: build-server-backend
	@HTTP_PORT=40001 UDP_PORT=40001 ROUTE_MATRIX_URI=http://127.0.0.1:30000/route_matrix_valve ./dist/server_backend

.PHONY: dev-billing
dev-billing: build-billing ## runs a local billing service
	@PORT=41000 ./dist/billing

.PHONY: dev-analytics
dev-analytics: build-analytics ## runs a local analytics service
	@PORT=41001 ./dist/analytics

.PHONY: dev-portal-cruncher
dev-portal-cruncher: build-portal-cruncher ## runs a local portal cruncher
	@HTTP_PORT=42000 CRUNCHER_PORT=5555 ./dist/portal_cruncher

.PHONY: dev-ghost-army
dev-ghost-army: build-ghost-army ## runs a local ghost army
	@./dist/ghost_army

.PHONY: dev-reference-backend3
dev-reference-backend3: ## runs a local reference backend (sdk3)
	$(GO) run reference/backend3/backend3.go

.PHONY: dev-reference-backend4
dev-reference-backend4: ## runs a local reference backend (sdk4)
	$(GO) run reference/backend4/backend4.go

.PHONY: dev-reference-relay
dev-reference-relay: build-relay-ref ## runs a local reference relay
	@$(DIST_DIR)/reference_relay

.PHONY: dev-client3
dev-client3: build-client3  ## runs a local client (sdk3)
	@./dist/client3

.PHONY: dev-client4
dev-client4: build-client4  ## runs a local client (sdk4)
	@./dist/client4

.PHONY: dev-multi-clients3
dev-multi-clients3: build-client3  ## runs 10 local clients (sdk3)
	@./scripts/client-spawner.sh -n 10 -v 3

.PHONY: dev-multi-clients4
dev-multi-clients4: build-client4  ## runs 10 local clients (sdk4)
	@./scripts/client-spawner.sh -n 10 -v 4

.PHONY: dev-server3
dev-server3: build-sdk3 build-server3  ## runs a local server (sdk3)
	@./dist/server3

.PHONY: dev-server4
dev-server4: build-sdk4 build-server4  ## runs a local server (sdk4)
	@./dist/server4

$(DIST_DIR)/$(SDK3NAME).so: dist
	@printf "Building sdk3... "
	@$(CXX) -fPIC -Isdk3/include -shared -o $(DIST_DIR)/$(SDK3NAME).so ./sdk3/source/next.cpp ./sdk3/source/next_ios.cpp ./sdk3/source/next_linux.cpp ./sdk3/source/next_mac.cpp ./sdk3/source/next_ps4.cpp ./sdk3/source/next_switch.cpp ./sdk3/source/next_windows.cpp ./sdk3/source/next_xboxone.cpp $(LDFLAGS)
	@printf "done\n"

$(DIST_DIR)/$(SDK4NAME).so: dist
	@printf "Building sdk4... "
	@$(CXX) -fPIC -Isdk4/include -shared -o $(DIST_DIR)/$(SDK4NAME).so ./sdk4/source/next.cpp ./sdk4/source/next_ios.cpp ./sdk4/source/next_linux.cpp ./sdk4/source/next_mac.cpp ./sdk4/source/next_ps4.cpp ./sdk4/source/next_switch.cpp ./sdk4/source/next_windows.cpp ./sdk4/source/next_xboxone.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-sdk3
build-sdk3: $(DIST_DIR)/$(SDK3NAME).so

.PHONY: build-sdk4
build-sdk4: $(DIST_DIR)/$(SDK4NAME).so

PHONY: build-load-test
build-load-test:
	@printf "Building load test... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/load-test ./cmd/load_test/load_tests.go
	@printf "done\n"

PHONY: build-portal-cruncher
build-portal-cruncher:
	@printf "Building portal_cruncher... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/portal_cruncher ./cmd/portal_cruncher/portal_cruncher.go
	@printf "done\n"

.PHONY: build-portal
build-portal:
	@printf "Building portal... \n"
	@printf "TIMESTAMP: ${TIMESTAMP}\n"
	@printf "SHA: ${SHA}\n"
	@printf "RELEASE: ${RELEASE}\n"
	@printf "COMMITMESSAGE: ${COMMITMESSAGE}\n"
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/portal ./cmd/portal/portal.go
	@printf "done\n"

.PHONY: build-relay-backend
build-relay-backend:
	@printf "Building relay backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/relay_backend ./cmd/relay_backend/relay_backend.go
	@printf "done\n"

.PHONY: build-server-backend
build-server-backend:
	@printf "Building server backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/server_backend ./cmd/server_backend/server_backend.go
	@printf "done\n"

.PHONY: build-server-backend4
build-server-backend4:
	@printf "Building server_backend4... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/server_backend4 ./cmd/server_backend4/server_backend4.go
	@printf "done\n"

.PHONY: build-billing
build-billing:
	@printf "Building billing... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/billing ./cmd/billing/billing.go
	@printf "done\n"

.PHONY: deploy-relay-backend-dev
deploy-relay-backend-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t relay-backend -n relay_backend -b gs://development_artifacts

.PHONY: deploy-portal-cruncher-dev
deploy-portal-cruncher-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t portal-cruncher -n portal_cruncher -b gs://development_artifacts

.PHONY: deploy-relay-backend-staging
deploy-relay-backend-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t relay-backend -n relay_backend -b gs://staging_artifacts

.PHONY: deploy-portal-cruncher-staging
deploy-portal-cruncher-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t portal-cruncher -n portal_cruncher -b gs://staging_artifacts

.PHONY: deploy-relay-backend-prod
deploy-relay-backend-prod:
	./deploy/deploy.sh -e prod -c mig-jcr6 -t relay-backend -n relay_backend -b gs://prod_artifacts

.PHONY: deploy-portal-cruncher-prod
deploy-portal-cruncher-prod:
	./deploy/deploy.sh -e prod -c mig-07q1 -t portal-cruncher -n portal_cruncher -b gs://prod_artifacts

.PHONY: deploy-server-backend-dev-1
deploy-server-backend-dev-1:
	./deploy/deploy.sh -e dev -c dev-1 -t server-backend -n server_backend -b gs://development_artifacts

.PHONY: deploy-server-backend-dev-2
deploy-server-backend-dev-2:
	./deploy/deploy.sh -e dev -c dev-2 -t server-backend -n server_backend -b gs://development_artifacts

.PHONY: deploy-server-backend-staging
deploy-server-backend-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t server-backend -n server_backend -b gs://staging_artifacts

.PHONY: deploy-server-backend-psyonix
deploy-server-backend-psyonix:
	./deploy/deploy.sh -e prod -c psyonix -t server-backend -n server_backend -b gs://prod_artifacts

.PHONY: deploy-server-backend-liquidbit
deploy-server-backend-liquidbit:
	./deploy/deploy.sh -e prod -c prod-42rz -t server-backend -n server_backend -b gs://prod_artifacts

.PHONY: deploy-server-backend-valve
deploy-server-backend-valve:
	./deploy/deploy.sh -e prod -c valve-r57d -t server-backend -n server_backend -b gs://prod_artifacts

.PHONY: deploy-server-backend-velan
deploy-server-backend-velan:
	./deploy/deploy.sh -e prod -c velan-730n -t server-backend -n server_backend -b gs://prod_artifacts

.PHONY: deploy-server-backend-esl
deploy-server-backend-esl:
	./deploy/deploy.sh -e prod -c esl-22dr -t server-backend -n server_backend -b gs://prod_artifacts

.PHONY: deploy-server-backend4-dev
deploy-server-backend4-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t server-backend4 -n server_backend4 -b gs://development_artifacts

.PHONY: deploy-ghost-army-dev
deploy-ghost-army-dev:
	./deploy/deploy.sh -e dev -c 1 -t ghost-army -n ghost_army -b gs://development_artifacts

.PHONY: deploy-ghost-army-staging
deploy-ghost-army-staging:
	./deploy/deploy.sh -e staging -c 1 -t ghost-army -n ghost_army -b gs://staging_artifacts

.PHONY: deploy-ghost-army-prod
deploy-ghost-army-prod:
	./deploy/deploy.sh -e prod -c 1 -t ghost-army -n ghost_army -b gs://prod_artifacts

.PHONY: build-load-test-server-artifacts
build-load-test-server-artifacts: build-load-test-server
	./deploy/build-load-test-artifacts.sh -s load_test_server

.PHONY: build-load-test-client-artifacts
build-load-test-client-artifacts: build-load-test-client
	./deploy/build-load-test-artifacts.sh -s load_test_client

.PHONY: build-billing-artifacts-dev
build-billing-artifacts-dev: build-billing
	./deploy/build-artifacts.sh -e dev -s billing

.PHONY: build-analytics-artifacts-dev
build-analytics-artifacts-dev: build-analytics
	./deploy/build-artifacts.sh -e dev -s analytics

.PHONY: build-relay-artifacts-dev
build-relay-artifacts-dev: build-relay
	./deploy/build-artifacts.sh -e dev -s relay

.PHONY: build-portal-artifacts-dev
build-portal-artifacts-dev: build-portal
	./deploy/build-artifacts.sh -e dev -s portal

.PHONY: build-portal-cruncher-artifacts-dev
build-portal-cruncher-artifacts-dev: build-portal-cruncher
	./deploy/build-artifacts.sh -e dev -s portal_cruncher

.PHONY: build-relay-backend-artifacts-dev
build-relay-backend-artifacts-dev: build-relay-backend
	./deploy/build-artifacts.sh -e dev -s relay_backend

.PHONY: build-server-backend-artifacts-dev
build-server-backend-artifacts-dev: build-server-backend
	./deploy/build-artifacts.sh -e dev -s server_backend

.PHONY: build-server-backend4-artifacts-dev
build-server-backend4-artifacts-dev: build-server-backend4
	./deploy/build-artifacts.sh -e dev -s server_backend4

.PHONY: build-ghost-army-artifacts-dev
build-ghost-army-artifacts-dev: build-ghost-army
	./deploy/build-artifacts.sh -e dev -s ghost_army

.PHONY: build-billing-artifacts-staging
build-billing-artifacts-staging: build-billing
	./deploy/build-artifacts.sh -e staging -s billing

.PHONY: build-analytics-artifacts-staging
build-analytics-artifacts-staging: build-analytics
	./deploy/build-artifacts.sh -e staging -s analytics

.PHONY: build-relay-artifacts-staging
build-relay-artifacts-staging: build-relay
	./deploy/build-artifacts.sh -e staging -s relay

.PHONY: build-portal-artifacts-staging
build-portal-artifacts-staging: build-portal
	./deploy/build-artifacts.sh -e staging -s portal

.PHONY: build-relay-backend-artifacts-staging
build-relay-backend-artifacts-staging: build-relay-backend
	./deploy/build-artifacts.sh -e staging -s relay_backend

.PHONY: build-portal-cruncher-artifacts-staging
build-portal-cruncher-artifacts-staging: build-portal-cruncher
	./deploy/build-artifacts.sh -e staging -s portal_cruncher

.PHONY: build-load-test-artifacts-staging
build-load-test-artifacts-staging: build-load-test
	./deploy/build-artifacts.sh -e staging -s load-test

.PHONY: build-server-backend-artifacts-staging
build-server-backend-artifacts-staging: build-server-backend
	./deploy/build-artifacts.sh -e staging -s server_backend

.PHONY: build-server-backend4-artifacts-staging
build-server-backend4-artifacts-staging: build-server-backend4
	./deploy/build-artifacts.sh -e staging -s server_backend4

.PHONY: build-ghost-army-artifacts-staging
build-ghost-army-artifacts-staging: build-ghost-army
	./deploy/build-artifacts.sh -e staging -s ghost_army

.PHONY: build-billing-artifacts-prod
build-billing-artifacts-prod: build-billing
	./deploy/build-artifacts.sh -e prod -s billing

.PHONY: build-analytics-artifacts-prod
build-analytics-artifacts-prod: build-analytics
	./deploy/build-artifacts.sh -e prod -s analytics

.PHONY: build-relay-artifacts-prod
build-relay-artifacts-prod: build-relay
	./deploy/build-artifacts.sh -e prod -s relay

.PHONY: build-portal-artifacts-prod
build-portal-artifacts-prod: build-portal
	./deploy/build-artifacts.sh -e prod -s portal

.PHONY: build-portal-cruncher-artifacts-prod
build-portal-cruncher-artifacts-prod: build-portal-cruncher
	./deploy/build-artifacts.sh -e prod -s portal_cruncher

.PHONY: build-relay-backend-artifacts-prod
build-relay-backend-artifacts-prod: build-relay-backend
	./deploy/build-artifacts.sh -e prod -s relay_backend

.PHONY: build-server-backend-artifacts-prod
build-server-backend-artifacts-prod: build-server-backend
	./deploy/build-artifacts.sh -e prod -s server_backend

.PHONY: build-server-backend4-artifacts-prod
build-server-backend4-artifacts-prod: build-server-backend4
	./deploy/build-artifacts.sh -e prod -s server_backend4

.PHONY: build-ghost-army-artifacts-prod
build-ghost-army-artifacts-prod: build-ghost-army
	./deploy/build-artifacts.sh -e prod -s ghost_army

.PHONY: publish-billing-artifacts-dev
publish-billing-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s billing

.PHONY: publish-analytics-artifacts-dev
publish-analytics-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s analytics

.PHONY: publish-relay-artifacts-dev
publish-relay-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s relay

.PHONY: publish-portal-artifacts-dev
publish-portal-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s portal

.PHONY: publish-portal-cruncher-artifacts-dev
publish-portal-cruncher-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s portal_cruncher

.PHONY: publish-relay-backend-artifacts-dev
publish-relay-backend-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s relay_backend

.PHONY: publish-server-backend-artifacts-dev
publish-server-backend-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s server_backend

.PHONY: publish-server-backend4-artifacts-dev
publish-server-backend4-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s server_backend4

.PHONY: publish-ghost-army-artifacts-dev
publish-ghost-army-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s ghost_army

.PHONY: publish-billing-artifacts-staging
publish-billing-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s billing

.PHONY: publish-analytics-artifacts-staging
publish-analytics-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s analytics

.PHONY: publish-relay-artifacts-staging
publish-relay-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s relay

.PHONY: publish-portal-artifacts-staging
publish-portal-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s portal

.PHONY: publish-portal-cruncher-artifacts-staging
publish-portal-cruncher-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s portal_cruncher

.PHONY: publish-load-test-artifacts-staging
publish-load-test-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s load-test

.PHONY: publish-relay-backend-artifacts-staging
publish-relay-backend-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s relay_backend

.PHONY: publish-server-backend-artifacts-staging
publish-server-backend-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s server_backend

.PHONY: publish-server-backend4-artifacts-staging
publish-server-backend4-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s server_backend4

.PHONY: publish-ghost-army-artifacts-staging
publish-ghost-army-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s ghost_army

.PHONY: publish-load-test-server-artifacts
publish-load-test-server-artifacts:
	./deploy/publish-load-test-artifacts.sh -b $(ARTIFACT_BUCKET_STAGING) -s load_test_server

.PHONY: publish-load-test-client-artifacts
publish-load-test-client-artifacts:
	./deploy/publish-load-test-artifacts.sh -b $(ARTIFACT_BUCKET_STAGING) -s load_test_client

.PHONY: publish-load-test-server-list
publish-load-test-server-list:
	./deploy/publish-load-test-artifacts.sh -b $(ARTIFACT_BUCKET_STAGING) -s staging_servers.txt

.PHONY: publish-billing-artifacts-prod
publish-billing-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s billing

.PHONY: publish-analytics-artifacts-prod
publish-analytics-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s analytics

.PHONY: publish-relay-artifacts-prod
publish-relay-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s relay

.PHONY: publish-portal-artifacts-prod
publish-portal-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s portal

.PHONY: publish-portal-cruncher-artifacts-prod
publish-portal-cruncher-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s portal_cruncher

.PHONY: publish-relay-backend-artifacts-prod
publish-relay-backend-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s relay_backend

.PHONY: publish-server-backend-artifacts-prod
publish-server-backend-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s server_backend

.PHONY: publish-server-backend4-artifacts-prod
publish-server-backend4-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s server_backend4

.PHONY: publish-ghost-army-artifacts-prod
publish-ghost-army-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s ghost_army

.PHONY: publish-bootstrap-script-dev
publish-bootstrap-script-dev:
	@printf "Publishing bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/bootstrap.sh $(ARTIFACT_BUCKET)/bootstrap.sh
	@printf "done\n"

.PHONY: publish-bootstrap-script-staging
publish-bootstrap-script-staging:
	@printf "Publishing bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/bootstrap.sh $(ARTIFACT_BUCKET_STAGING)/bootstrap.sh
	@printf "done\n"

.PHONY: publish-client-bootstrap-script-staging
publish-client-bootstrap-script-staging:
	@printf "Publishing client bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/client_bootstrap.sh $(ARTIFACT_BUCKET_STAGING)/client_bootstrap.sh
	@printf "done\n"

.PHONY: publish-bootstrap-script-prod
publish-bootstrap-script-prod:
	@printf "Publishing bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/bootstrap.sh $(ARTIFACT_BUCKET_PROD)/bootstrap.sh
	@printf "done\n"

.PHONY: publish-ghost-army-bootstrap-script-dev
publish-ghost-army-bootstrap-script-dev:
	@printf "Publishing ghost army bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/ghost_army_bootstrap.sh $(ARTIFACT_BUCKET)/ghost_army_bootstrap.sh
	@printf "done\n"

.PHONY: publish-ghost-army-bootstrap-script-staging
publish-ghost-army-bootstrap-script-staging:
	@printf "Publishing ghost army bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/ghost_army_bootstrap.sh $(ARTIFACT_BUCKET_STAGING)/ghost_army_bootstrap.sh
	@printf "done\n"

.PHONY: publish-ghost-army-bootstrap-script-prod
publish-ghost-army-bootstrap-script-prod:
	@printf "Publishing ghost army bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/ghost_army_bootstrap.sh $(ARTIFACT_BUCKET_PROD)/ghost_army_bootstrap.sh
	@printf "done\n"

.PHONY: build-server3
build-server3: build-sdk3
	@printf "Building server3... "
	@$(CXX) -Isdk3/include -o $(DIST_DIR)/server3 ./cmd/server/server.cpp $(DIST_DIR)/$(SDK3NAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-server4
build-server4: build-sdk4
	@printf "Building server4... "
	@$(CXX) -Isdk4/include -o $(DIST_DIR)/server4 ./cmd/server/server.cpp $(DIST_DIR)/$(SDK4NAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-server
build-functional-server: build-sdk3
	@printf "Building functional server... "
	@$(CXX) -Isdk3/include -o $(DIST_DIR)/func_server ./cmd/func_server/func_server.cpp $(DIST_DIR)/$(SDK3NAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-client
build-functional-client:
	@printf "Building functional client... "
	@$(CXX) -Isdk3/include -o $(DIST_DIR)/func_client ./cmd/func_client/func_client.cpp $(DIST_DIR)/$(SDK3NAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional
build-functional: build-functional-client build-functional-server build-functional-backend build-functional-tests

.PHONY: build-client3
build-client3: build-sdk3
	@printf "Building client3... "
	@$(CXX) -Isdk3/include -o $(DIST_DIR)/client3 ./cmd/client/client.cpp $(DIST_DIR)/$(SDK3NAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-client4
build-client4: build-sdk4
	@printf "Building client4... "
	@$(CXX) -Isdk4/include -o $(DIST_DIR)/client4 ./cmd/client/client.cpp $(DIST_DIR)/$(SDK4NAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-next
build-next:
	@printf "Building operator tool... "
	@$(GO) build -o ./dist/next ./cmd/next/*.go
	@printf "done\n"

.PHONY: build-ghost-army
build-ghost-army:
	@printf "Building ghost army... "
	@$(GO) build -o ./dist/ghost_army ./cmd/ghost_army/*.go
	@printf "done\n"

.PHONY: build-ghost-army-generator
build-ghost-army-generator:
	@printf "Building ghost army generator... "
	@$(GO) build -o ./dist/gag ./cmd/ghost_army_generator/*.go
	@printf "done\n"

#######################
# Relay Build Process #
#######################

RELAY_DIR := ./cmd/relay
RELAY_MAKEFILE := Makefile
RELAY_EXE := relay

.PHONY: build-relay-ref
build-relay-ref:
	@printf "Building reference relay... "
	@$(CXX) $(CXX_FLAGS) -o $(DIST_DIR)/reference_relay reference/relay/*.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-relay
build-relay:
	@printf "Building relay... "
	@mkdir -p $(DIST_DIR)
	@cd $(RELAY_DIR) && $(MAKE) release
	@cp cmd/relay/bin/relay $(DIST_DIR)
	@echo "done"

.PHONY: dev-relay
dev-relay: build-relay ## runs a local relay
	@$(DIST_DIR)/$(RELAY_EXE)

.PHONY: dev-multi-relays
dev-multi-relays: build-relay ## runs 10 local relays
	./scripts/relay-spawner.sh -n 20 -p 10000

#######################

.PHONY: format
format:
	@$(GOFMT) -s -w .
	@printf "\n"

.PHONY: build-all
build-all: build-sdk3 build-sdk4 build-load-test build-portal-cruncher build-analytics build-billing build-relay-backend build-server-backend build-relay-ref build-client3 build-client4 build-server3 build-server4 build-functional build-next ## builds everything

.PHONY: rebuild-all
rebuild-all: clean build-all ## rebuilds enerything

.PHONY: update-submodules
update-submodules:
	git submodule update --remote --merge

.PHONY: clean
clean: ## cleans everything
	@rm -fr $(DIST_DIR)
	@mkdir $(DIST_DIR)
