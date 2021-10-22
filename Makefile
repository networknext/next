CXX_FLAGS := -Wall -Wextra -std=c++17
GO = go
GOFMT = gofmt
TAR = tar

OS := $(shell uname -s | tr A-Z a-z)
ifeq ($(OS),darwin)
	LDFLAGS = -lsodium -lcurl -lpthread -lm -framework CoreFoundation -framework SystemConfiguration -DNEXT_DEVELOPMENT
	CXX = g++
else
	LDFLAGS = -lsodium -lcurl -lpthread -lm -DNEXT_DEVELOPMENT
	CXX = g++-8
endif

SDKNAME = libnext

TIMESTAMP ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA ?= $(shell git rev-parse --short HEAD)
RELEASE ?= $(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE ?= $(shell git log -1 --pretty=%B | tr '\n' ' ')

CURRENT_DIR = $(shell pwd -P)
DEPLOY_DIR = ./deploy
DIST_DIR = ./dist
PORTAL_DIR=./cmd/portal
ARTIFACT_BUCKET = gs://development_artifacts
ARTIFACT_BUCKET_STAGING = gs://staging_artifacts
ARTIFACT_BUCKET_PROD = gs://prod_artifacts
ARTIFACT_BUCKET_PROD_DEBUG = gs://prod_debug_artifacts
ARTIFACT_BUCKET_RELAY = gs://relay_artifacts
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
export NEXT_BEACON_ADDRESS = 127.0.0.1:35000
export NEXT_DEBUG_LOGS=1

####################
##    RELAY ENV   ##
####################

export RELAY_BINARY_NAME = relay-2.0.6

ifndef RELAY_BACKEND_HOSTNAME
export RELAY_BACKEND_HOSTNAME = http://127.0.0.1:30002
endif

ifndef RELAY_GATEWAY
export RELAY_GATEWAY = http://127.0.0.1:30000
endif

ifndef RELAY_FRONTEND
export RELAY_FRONTEND = http://127.0.0.1:30005
endif

ifndef RELAY_FORWARDER
export RELAY_FORWARDER =
endif

ifndef MONDAY_API_KEY
export MONDAY_API_KEY = eyJhbGciOiJIUzI1NiJ9.eyJ0aWQiOjExNDIwOTg2NCwidWlkIjoxMzkwNDcyNSwiaWFkIjoiMjAyMS0wNi0xOFQxNjoyODo0MS44ODRaIiwicGVyIjoibWU6d3JpdGUiLCJhY3RpZCI6NTAyNzE4MCwicmduIjoidXNlMSJ9.0lFdTkvvUL1qFyWSQmgIehQZ_9nlEgrDHwKUQ9qQL24
endif

ifndef RELAY_ADDRESS
#export RELAY_ADDRESS = 127.0.0.1
export RELAY_ADDRESS = 127.0.0.1:35000
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
# export ROUTE_MATRIX_URI = http://127.0.0.1:30000/route_matrix
export ROUTE_MATRIX_URI = http://127.0.0.1:30005/route_matrix
endif

ifndef ROUTE_MATRIX_SYNC_INTERVAL
export ROUTE_MATRIX_SYNC_INTERVAL = 1s
endif

ifndef COST_MATRIX_INTERVAL
export COST_MATRIX_INTERVAL = 1s
endif

ifndef MAXMIND_CITY_DB_FILE
export MAXMIND_CITY_DB_FILE = testdata/GeoIP2-City-Test.mmdb
endif

ifndef MAXMIND_ISP_DB_FILE
export MAXMIND_ISP_DB_FILE = testdata/GeoIP2-ISP-Test.mmdb
endif

ifndef LOCAL_RELAYS
export LOCAL_RELAYS = 10
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

ifndef RELAY_STORE_ADDRESS
export RELAY_STORE_ADDRESS = 127.0.0.1:6379
endif

ifndef AUTH0_DOMAIN
export AUTH0_DOMAIN = networknext-dev.us.auth0.com
endif
ifndef AUTH0_CLIENTID
export AUTH0_CLIENTID = 4j7UFJkp3x7rk5RudzxC6gToSes6dIn6
endif
ifndef AUTH0_CLIENTSECRET
export AUTH0_CLIENTSECRET = q5bMLiO8BoXcIy1CFe-sxy2eOYfn0IU0ByBZeeQkpckhV6_sQFR22EBDioyubwb6
endif
ifndef AUTH0_ISSUER
export AUTH0_ISSUER = https://auth-dev.networknext.com/
endif

ifndef GOOGLE_FIRESTORE_SYNC_INTERVAL
export GOOGLE_FIRESTORE_SYNC_INTERVAL = 10s
endif

ifndef GOOGLE_CLOUD_SQL_SYNC_INTERVAL
export GOOGLE_CLOUD_SQL_SYNC_INTERVAL = 10s
endif

ifndef PORTAL_CRUNCHER_HOSTS
export PORTAL_CRUNCHER_HOSTS = tcp://127.0.0.1:5555,tcp://127.0.0.1:5556
endif

ifndef ALLOWED_ORIGINS
export ALLOWED_ORIGINS = http://127.0.0.1:8080,http://127.0.0.1:8081
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

# Bigtable emulator must be running before testing bigtable in happy path
ifndef FEATURE_BIGTABLE
export FEATURE_BIGTABLE = false
endif

ifndef BIGTABLE_CF_NAME
export BIGTABLE_CF_NAME = portal-session-history
endif

ifndef BIGTABLE_TABLE_NAME
export BIGTABLE_TABLE_NAME = portal-session-history
endif

ifndef BIGTABLE_HISTORICAL_TXT
export BIGTABLE_HISTORICAL_TXT = ./testdata/bigtable_historical.txt
endif

ifndef FEATURE_VANITY_METRIC
export FEATURE_VANITY_METRIC = false
endif

ifndef BEACON_ENTRY_VETO
export BEACON_ENTRY_VETO = false
endif

## New Relay Backend

ifndef MATRIX_STORE_ADDRESS
export MATRIX_STORE_ADDRESS = 127.0.0.1:6379
endif

ifndef RELAY_BACKEND_ADDRESSES
export RELAY_BACKEND_ADDRESSES = 127.0.0.1:30001,127.0.0.1:30002
endif

.PHONY: help
help:
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\033[36m\1\\033[m:\2/' | column -c2 -t -s :)"

.PHONY: dist
dist:
	mkdir -p $(DIST_DIR)

#####################
##   Happy Path    ##
#####################

# Always run sqlite3
export FEATURE_POSTGRESQL=false
export JWT_AUDIENCE=S4WGLze2EZCPG9MeZ5509BedlWlHZGFt
export SLACK_WEBHOOK_URL=https://hooks.slack.com/services/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
export SLACK_CHANNEL=portal-test
export LOOKER_SECRET=d61764ff20f99e672af3ec7fde75531a790acdb6d58bf46dbe55dac06a6019c0
export GITHUB_ACCESS_TOKEN=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
export RELEASE_NOTES_INTERVAL=30s

.PHONY: dev-relay-gateway
dev-relay-gateway: build-relay-gateway ## runs a local relay gateway
	@PORT=30000 ./dist/relay_gateway

.PHONY: dev-relay-backend-1
dev-relay-backend-1: build-relay-backend ## runs a local relay backend
	@PORT=30001 ./dist/relay_backend

.PHONY: dev-relay-backend-2
dev-relay-backend-2: build-relay-backend ## runs a local relay backend
	@PORT=30002 ./dist/relay_backend

.PHONY: dev-debug-relay-backend
dev-debug-relay-backend: build-relay-backend ## runs a local debug relay backend
	@PORT=30003 ./dist/relay_backend

.PHONY: dev-relay-frontend
dev-relay-frontend: build-relay-frontend ## runs a local route matrix selector
	@PORT=30005 ./dist/relay_frontend

.PHONY: dev-server-backend
dev-server-backend: build-server-backend ## runs a local server backend
	@HTTP_PORT=40000 UDP_PORT=40000 ./dist/server_backend

.PHONY: dev-relay
dev-relay: build-reference-relay  ## runs a local relay
	@./scripts/relay-spawner.sh -n 1

.PHONY: dev-relays
dev-relays: build-reference-relay  ## runs 10 local relays
	@./scripts/relay-spawner.sh -n 10

.PHONY: dev-client
dev-client: build-client  ## runs a local client
	@./scripts/client-spawner.sh -n 1

.PHONY: dev-clients
dev-clients: build-client  ## runs 10 local clients
	@./scripts/client-spawner.sh -n 10

.PHONY: dev-server
dev-server: build-sdk build-server  ## runs a local server
	@./dist/server

.PHONY: dev-portal
dev-portal: build-portal ## runs a local portal
	@PORT=20000 BASIC_AUTH_USERNAME=local BASIC_AUTH_PASSWORD=local UI_DIR=./cmd/portal/dist RELAY_FRONTEND=http://localhost:30005 ./dist/portal

.PHONY: dev-beacon
dev-beacon: build-beacon ## runs a local beacon
	@HTTP_PORT=35000 UDP_PORT=35000 ./dist/beacon

.PHONY: dev-beacon-inserter
dev-beacon-inserter: build-beacon-inserter ## runs a local beacon inserter
	@PORT=35001 ./dist/beacon_inserter

.PHONY: dev-billing
dev-billing: build-billing ## runs a local billing service
	@PORT=41000 ./dist/billing

.PHONY: dev-analytics-pusher
dev-analytics-pusher: build-analytics-pusher ## runs a local analytics pusher service
	@PORT=41002 ./dist/analytics_pusher

.PHONY: dev-analytics
dev-analytics: build-analytics ## runs a local analytics service
	@PORT=41001 ./dist/analytics

.PHONY: dev-portal-cruncher-1
dev-portal-cruncher-1: build-portal-cruncher ## runs a local portal cruncher
	@HTTP_PORT=42000 CRUNCHER_PORT=5555 ./dist/portal_cruncher

.PHONY: dev-portal-cruncher-2
dev-portal-cruncher-2: build-portal-cruncher ## runs a local portal cruncher
	@HTTP_PORT=42001 CRUNCHER_PORT=5556 ./dist/portal_cruncher

.PHONY: dev-api
dev-api: build-api ## runs a local api endpoint service
	@PORT=41003 ENABLE_STACKDRIVER_METRICS=true ./dist/api

.PHONY: dev-vanity
dev-vanity: build-vanity ## runs insertion and updating of vanity metrics
	@HTTP_PORT=41005 FEATURE_VANITY_METRIC_PORT=6666 ./dist/vanity

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
build-load-test-server: dist build-sdk
	@printf "Building load test server... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/load_test_server ./cmd/load_test_server/load_test_server.cpp  $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"
else
.PHONY: build-load-test-server
build-load-test-server: dist build-sdk
	@printf "Building load test server... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/load_test_server ./cmd/load_test_server/load_test_server.cpp -L./dist -lnext $(LDFLAGS)
	@printf "done\n"
endif

ifeq ($(OS),darwin)
.PHONY: build-load-test-client
build-load-test-client: dist build-sdk
	@printf "Building load test client... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/load_test_client ./cmd/load_test_client/load_test_client.cpp  $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"
else
.PHONY: build-load-test-client
build-load-test-client: dist build-sdk
	@printf "Building load test client... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/load_test_client ./cmd/load_test_client/load_test_client.cpp -L./dist -lnext $(LDFLAGS)
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
build-test-func: clean dist build-sdk build-reference-relay build-functional-server build-functional-client build-functional-backend build-functional-tests

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

#######################

.PHONY: dev-reference-backend
dev-reference-backend: ## runs a local reference backend
	$(GO) run reference/backend/backend.go

.PHONY: dev-mock-relay
dev-mock-relay: ## runs a local mock relay
	$(GO) build -o ./dist/mock_relay ./cmd/mock_relay/mock_relay.go
	./dist/mock_relay

.PHONY: dev-fake-server
dev-fake-server: build-fake-server ## runs a fake server that simulates 2 servers and 400 clients locally
	@HTTP_PORT=50001 UDP_PORT=50000 ./dist/fake_server

$(DIST_DIR)/$(SDKNAME).so: dist
	@printf "Building sdk... "
	@$(CXX) -fPIC -Isdk/include -shared -o $(DIST_DIR)/$(SDKNAME).so ./sdk/source/*.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-sdk
build-sdk: $(DIST_DIR)/$(SDKNAME).so

PHONY: build-portal-cruncher
build-portal-cruncher:
	@printf "Building portal cruncher... "
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

.PHONY: build-beacon
build-beacon:
	@printf "Building beacon..."
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/beacon ./cmd/beacon/beacon.go
	@printf "done\n"

.PHONY: build-beacon-inserter
build-beacon-inserter:
	@printf "Building beacon inserter..."
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/beacon_inserter ./cmd/beacon_inserter/beacon_inserter.go
	@printf "done\n"

.PHONY: build-server-backend
build-server-backend:
	@printf "Building server backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/server_backend ./cmd/server_backend/server_backend.go
	@printf "done\n"

.PHONY: build-billing
build-billing:
	@printf "Building billing... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/billing ./cmd/billing/billing.go
	@printf "done\n"

.PHONY: build-analytics-pusher
build-analytics-pusher:
	@printf "Building analytics pusher... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/analytics_pusher ./cmd/analytics_pusher/analytics_pusher.go
	@printf "done\n"

.PHONY: build-api
build-api: dist
	@printf "Building api... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/api ./cmd/api/api.go
	@printf "done\n"

.PHONY: build-vanity
build-vanity: dist
	@printf "Building vanity metrics ..."
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/vanity ./cmd/vanity/vanity.go
	@printf "done\n"

.PHONY: build-fake-server
build-fake-server: dist
	@printf "Building fake server..."
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/fake_server ./cmd/fake_server/fake_server.go
	@printf "done\n"

.PHONY: deploy-portal-crunchers-dev
deploy-portal-crunchers-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t portal-cruncher -n portal_cruncher -b gs://development_artifacts
	./deploy/deploy.sh -e dev -c dev-2 -t portal-cruncher -n portal_cruncher -b gs://development_artifacts

.PHONY: deploy-vanity-dev
deploy-vanity-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t vanity -n vanity -b gs://development_artifacts
	./deploy/deploy.sh -e dev -c dev-2 -t vanity -n vanity -b gs://development_artifacts

.PHONY: deploy-analytics-pusher-dev
deploy-analytics-pusher-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t analytics-pusher -n analytics_pusher -b gs://development_artifacts

.PHONY: deploy-portal-crunchers-staging
deploy-portal-crunchers-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t portal-cruncher -n portal_cruncher -b gs://staging_artifacts
	./deploy/deploy.sh -e staging -c staging-2 -t portal-cruncher -n portal_cruncher -b gs://staging_artifacts
	./deploy/deploy.sh -e staging -c staging-3 -t portal-cruncher -n portal_cruncher -b gs://staging_artifacts
	./deploy/deploy.sh -e staging -c staging-4 -t portal-cruncher -n portal_cruncher -b gs://staging_artifacts

.PHONY: deploy-vanity-staging
deploy-vanity-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t vanity -n vanity -b gs://staging_artifacts
	./deploy/deploy.sh -e staging -c staging-2 -t vanity -n vanity -b gs://staging_artifacts
	./deploy/deploy.sh -e staging -c staging-3 -t vanity -n vanity -b gs://staging_artifacts
	./deploy/deploy.sh -e staging -c staging-4 -t vanity -n vanity -b gs://staging_artifacts

.PHONY: deploy-analytics-pusher-staging
deploy-analytics-pusher-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t analytics-pusher -n analytics_pusher -b gs://staging_artifacts

.PHONY: deploy-portal-crunchers-prod
deploy-portal-crunchers-prod:
	./deploy/deploy.sh -e prod -c prod-1 -t portal-cruncher -n portal_cruncher -b gs://prod_artifacts
	./deploy/deploy.sh -e prod -c prod-2 -t portal-cruncher -n portal_cruncher -b gs://prod_artifacts
	./deploy/deploy.sh -e prod -c prod-3 -t portal-cruncher -n portal_cruncher -b gs://prod_artifacts
	./deploy/deploy.sh -e prod -c prod-4 -t portal-cruncher -n portal_cruncher -b gs://prod_artifacts

.PHONY: deploy-vanity-prod
deploy-vanity-prod:
	./deploy/deploy.sh -e prod -c prod-1 -t vanity -n vanity -b gs://prod_artifacts
	./deploy/deploy.sh -e prod -c prod-2 -t vanity -n vanity -b gs://prod_artifacts
	./deploy/deploy.sh -e prod -c prod-3 -t vanity -n vanity -b gs://prod_artifacts
	./deploy/deploy.sh -e prod -c prod-4 -t vanity -n vanity -b gs://prod_artifacts

.PHONY: deploy-analytics-pusher-prod
deploy-analytics-pusher-prod:
	./deploy/deploy.sh -e prod -c prod-1 -t analytics-pusher -n analytics_pusher -b gs://prod_artifacts

.PHONY: build-fake-server-artifacts-staging
build-fake-server-artifacts-staging: build-fake-server
	./deploy/build-artifacts.sh -e staging -s fake_server

.PHONY: build-load-test-server-artifacts
build-load-test-server-artifacts: build-load-test-server
	./deploy/build-load-test-artifacts.sh -s load_test_server

.PHONY: build-load-test-client-artifacts
build-load-test-client-artifacts: build-load-test-client
	./deploy/build-load-test-artifacts.sh -s load_test_client

.PHONY: build-billing-artifacts-dev
build-billing-artifacts-dev: build-billing
	./deploy/build-artifacts.sh -e dev -s billing

.PHONY: build-beacon-artifacts-dev
build-beacon-artifacts-dev: build-beacon
	./deploy/build-artifacts.sh -e dev -s beacon

.PHONY: build-beacon-inserter-artifacts-dev
build-beacon-inserter-artifacts-dev: build-beacon-inserter
	./deploy/build-artifacts.sh -e dev -s beacon_inserter

.PHONY: build-analytics-pusher-artifacts-dev
build-analytics-pusher-artifacts-dev: build-analytics-pusher
	./deploy/build-artifacts.sh -e dev -s analytics_pusher

.PHONY: build-analytics-artifacts-dev
build-analytics-artifacts-dev: build-analytics
	./deploy/build-artifacts.sh -e dev -s analytics

.PHONY: build-api-artifacts-dev
build-api-artifacts-dev: build-api
	./deploy/build-artifacts.sh -e dev -s api

.PHONY: build-vanity-artifacts-dev
build-vanity-artifacts-dev: build-vanity
	./deploy/build-artifacts.sh -e dev -s vanity

.PHONY: build-relay-artifacts-dev
build-relay-artifacts-dev: build-relay
	./deploy/build-artifacts.sh -e dev -s relay

.PHONY: build-portal-artifacts-dev
build-portal-artifacts-dev: build-portal
	./deploy/build-artifacts.sh -e dev -s portal -b $(ARTIFACT_BUCKET)

.PHONY: build-portal-artifacts-dev-old
build-portal-artifacts-dev-old: build-portal
	./deploy/build-artifacts.sh -e dev -s portal-old -b $(ARTIFACT_BUCKET)

.PHONY: build-portal-cruncher-artifacts-dev
build-portal-cruncher-artifacts-dev: build-portal-cruncher
	./deploy/build-artifacts.sh -e dev -s portal_cruncher

.PHONY: build-server-backend-artifacts-dev
build-server-backend-artifacts-dev: build-server-backend
	./deploy/build-artifacts.sh -e dev -s server_backend

.PHONY: build-billing-artifacts-staging
build-billing-artifacts-staging: build-billing
	./deploy/build-artifacts.sh -e staging -s billing

.PHONY: build-beacon-artifacts-staging
build-beacon-artifacts-staging: build-beacon
	./deploy/build-artifacts.sh -e staging -s beacon

.PHONY: build-beacon-inserter-artifacts-staging
build-beacon-inserter-artifacts-staging: build-beacon-inserter
	./deploy/build-artifacts.sh -e staging -s beacon_inserter

.PHONY: build-analytics-pusher-artifacts-staging
build-analytics-pusher-artifacts-staging: build-analytics-pusher
	./deploy/build-artifacts.sh -e staging -s analytics_pusher

.PHONY: build-analytics-artifacts-staging
build-analytics-artifacts-staging: build-analytics
	./deploy/build-artifacts.sh -e staging -s analytics

.PHONY: build-api-artifacts-staging
build-api-artifacts-staging: build-api
	./deploy/build-artifacts.sh -e staging -s api

.PHONY: build-vanity-artifacts-staging
build-vanity-artifacts-staging: build-vanity
	./deploy/build-artifacts.sh -e staging -s vanity

.PHONY: build-relay-artifacts-staging
build-relay-artifacts-staging: build-relay
	./deploy/build-artifacts.sh -e staging -s relay

.PHONY: build-portal-artifacts-staging
build-portal-artifacts-staging: build-portal
	./deploy/build-artifacts.sh -e staging -s portal -b $(ARTIFACT_BUCKET_STAGING)

.PHONY: build-portal-cruncher-artifacts-staging
build-portal-cruncher-artifacts-staging: build-portal-cruncher
	./deploy/build-artifacts.sh -e staging -s portal_cruncher

.PHONY: build-server-backend-artifacts-staging
build-server-backend-artifacts-staging: build-server-backend
	./deploy/build-artifacts.sh -e staging -s server_backend

.PHONY: build-billing-artifacts-prod
build-billing-artifacts-prod: build-billing
	./deploy/build-artifacts.sh -e prod -s billing

.PHONY: build-debug-billing-artifacts-prod-debug
build-debug-billing-artifacts-prod-debug: build-billing
	./deploy/build-artifacts.sh -e prod -s debug_billing

.PHONY: build-beacon-artifacts-prod
build-beacon-artifacts-prod: build-beacon
	./deploy/build-artifacts.sh -e prod -s beacon

.PHONY: build-beacon-inserter-artifacts-prod
build-beacon-inserter-artifacts-prod: build-beacon-inserter
	./deploy/build-artifacts.sh -e prod -s beacon_inserter

.PHONY: build-analytics-pusher-artifacts-prod
build-analytics-pusher-artifacts-prod: build-analytics-pusher
	./deploy/build-artifacts.sh -e prod -s analytics_pusher

.PHONY: build-analytics-artifacts-prod
build-analytics-artifacts-prod: build-analytics
	./deploy/build-artifacts.sh -e prod -s analytics

.PHONY: build-api-artifacts-prod
build-api-artifacts-prod: build-api
	./deploy/build-artifacts.sh -e prod -s api

.PHONY: build-vanity-artifacts-prod
build-vanity-artifacts-prod: build-vanity
	./deploy/build-artifacts.sh -e prod -s vanity

.PHONY: build-relay-artifacts-prod
build-relay-artifacts-prod: build-relay
	./deploy/build-artifacts.sh -e prod -s relay

.PHONY: build-portal-artifacts-prod
build-portal-artifacts-prod: build-portal
	./deploy/build-artifacts.sh -e prod -s portal -b $(ARTIFACT_BUCKET_PROD)

.PHONY: build-portal-cruncher-artifacts-prod
build-portal-cruncher-artifacts-prod: build-portal-cruncher
	./deploy/build-artifacts.sh -e prod -s portal_cruncher

.PHONY: build-server-backend-artifacts-prod
build-server-backend-artifacts-prod: build-server-backend
	./deploy/build-artifacts.sh -e prod -s server_backend

.PHONY: publish-billing-artifacts-dev
publish-billing-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s billing

.PHONY: publish-beacon-artifacts-dev
publish-beacon-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s beacon

.PHONY: publish-beacon-inserter-artifacts-dev
publish-beacon-inserter-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s beacon_inserter

.PHONY: publish-analytics-pusher-artifacts-dev
publish-analytics-pusher-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s analytics_pusher

.PHONY: publish-analytics-artifacts-dev
publish-analytics-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s analytics

.PHONY: publish-api-artifacts-dev
publish-api-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s api

.PHONY: publish-vanity-artifacts-dev
publish-vanity-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s vanity

.PHONY: publish-relay-artifacts-dev
publish-relay-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s relay

.PHONY: publish-portal-artifacts-dev
publish-portal-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s portal

.PHONY: publish-portal-artifacts-dev-test
publish-portal-artifacts-dev-test:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s portal-test

.PHONY: publish-portal-cruncher-artifacts-dev
publish-portal-cruncher-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s portal_cruncher

.PHONY: publish-server-backend-artifacts-dev
publish-server-backend-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s server_backend

.PHONY: publish-billing-artifacts-staging
publish-billing-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s billing

.PHONY: publish-beacon-artifacts-staging
publish-beacon-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s beacon

.PHONY: publish-beacon-inserter-artifacts-staging
publish-beacon-inserter-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s beacon_inserter

.PHONY: publish-analytics-pusher-artifacts-staging
publish-analytics-pusher-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s analytics_pusher

.PHONY: publish-analytics-artifacts-staging
publish-analytics-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s analytics

.PHONY: publish-api-artifacts-staging
publish-api-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s api

.PHONY: publish-vanity-artifacts-staging
publish-vanity-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s vanity

.PHONY: publish-relay-artifacts-staging
publish-relay-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s relay

.PHONY: publish-portal-artifacts-staging
publish-portal-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s portal

.PHONY: publish-portal-cruncher-artifacts-staging
publish-portal-cruncher-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s portal_cruncher

.PHONY: publish-server-backend-artifacts-staging
publish-server-backend-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s server_backend

.PHONY: publish-fake-server-artifacts-staging
publish-fake-server-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s fake_server

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

.PHONY: publish-debug-billing-artifacts-prod-debug
publish-debug-billing-artifacts-prod-debug:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD_DEBUG) -s debug_billing

.PHONY: deploy-debug-billing-prod-billing
deploy-debug-billing-prod-debug:
	./deploy/deploy.sh -e prod -c prod-1 -t debug-billing -n debug_billing -b $(ARTIFACT_BUCKET_PROD_DEBUG)

.PHONY: publish-beacon-artifacts-prod
publish-beacon-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s beacon

.PHONY: publish-beacon-inserter-artifacts-prod
publish-beacon-inserter-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s beacon_inserter

.PHONY: publish-api-artifacts-prod
publish-api-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s api

.PHONY: publish-vanity-artifacts-prod
publish-vanity-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s vanity

.PHONY: publish-analytics-pusher-artifacts-prod
publish-analytics-pusher-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s analytics_pusher

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

.PHONY: publish-server-backend-artifacts-prod
publish-server-backend-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s server_backend

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

.PHONY: publish-bootstrap-script-prod
publish-bootstrap-script-prod:
	@printf "Publishing bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/bootstrap.sh $(ARTIFACT_BUCKET_PROD)/bootstrap.sh
	@printf "done\n"

.PHONY: publish-bootstrap-script-prod-debug
publish-bootstrap-script-prod-debug:
	@printf "Publishing bootstrap script... \n\n"
	@gsutil cp $(DEPLOY_DIR)/bootstrap.sh $(ARTIFACT_BUCKET_PROD_DEBUG)/bootstrap.sh
	@printf "done\n"

.PHONY: build-server
build-server: build-sdk
	@printf "Building server... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/server ./cmd/server/server.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-server
build-functional-server: build-sdk
	@printf "Building functional server... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/func_server ./cmd/func_server/func_server.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-client
build-functional-client: build-sdk
	@printf "Building functional client... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/func_client ./cmd/func_client/func_client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional
build-functional: build-functional-client build-functional-server build-functional-backend build-functional-tests

.PHONY: build-client
build-client: build-sdk
	@printf "Building client... "
	@$(CXX) -Isdk/include -o $(DIST_DIR)/client ./cmd/client/client.cpp $(DIST_DIR)/$(SDKNAME).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-next
build-next:
	@printf "Building operator tool... "
	@$(GO) build -o ./dist/next ./cmd/next/*.go
	@printf "done\n"

#######################
#    Relay Pusher    #
#######################

.PHONY: build-relay-pusher
build-relay-pusher:
	@printf "Building relay pusher... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE)" -o ${DIST_DIR}/relay_pusher ./cmd/relay_pusher/relay_pusher.go
	@printf "done\n"

.PHONY: build-relay-pusher-artifacts-dev
build-relay-pusher-artifacts-dev: build-relay-pusher
	./deploy/build-artifacts.sh -e dev -s relay_pusher

.PHONY: build-relay-pusher-artifacts-staging
build-relay-pusher-artifacts-staging: build-relay-pusher
	./deploy/build-artifacts.sh -e staging -s relay_pusher

.PHONY: build-relay-pusher-artifacts-prod
build-relay-pusher-artifacts-prod: build-relay-pusher
	./deploy/build-artifacts.sh -e prod -s relay_pusher

.PHONY: deploy-relay-pusher-dev
deploy-relay-pusher-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t relay-pusher -n relay_pusher -b gs://development_artifacts

.PHONY: deploy-relay-pusher-staging
deploy-relay-pusher-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t relay-pusher -n relay_pusher -b gs://staging_artifacts

.PHONY: deploy-relay-pusher-prod
deploy-relay-pusher-prod:
	./deploy/deploy.sh -e prod -c prod-1 -t relay-pusher -n relay_pusher -b gs://prod_artifacts

.PHONY: publish-relay-pusher-artifacts-dev
publish-relay-pusher-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s relay_pusher

.PHONY: publish-relay-pusher-artifacts-staging
publish-relay-pusher-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s relay_pusher

.PHONY: publish-relay-pusher-artifacts-prod
publish-relay-pusher-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s relay_pusher

#############################
#    Debug Server Backend   #
#############################

.PHONY: build-debug-server-backend-artifacts-dev
build-debug-server-backend-artifacts-dev: build-server-backend
	./deploy/build-artifacts.sh -e dev -s debug_server_backend

.PHONY: build-debug-server-backend-artifacts-staging
build-debug-server-backend-artifacts-staging: build-server-backend
	./deploy/build-artifacts.sh -e staging -s debug_server_backend

.PHONY: build-debug-server-backend-artifacts-prod
build-debug-server-backend-artifacts-prod: build-server-backend
	./deploy/build-artifacts.sh -e prod -s debug_server_backend

.PHONY: build-debug-server-backend-artifacts-prod-debug
build-debug-server-backend-artifacts-prod-debug: build-server-backend
	./deploy/build-artifacts.sh -e prod -s debug_server_backend

.PHONY: publish-debug-server-backend-artifacts-dev
publish-debug-server-backend-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s debug_server_backend

.PHONY: publish-debug-server-backend-artifacts-staging
publish-debug-server-backend-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s debug_server_backend

.PHONY: publish-debug-server-backend-artifacts-prod
publish-debug-server-backend-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s debug_server_backend

.PHONY: publish-debug-server-backend-artifacts-prod-debug
publish-debug-server-backend-artifacts-prod-debug:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD_DEBUG) -s debug_server_backend

.PHONY: deploy-debug-server-backend-dev
deploy-debug-server-backend-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t debug-server-backend -n debug_server_backend -b $(ARTIFACT_BUCKET)

.PHONY: deploy-debug-server-backend-staging
deploy-debug-server-backend-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t debug-server-backend -n debug_server_backend -b $(ARTIFACT_BUCKET_STAGING)

.PHONY: deploy-debug-server-backend-prod
deploy-debug-server-backend-prod:
	./deploy/deploy.sh -e prod -c prod-1 -t debug-server-backend -n debug_server_backend -b $(ARTIFACT_BUCKET_PROD)

.PHONY: deploy-debug-server-backend-prod-debug
deploy-debug-server-backend-prod-debug:
	./deploy/deploy.sh -e prod -c prod-1 -t debug-server-backend -n debug_server_backend -b $(ARTIFACT_BUCKET_PROD_DEBUG)

#############################
#    Debug Relay Backend   #
#############################

.PHONY: build-debug-relay-backend-artifacts-dev
build-debug-relay-backend-artifacts-dev: build-relay-backend
	./deploy/build-artifacts.sh -e dev -s debug_relay_backend

.PHONY: build-debug-relay-backend-artifacts-staging
build-debug-relay-backend-artifacts-staging: build-relay-backend
	./deploy/build-artifacts.sh -e staging -s debug_relay_backend

.PHONY: build-debug-relay-backend-artifacts-prod
build-debug-relay-backend-artifacts-prod: build-relay-backend
	./deploy/build-artifacts.sh -e prod -s debug_relay_backend

.PHONY: build-debug-relay-backend-artifacts-prod-debug
build-debug-relay-backend-artifacts-prod-debug: build-relay-backend
	./deploy/build-artifacts.sh -e prod -s debug_relay_backend

.PHONY: publish-debug-relay-backend-artifacts-dev
publish-debug-relay-backend-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s debug_relay_backend

.PHONY: publish-debug-relay-backend-artifacts-staging
publish-debug-relay-backend-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s debug_relay_backend

.PHONY: publish-debug-relay-backend-artifacts-prod
publish-debug-relay-backend-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s debug_relay_backend

.PHONY: publish-debug-relay-backend-artifacts-prod-debug
publish-debug-relay-backend-artifacts-prod-debug:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD_DEBUG) -s debug_relay_backend

.PHONY: deploy-debug-relay-backend-dev
deploy-debug-relay-backend-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t debug-relay-backend -n debug_relay_backend -b $(ARTIFACT_BUCKET)

.PHONY: deploy-debug-relay-backend-staging
deploy-debug-relay-backend-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t debug-relay-backend -n debug_relay_backend -b $(ARTIFACT_BUCKET_STAGING)

.PHONY: deploy-debug-relay-backend-prod
deploy-debug-relay-backend-prod:
	./deploy/deploy.sh -e prod -c prod-1 -t debug-relay-backend -n debug_relay_backend -b $(ARTIFACT_BUCKET_PROD)

.PHONY: deploy-debug-relay-backend-prod-debug
deploy-debug-relay-backend-prod-debug:
	./deploy/deploy.sh -e prod -c prod-1 -t debug-relay-backend -n debug_relay_backend -b $(ARTIFACT_BUCKET_PROD_DEBUG)

#######################
#    Relay Backend    #
#######################

.PHONY: build-relay-backend
build-relay-backend:
	@printf "Building relay backend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/relay_backend ./cmd/relay_backend/relay_backend.go
	@printf "done\n"

.PHONY: build-relay-backend-artifacts-dev
build-relay-backend-artifacts-dev: build-relay-backend
	./deploy/build-artifacts.sh -e dev -s relay_backend

.PHONY: build-relay-backend-artifacts-staging
build-relay-backend-artifacts-staging: build-relay-backend
	./deploy/build-artifacts.sh -e staging -s relay_backend

.PHONY: build-relay-backend-artifacts-prod
build-relay-backend-artifacts-prod: build-relay-backend
	./deploy/build-artifacts.sh -e prod -s relay_backend

.PHONY: deploy-relay-backend-dev-1
deploy-relay-backend-dev-1:
	./deploy/deploy.sh -e dev -c dev-1 -t relay-backend -n relay_backend -b gs://development_artifacts

.PHONY: deploy-relay-backend-dev-2
deploy-relay-backend-dev-2:
	./deploy/deploy.sh -e dev -c dev-2 -t relay-backend -n relay_backend -b gs://development_artifacts

.PHONY: deploy-relay-backend-staging-1
deploy-relay-backend-staging-1:
	./deploy/deploy.sh -e staging -c staging-1 -t relay-backend -n relay_backend -b gs://staging_artifacts

.PHONY: deploy-relay-backend-staging-2
deploy-relay-backend-staging-2:
	./deploy/deploy.sh -e staging -c staging-2 -t relay-backend -n relay_backend -b gs://staging_artifacts

.PHONY: deploy-relay-backend-prod-1
deploy-relay-backend-prod-1:
	./deploy/deploy.sh -e prod -c mig-jcr6 -t relay-backend -n relay_backend -b gs://prod_artifacts

.PHONY: deploy-relay-backend-prod-2
deploy-relay-backend-prod-2:
	./deploy/deploy.sh -e prod -c prod-2 -t relay-backend -n relay_backend -b gs://prod_artifacts

.PHONY: publish-relay-backend-artifacts-dev
publish-relay-backend-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s relay_backend

.PHONY: publish-relay-backend-artifacts-staging
publish-relay-backend-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s relay_backend

.PHONY: publish-relay-backend-artifacts-prod
publish-relay-backend-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s relay_backend

#######################
#     Ghost Army      #
#######################

.PHONY: dev-ghost-army
dev-ghost-army: build-ghost-army ## runs a local ghost army
	@./dist/ghost_army

.PHONY: build-ghost-army
build-ghost-army:
	@printf "Building ghost army... "
	@$(GO) build -o ./dist/ghost_army ./cmd/ghost_army/*.go
	@printf "done\n"

.PHONY: build-ghost-army-generator
build-ghost-army-generator:
	@printf "Building ghost army generator... "
	@$(GO) build -o ./dist/ghost_army_generator ./cmd/ghost_army_generator/*.go
	@printf "done\n"

.PHONY: build-ghost-army-analyzer
build-ghost-army-analyzer:
	@printf "Building ghost army analyzer... "
	@$(GO) build -o ./dist/ghost_army_analyzer ./cmd/ghost_army_analyzer/*.go
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

.PHONY: build-ghost-army-artifacts-dev
build-ghost-army-artifacts-dev: build-ghost-army
	./deploy/build-artifacts.sh -e dev -s ghost_army

.PHONY: build-ghost-army-artifacts-staging
build-ghost-army-artifacts-staging: build-ghost-army
	./deploy/build-artifacts.sh -e staging -s ghost_army

.PHONY: build-ghost-army-artifacts-prod
build-ghost-army-artifacts-prod: build-ghost-army
	./deploy/build-artifacts.sh -e prod -s ghost_army

.PHONY: publish-ghost-army-artifacts-dev
publish-ghost-army-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s ghost_army

.PHONY: publish-ghost-army-artifacts-staging
publish-ghost-army-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s ghost_army

.PHONY: publish-ghost-army-artifacts-prod
publish-ghost-army-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s ghost_army

.PHONY: deploy-ghost-army-dev
deploy-ghost-army-dev:
	./deploy/deploy.sh -e dev -c 1 -t ghost-army -n ghost_army -b gs://development_artifacts

.PHONY: deploy-ghost-army-staging
deploy-ghost-army-staging:
	./deploy/deploy.sh -e staging -c 1 -t ghost-army -n ghost_army -b gs://staging_artifacts

.PHONY: deploy-ghost-army-prod
deploy-ghost-army-prod:
	./deploy/deploy.sh -e prod -c 1 -t ghost-army -n ghost_army -b gs://prod_artifacts

#######################
#     Fake Relay      #
#######################

.PHONY: dev-fake-relays
dev-fake-relays: build-fake-relays ## runs a local relay forwarder
	@PORT=30007 ./dist/fake_relays

.PHONY: build-fake-relays
build-fake-relays:
	@printf "Building fake relays... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/fake_relays ./cmd/fake_relays/fake_relays.go
	@printf "done\n"

.PHONY: build-fake-relays-artifacts-dev
build-fake-relays-artifacts-dev: build-fake-relays
	./deploy/build-artifacts.sh -e dev -s fake_relays

.PHONY: publish-fake-relays-artifacts-dev
publish-fake-relays-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s fake_relays

.PHONY: build-fake-relays-artifacts-staging
build-fake-relays-artifacts-staging: build-fake-relays
	./deploy/build-artifacts.sh -e staging -s fake_relays

.PHONY: publish-fake-relays-artifacts-staging
publish-fake-relays-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s fake_relays

.PHONY: build-fake-relays-artifacts-prod
build-fake-relays-artifacts-prod: build-fake-relays
	./deploy/build-artifacts.sh -e prod -s fake_relays

.PHONY: publish-fake-relays-artifacts-prod
publish-fake-relays-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s fake_relays

#######################
# Relay Build Process #
#######################

RELAY_DIR := ./relay
RELAY_MAKEFILE := Makefile
RELAY_EXE := relay

.PHONY: build-reference-relay
build-reference-relay:
	@printf "Building reference relay... "
	@$(CXX) $(CXX_FLAGS) -o $(DIST_DIR)/reference_relay reference/relay/*.cpp $(LDFLAGS)
	@printf "done\n"

#######################
#   Relay Forwarder   #
#######################

.PHONY: dev-relay-forwarder
dev-relay-forwarder: build-relay-forwarder ## runs a local relay forwarder
	@PORT=30006 ./dist/relay_forwarder

.PHONY: build-relay-forwarder
build-relay-forwarder:
	@printf "Building relay forwarder... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/relay_forwarder ./cmd/relay_forwarder/relay_forwarder.go
	@printf "done\n"

.PHONY: build-relay-forwarder-artifacts-dev
build-relay-forwarder-artifacts-dev: build-relay-forwarder
	./deploy/build-artifacts.sh -e dev -s relay_forwarder

.PHONY: build-relay-forwarder-artifacts-staging
build-relay-forwarder-artifacts-staging: build-relay-forwarder
	./deploy/build-artifacts.sh -e staging -s relay_forwarder

.PHONY: build-relay-forwarder-artifacts-prod
build-relay-forwarder-artifacts-prod: build-relay-forwarder
	./deploy/build-artifacts.sh -e prod -s relay_forwarder

.PHONY: publish-relay-forwarder-artifacts-dev
publish-relay-forwarder-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s relay_forwarder

.PHONY: publish-relay-forwarder-artifacts-staging
publish-relay-forwarder-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s relay_forwarder

.PHONY: publish-relay-forwarder-artifacts-prod
publish-relay-forwarder-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s relay_forwarder

.PHONY: deploy-relay-forwarder-dev
deploy-relay-forwarder-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t relay-forwarder -n relay_forwarder -b gs://development_artifacts

.PHONY: deploy-relay-forwarder-staging
deploy-relay-forwarder-staging:
	./deploy/deploy.sh -e staging -c staging-1 -t relay-forwarder -n relay_forwarder -b gs://staging_artifacts

.PHONY: deploy-relay-forwarder-prod
deploy-relay-forwarder-prod:
	./deploy/deploy.sh -e prod -c prod-1 -t relay-forwarder -n relay_forwarder -b gs://prod_artifacts

#######################
#    Relay Gateway    #
#######################

.PHONY: build-relay-gateway
build-relay-gateway:
	@printf "Building relay gateway... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/relay_gateway ./cmd/relay_gateway/relay_gateway.go
	@printf "done\n"

.PHONY: build-relay-gateway-artifacts-dev
build-relay-gateway-artifacts-dev: build-relay-gateway
	./deploy/build-artifacts.sh -e dev -s relay_gateway

.PHONY: build-relay-gateway-artifacts-staging
build-relay-gateway-artifacts-staging: build-relay-gateway
	./deploy/build-artifacts.sh -e staging -s relay_gateway

.PHONY: build-relay-gateway-artifacts-prod
build-relay-gateway-artifacts-prod: build-relay-gateway
	./deploy/build-artifacts.sh -e prod -s relay_gateway

.PHONY: publish-relay-gateway-artifacts-dev
publish-relay-gateway-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s relay_gateway

.PHONY: publish-relay-gateway-artifacts-staging
publish-relay-gateway-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s relay_gateway

.PHONY: publish-relay-gateway-artifacts-prod
publish-relay-gateway-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s relay_gateway

#######################
##   Relay Frontend  ##
#######################

.PHONY: build-relay-frontend
build-relay-frontend:
	@printf "Building relay frontend... "
	@$(GO) build -ldflags "-s -w -X main.buildtime=$(TIMESTAMP) -X main.sha=$(SHA) -X main.release=$(RELEASE) -X main.commitMessage=$(echo "$COMMITMESSAGE")" -o ${DIST_DIR}/relay_frontend ./cmd/relay_frontend/relay_frontend.go
	@printf "done\n"

.PHONY: build-relay-frontend-artifacts-dev
build-relay-frontend-artifacts-dev: build-relay-frontend
	./deploy/build-artifacts.sh -e dev -s relay_frontend

.PHONY: build-relay-frontend-artifacts-staging
build-relay-frontend-artifacts-staging: build-relay-frontend
	./deploy/build-artifacts.sh -e staging -s relay_frontend

.PHONY: build-relay-frontend-artifacts-prod
build-relay-frontend-artifacts-prod: build-relay-frontend
	./deploy/build-artifacts.sh -e prod -s relay_frontend

.PHONY: publish-relay-frontend-artifacts-dev
publish-relay-frontend-artifacts-dev:
	./deploy/publish.sh -e dev -b $(ARTIFACT_BUCKET) -s relay_frontend

.PHONY: publish-relay-frontend-artifacts-staging
publish-relay-frontend-artifacts-staging:
	./deploy/publish.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s relay_frontend

.PHONY: publish-relay-frontend-artifacts-prod
publish-relay-frontend-artifacts-prod:
	./deploy/publish.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s relay_frontend

#######################

.PHONY: format
format:
	@$(GOFMT) -s -w .
	@printf "\n"

.PHONY: build-all
build-all: build-sdk build-portal-cruncher build-analytics-pusher build-analytics build-api build-vanity build-billing build-beacon build-beacon-inserter build-relay-gateway build-relay-backend build-relay-frontend build-relay-forwarder build-relay-pusher build-server-backend build-client build-server build-functional build-next ## builds everything

.PHONY: rebuild-all
rebuild-all: clean build-all ## rebuilds everything

.PHONY: update-sdk
update-sdk:
	git submodule update --remote --merge sdk

.PHONY: clean
clean: ## cleans everything
	@rm -fr $(DIST_DIR)
	@mkdir $(DIST_DIR)
