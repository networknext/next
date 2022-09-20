#!make

# IMPORTANT: Select environment before you run this makefile, eg. "next select local", "next select dev5"
include .env
export $(shell sed 's/=.*//' .env)

CXX_FLAGS := -g -Wall -Wextra -DNEXT_DEVELOPMENT=1
GO = go
GOFMT = gofmt

OS := $(shell uname -s | tr A-Z a-z)
ifeq ($(OS),darwin)
	LDFLAGS = -lsodium -lcurl -lpthread -lm -framework CoreFoundation -framework SystemConfiguration
	CXX = g++
else
	LDFLAGS = -lsodium -lcurl -lpthread -lm
	CXX = g++
endif

SDKNAME4 = libnext4
SDKNAME5 = libnext5

RELAY_PORT ?= 2000

MODULE ?= "github.com/networknext/backend/modules/common"

BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
COMMIT_MESSAGE ?= $(shell git log -1 --pretty=%B | tr "\n" " " | tr \' '*')
COMMIT_HASH ?= $(shell git rev-parse --short HEAD) 

ARTIFACT_BUCKET = gs://development_artifacts
ARTIFACT_BUCKET_STAGING = gs://staging_artifacts
ARTIFACT_BUCKET_PROD = gs://production_artifacts

####################
##    RELAY ENV   ##
####################

ifndef RELAY_BACKEND_HOSTNAME
export RELAY_BACKEND_HOSTNAME = http://127.0.0.1:30001
endif

ifndef RELAY_GATEWAY
export RELAY_GATEWAY = http://127.0.0.1:30000
endif

ifndef RELAY_FRONTEND
export RELAY_FRONTEND = http://127.0.0.1:30002
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

ifndef SERVER_BACKEND_IP
export SERVER_BACKEND_IP = 127.0.0.1:40000
endif

ifndef MAGIC_URI
export MAGIC_URI = http://127.0.0.1:41007/magic
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

ifndef RELAY_ROUTER_MAX_BANDWIDTH_PERCENTAGE
export RELAY_ROUTER_MAX_BANDWIDTH_PERCENTAGE = 90.0
endif

ifndef ROUTE_MATRIX_URI
export ROUTE_MATRIX_URI = http://127.0.0.1:30001/route_matrix
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

ifndef FEATURE_LOOKER_BIGTABLE_REPLACEMENT
export FEATURE_LOOKER_BIGTABLE_REPLACEMENT = true
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

ifndef FEATURE_RELAY_FULL_BANDWIDTH
export FEATURE_RELAY_FULL_BANDWIDTH = false
endif

.PHONY: help
help:
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\033[36m\1\\033[m:\2/' | column -c2 -t -s :)"

.PHONY: dist
dist:
	@mkdir -p dist

# Build most golang services

dist/%: cmd/%/*.go $(shell find modules -name '*.go') dist
	@echo "Building $(@F)"
	@go build -ldflags "-s -w -X $(MODULE).buildTime=$(BUILD_TIME) -X \"$(MODULE).commitMessage=$(COMMIT_MESSAGE)\" -X $(MODULE).commitHash=$(COMMIT_HASH)" -o $@ $(<D)/*.go

# Build most artifacts

dist/%.dev.tar.gz: dist/%
	@go run scripts/build_artifact/build_artifact.go $@ dev

dist/%.prod.tar.gz: dist/%
	@go run scripts/build_artifact/build_artifact.go $@ prod

# Format golang code

.PHONY: format
format:
	@$(GOFMT) -s -w .
	@printf "\n"

# Clean, build all, rebuild all

.PHONY: clean
clean: ## cleans everything
	@rm -rf dist
	@mkdir -p dist

.PHONY: build-all
build-all: build-sdk4 build-sdk5 $(shell ./scripts/all_commands.sh) ## builds everything

.PHONY: rebuild-all
rebuild-all: clean build-all ## rebuilds everything

#####################
##   Happy Path    ##
#####################

# Always run sqlite3
export FEATURE_POSTGRESQL=false
export JWT_AUDIENCES=S4WGLze2EZCPG9MeZ5509BedlWlHZGFt,dJFD1rg3Zd8PYhAXbIb0UCKFJk7IE4ho
export SLACK_WEBHOOK_URL=https://hooks.slack.com/services/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
export SLACK_CHANNEL=portal-test
export LOOKER_SECRET=d61764ff20f99e672af3ec7fde75531a790acdb6d58bf46dbe55dac06a6019c0
export LOOKER_HOST=networknextexternal.cloud.looker.com
export GITHUB_ACCESS_TOKEN=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
export RELEASE_NOTES_INTERVAL=30s

# TODO: Change these to a permanent API user in looker
export LOOKER_API_CLIENT_ID=QXG3cfyWd8xqsVnT7QbT
export LOOKER_API_CLIENT_SECRET=JT2BpTYNc7fybyHNGs3S24g7

.PHONY: dev-redis-monitor
dev-redis-monitor: dist/redis_monitor ## runs a local redis monitor
	@HTTP_PORT=41008 ./dist/redis_monitor

.PHONY: dev-magic-backend
dev-magic-backend: dist/magic_backend ## runs a local magic backend
	@HTTP_PORT=41007 ./dist/magic_backend

.PHONY: dev-relay-gateway
dev-relay-gateway: ./dist/relay_gateway ## runs a local relay gateway
	@HTTP_PORT=30000 RELAY_UPDATE_BATCH_DURATION=1s ./dist/relay_gateway

.PHONY: dev-relay-backend
dev-relay-backend: ./dist/relay_backend ## runs a local relay backend (#1)
	@HTTP_PORT=30001 READY_DELAY=5s ./dist/relay_backend

.PHONY: dev-relay-backend-2
dev-relay-backend-2: ./dist/relay_backend ## runs a local relay backend (#2)
	@HTTP_PORT=30002 READY_DELAY=5s ./dist/relay_backend

.PHONY: dev-relay
dev-relay: build-reference-relay  ## runs a local relay
	@RELAY_DEBUG=1 RELAY_ADDRESS=127.0.0.1:$(RELAY_PORT) ./dist/reference_relay

.PHONY: dev-server-backend4
dev-server-backend4: build-server-backend4 ## runs a local server backend (sdk4)
	@HTTP_PORT=40000 UDP_PORT=40000 ./dist/server_backend4

.PHONY: dev-server-backend5
dev-server-backend5: build-server-backend5 ## runs a local server backend (sdk5)
	@HTTP_PORT=45000 UDP_PORT=45000 ./dist/server_backend5

##############################################

.PHONY: build-test-server4
build-test-server4: build-sdk4
	@printf "Building test server 4... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o test_server4 ../cmd/test_server4/test_server4.cpp $(SDKNAME4).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-server4
build-server4: build-sdk4
	@printf "Building server 4... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o server4 ../cmd/server4/server4.cpp $(SDKNAME4).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-client4
build-client4: build-sdk4
	@printf "Building client 4... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o client4 ../cmd/client4/client4.cpp $(SDKNAME4).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-test4
build-test4: build-sdk4
	@printf "Building test 4... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o test4 ../sdk4/test.cpp $(SDKNAME4).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-test-server5
build-test-server5: build-sdk5
	@printf "Building test server 5... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o test_server5 ../cmd/test_server5/test_server5.cpp $(SDKNAME5).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-server5
build-server5: build-sdk5
	@printf "Building server 5... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o server5 ../cmd/server5/server5.cpp $(SDKNAME5).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-client5
build-client5: build-sdk5
	@printf "Building client 5... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o client5 ../cmd/client5/client5.cpp $(SDKNAME5).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-test5
build-test5: build-sdk5
	@printf "Building test 5... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o test5 ../sdk5/test.cpp $(SDKNAME5).so $(LDFLAGS)
	@printf "done\n"

.PHONY: dev-client4
dev-client4: build-client4  ## runs a local client (sdk4)
	@cd dist && ./client4

.PHONY: dev-server4
dev-server4: build-sdk4 build-server4  ## runs a local server (sdk4)
	@cd dist && ./server4

.PHONY: dev-client5
dev-client5: build-client5  ## runs a local client (sdk5)
	@cd dist && ./client5

.PHONY: dev-server5
dev-server5: build-sdk5 build-server5  ## runs a local server (sdk5)
	@cd dist && ./server5

##########################################

.PHONY: dev-portal
dev-portal: build-portal ## runs a local portal
	@PORT=20000 BASIC_AUTH_USERNAME=local BASIC_AUTH_PASSWORD=local ANALYTICS_MIG=localhost:41001 ANALYTICS_PUSHER_URI=localhost:41002 PORTAL_BACKEND_MIG=localhost:20000 PORTAL_CRUNCHER_URI=localhost:42000 BILLING_MIG=localhost:41000 RELAY_FRONTEND_URI=localhost:30005 RELAY_GATEWAY_URI=localhost:30000 RELAY_PUSHER_URI=localhost:30004 SERVER_BACKEND_MIG=localhost:40000 ./dist/portal

.PHONY: dev-analytics
dev-analytics: dist/analytics ## runs a local analytics service
	@PORT=41001 ./dist/analytics

.PHONY: dev-portal-cruncher-1
dev-portal-cruncher-1: build-portal-cruncher ## runs a local portal cruncher
	@HTTP_PORT=42000 CRUNCHER_PORT=5555 ./dist/portal_cruncher

.PHONY: dev-portal-cruncher-2
dev-portal-cruncher-2: build-portal-cruncher ## runs a local portal cruncher
	@HTTP_PORT=42001 CRUNCHER_PORT=5556 ./dist/portal_cruncher

.PHONY: dev-pingdom
dev-pingdom: build-pingdom ## runs the pulling and publishing of pingdom uptime
	@PORT=41006 ./dist/pingdom

#####################
## ESSENTIAL TOOLS ##
#####################

.PHONY: test
test: clean ## runs backend unit tests
	@./scripts/test-unit-backend.sh

.PHONY: test-sdk4
test-sdk4: dist build-test4 ## runs sdk4 unit tests
	@cd dist && ./test4

.PHONY: test-sdk5
test-sdk5: dist build-test5 ## runs sdk5 unit tests
	@cd dist && ./test5

.PHONY: test-relay
test-relay: dist build-reference-relay ## runs relay unit tests
	cd dist && ./reference_relay test

ifeq ($(OS),darwin)
.PHONY: build-load-test-server
build-load-test-server: dist build-sdk4
	@printf "Building load test server... "
	@$(CXX) $(CXX_FLAGS) -Isdk4/include -o dist/load_test_server ./cmd/load_test_server/load_test_server.cpp  dist/$(SDKNAME4).so $(LDFLAGS)
	@printf "done\n"
else
.PHONY: build-load-test-server
build-load-test-server: dist build-sdk4
	@printf "Building load test server... "
	@$(CXX) $(CXX_FLAGS) -Isdk4/include -o dist/load_test_server ./cmd/load_test_server/load_test_server.cpp -L./dist -lnext $(LDFLAGS)
	@printf "done\n"
endif

ifeq ($(OS),darwin)
.PHONY: build-load-test-client
build-load-test-client: dist build-sdk4
	@printf "Building load test client... "
	@$(CXX) $(CXX_FLAGS) -Isdk4/include -o dist/load_test_client ./cmd/load_test_client/load_test_client.cpp dist/$(SDKNAME4).so $(LDFLAGS)
	@printf "done\n"
else
.PHONY: build-load-test-client
build-load-test-client: dist build-sdk4
	@printf "Building load test client... "
	@$(CXX) $(CXX_FLAGS) -Isdk4/include -o dist/load_test_client ./cmd/load_test_client/load_test_client.cpp -L./dist -lnext $(LDFLAGS)
	@printf "done\n"
endif

########################

.PHONY: build-functional-tests-backend
build-functional-tests-backend: dist
	@printf "Building functional tests backend... " ; \
	$(GO) build -o ./dist/func_tests_backend ./cmd/func_tests_backend/*.go ; \
	printf "done\n" ; \

.PHONY: build-test-func-backend
build-test-func-backend: dist build-functional-tests-backend

.PHONY: run-test-func-backend
run-test-func-backend:
	@printf "\nRunning functional tests backend...\n\n" ; \
	cd dist && $(GO) run ../cmd/func_tests_backend/func_tests_backend.go $(test) ; \
	printf "\ndone\n\n"

.PHONY: test-func-backend
test-func-backend: build-test-func-backend run-test-func-backend ## runs functional tests (backend)

########################

.PHONY: build-functional-server4
build-functional-server4: build-sdk4
	@printf "Building functional server 4... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o func_server4 ../cmd/func_server4/func_server4.cpp $(SDKNAME4).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-client4
build-functional-client4: build-sdk4
	@printf "Building functional client 4... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o func_client4 ../cmd/func_client4/func_client4.cpp $(SDKNAME4).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-backend4
build-functional-backend4: dist
	@printf "Building functional backend 4... " ; \
	$(GO) build -o ./dist/func_backend4 ./cmd/func_backend4/*.go ; \
	printf "done\n" ; \

.PHONY: build-functional-tests-sdk4
build-functional-tests-sdk4: dist
	@printf "Building functional tests sdk4... " ; \
	$(GO) build -o ./dist/func_tests_sdk4 ./cmd/func_tests_sdk4/*.go ; \
	printf "done\n" ; \

.PHONY: build-test-func-sdk4
build-test-func-sdk4: clean dist build-sdk4 build-reference-relay build-functional-server4 build-functional-client4 build-functional-backend4 build-functional-tests-sdk4

.PHONY: run-test-func-sdk4
run-test-func-sdk4:
	@printf "\nRunning functional tests sdk4...\n\n" ; \
	cd dist && $(GO) run ../cmd/func_tests_sdk4/func_tests_sdk4.go $(test) ; \
	printf "\ndone\n\n"

.PHONY: test-func-sdk4
test-func-sdk4: build-test-func-sdk4 run-test-func-sdk4 ## runs functional tests (sdk4)

#######################

.PHONY: build-functional-server5
build-functional-server5: build-sdk5
	@printf "Building functional server 5... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o func_server5 ../cmd/func_server5/func_server5.cpp $(SDKNAME5).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-client5
build-functional-client5: build-sdk5
	@printf "Building functional client 5... "
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o func_client5 ../cmd/func_client5/func_client5.cpp $(SDKNAME5).so $(LDFLAGS)
	@printf "done\n"

.PHONY: build-functional-backend5
build-functional-backend5: dist
	@printf "Building functional backend 5... " ; \
	$(GO) build -o ./dist/func_backend5 ./cmd/func_backend5/*.go ; \
	printf "done\n" ; \

.PHONY: build-functional-tests-sdk5
build-functional-tests-sdk5: dist
	@printf "Building functional tests sdk5... " ; \
	$(GO) build -o ./dist/func_tests_sdk5 ./cmd/func_tests_sdk5/*.go ; \
	printf "done\n" ; \

.PHONY: build-test-func-sdk5
build-test-func-sdk5: clean dist build-sdk5 build-reference-relay build-functional-server5 build-functional-client5 build-functional-backend5 build-functional-tests-sdk5

.PHONY: run-test-func-sdk5
run-test-func-sdk5:
	@printf "\nRunning functional tests sdk5...\n\n" ; \
	cd dist && $(GO) run ../cmd/func_tests_sdk5/func_tests_sdk5.go $(test) ; \
	printf "\ndone\n\n"

.PHONY: test-func-sdk5
test-func-sdk5: build-test-func-sdk5 run-test-func-sdk5 ## runs functional tests (sdk5)

#######################

.PHONY: dev-happy-path
dev-happy-path: ## runs the happy path
	@printf "\ndon't worry. be happy.\n\n" ; \
	./build.sh
	$(GO) build -o ./dist/happy_path ./cmd/happy_path/happy_path.go
	./dist/happy_path

#######################

.PHONY: dev-ref-backend4
dev-ref-backend4: ## runs a local reference backend (sdk4)
	$(GO) run reference/backend4/backend4.go

.PHONY: dev-ref-backend5
dev-ref-backend5: ## runs a local reference backend (sdk5)
	$(GO) run reference/backend5/backend5.go

.PHONY: dev-mock-relay
dev-mock-relay: ## runs a local mock relay
	$(GO) build -o ./dist/mock_relay ./cmd/mock_relay/mock_relay.go
	./dist/mock_relay

.PHONY: dev-fake-server
dev-fake-server: build-fake-server ## runs a fake server that simulates 2 servers and 400 clients locally
	@HTTP_PORT=50001 UDP_PORT=50000 ./dist/fake_server

dist/$(SDKNAME4).so: dist
	@printf "Building sdk4... "
	@cd dist && $(CXX) $(CXX_FLAGS) -fPIC -I../sdk4/include -shared -o $(SDKNAME4).so ../sdk4/source/*.cpp $(LDFLAGS)
	@printf "done\n"

dist/$(SDKNAME5).so: dist
	@printf "Building sdk5... "
	@cd dist && $(CXX) $(CXX_FLAGS) -fPIC -DNEXT_COMPILE_WITH_TESTS=1 -I../sdk5/include -shared -o $(SDKNAME5).so ../sdk5/source/*.cpp $(LDFLAGS)
	@printf "done\n"

.PHONY: build-sdk4
build-sdk4: dist/$(SDKNAME4).so

.PHONY: build-sdk5
build-sdk5: dist/$(SDKNAME5).so

PHONY: build-portal-cruncher
build-portal-cruncher:
	@printf "Building portal cruncher... "
	@$(GO) build -ldflags "-s -w -X main.buildTime=$(BUILD_TIME) -X 'main.commitMessage=$(COMMIT_MESSAGE)' -X main.commitMessage=$(COMMIT_HASH)" -o ./dist/portal_cruncher ./cmd/portal_cruncher/portal_cruncher.go
	@printf "done\n"

.PHONY: build-portal
build-portal:
	@printf "Building portal... \n"
	@printf "TIMESTAMP: ${TIMESTAMP}\n"
	@printf "SHA: ${SHA}\n"
	@printf "RELEASE: ${RELEASE}\n"
	@printf "COMMITMESSAGE: ${COMMITMESSAGE}\n"
	@$(GO) build -ldflags "-s -w -X main.buildTime=$(BUILD_TIME) -X 'main.commitMessage=$(COMMIT_MESSAGE)' -X main.commitMessage=$(COMMIT_HASH)" -o dist/portal ./cmd/portal/portal.go
	@printf "done\n"

.PHONY: build-fake-server
build-fake-server: dist
	@printf "Building fake server... "
	@$(GO) build -ldflags "-s -w -X main.buildTime=$(BUILD_TIME) -X 'main.commitMessage=$(COMMIT_MESSAGE)' -X main.commitMessage=$(COMMIT_HASH)" -o dist/fake_server ./cmd/fake_server/fake_server.go
	@printf "done\n"

.PHONY: build-pingdom
build-pingdom: dist
	@printf "Building pingdom... "
	@$(GO) build -ldflags "-s -w -X main.buildTime=$(BUILD_TIME) -X 'main.commitMessage=$(COMMIT_MESSAGE)' -X main.commitMessage=$(COMMIT_HASH)" -o dist/pingdom ./cmd/pingdom/pingdom.go
	@printf "done\n"

.PHONY: deploy-portal-crunchers-dev
deploy-portal-crunchers-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t portal-cruncher -n portal_cruncher -b gs://development_artifacts
	./deploy/deploy.sh -e dev -c dev-2 -t portal-cruncher -n portal_cruncher -b gs://development_artifacts

.PHONY: deploy-pingdom-dev
deploy-pingdom-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t pingdom -n pingdom -b gs://development_artifacts

.PHONY: build-pingdom-artifacts-dev
build-pingdom-artifacts-dev: build-pingdom
	./deploy/build-artifacts.sh -e dev -s pingdom

.PHONY: build-portal-artifacts-dev
build-portal-artifacts-dev: build-portal
	./deploy/build-artifacts.sh -e dev -s portal -b $(ARTIFACT_BUCKET)

.PHONY: build-portal-artifacts-dev-old
build-portal-artifacts-dev-old: build-portal
	./deploy/build-artifacts.sh -e dev -s portal-old -b $(ARTIFACT_BUCKET)

.PHONY: build-portal-cruncher-artifacts-dev
build-portal-cruncher-artifacts-dev: build-portal-cruncher
	./deploy/build-artifacts.sh -e dev -s portal_cruncher

.PHONY: build-test-server4-artifacts-dev
build-test-server4-artifacts-dev: build-test-server4
	./deploy/build-artifacts.sh -e dev -s test_server4

.PHONY: build-test-server5-artifacts-dev
build-test-server5-artifacts-dev: build-test-server5
	./deploy/build-artifacts.sh -e dev -s test_server5

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
	@$(GO) build -ldflags "-s -w -X main.buildTime=$(BUILD_TIME) -X 'main.commitMessage=$(COMMIT_MESSAGE)' -X main.commitMessage=$(COMMIT_HASH)" -o dist/relay_pusher ./cmd/relay_pusher/relay_pusher.go
	@printf "done\n"

.PHONY: build-relay-pusher-artifacts-dev
build-relay-pusher-artifacts-dev: build-relay-pusher
	./deploy/build-artifacts.sh -e dev -s relay_pusher

.PHONY: deploy-relay-pusher-dev
deploy-relay-pusher-dev:
	./deploy/deploy.sh -e dev -c dev-1 -t relay-pusher -n relay_pusher -b gs://development_artifacts

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

.PHONY: build-ghost-army-artifacts-dev
build-ghost-army-artifacts-dev: build-ghost-army
	./deploy/build-artifacts.sh -e dev -s ghost_army

.PHONY: deploy-ghost-army-dev
deploy-ghost-army-dev:
	./deploy/deploy.sh -e dev -c 1 -t ghost-army -n ghost_army -b gs://development_artifacts

#######################
#     Fake Relay      #
#######################

.PHONY: dev-fake-relays
dev-fake-relays: build-fake-relays ## runs local fake relays
	@PORT=30007 ./dist/fake_relays

.PHONY: build-fake-relays
build-fake-relays:
	@printf "Building fake relays... "
	@$(GO) build -ldflags "-s -w -X main.buildTime=$(BUILD_TIME) -X 'main.commitMessage=$(COMMIT_MESSAGE)' -X main.commitMessage=$(COMMIT_HASH)" -o dist/fake_relays ./cmd/fake_relays/fake_relays.go
	@printf "done\n"

.PHONY: build-fake-relays-artifacts-dev
build-fake-relays-artifacts-dev: build-fake-relays
	./deploy/build-artifacts.sh -e dev -s fake_relays

#########
# Relay #
#########

.PHONY: build-reference-relay
build-reference-relay: dist
	@echo "Building reference relay..."
	@$(CXX) $(CXX_FLAGS) -o dist/reference_relay reference/relay/*.cpp $(LDFLAGS)

#######################

.PHONY: dev-pubsub-emulator
dev-pubsub-emulator:
	@-pkill -f "gcloud.py beta emulators pubsub"
	@-pkill -f "pubsub-emulator"
	gcloud beta emulators pubsub start --project=local --host-port=127.0.0.1:9000

.PHONY: dev-bigquery-emulator
dev-bigquery-emulator:
	bigquery-emulator --project="local" --dataset="local"
