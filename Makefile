# Network Next Makefile

CXX_FLAGS := -g -Wall -Wextra -DNEXT_DEVELOPMENT=1 -DNEXT_COMPILE_WITH_TESTS=1

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

MODULE ?= "github.com/networknext/backend/modules/common"

BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
COMMIT_MESSAGE ?= $(shell git log -1 --pretty=%B | tr "\n" " " | tr \' '*')
COMMIT_HASH ?= $(shell git rev-parse --short HEAD) 

.PHONY: help
help:
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\033[36m\1\\033[m:\2/' | column -c2 -t -s :)"

.PHONY: dist
dist:
	@mkdir -p dist

# Build most golang services

dist/%: cmd/%/*.go $(shell find modules -name '*.go') dist
	@go build -ldflags "-s -w -X $(MODULE).buildTime=$(BUILD_TIME) -X \"$(MODULE).commitMessage=$(COMMIT_MESSAGE)\" -X $(MODULE).commitHash=$(COMMIT_HASH)" -o $@ $(<D)/*.go

# Build most artifacts

dist/%.dev.tar.gz: dist/%
	@go run scripts/artifact/artifact.go $@ dev

dist/%.prod.tar.gz: dist/%
	@go run scripts/artifact/artifact.go $@ prod

# Clean and rebuild

.PHONY: clean
clean: ## clean everything
	@rm -rf dist
	@mkdir -p dist

.PHONY: build
build: sdk4 client4 server4 sdk5 client5 server5 $(shell ./scripts/all_commands.sh) ## build everything

.PHONY: rebuild
rebuild: clean build ## rebuild everything

# --------------------------------------------------------------

dist/$(SDKNAME4).so: dist
	@cd dist && $(CXX) $(CXX_FLAGS) -fPIC -I../sdk4/include -shared -o $(SDKNAME4).so ../sdk4/source/*.cpp $(LDFLAGS)

.PHONY: sdk4
sdk4: dist/$(SDKNAME4).so

.PHONY: client4
client4: sdk4 ## build client (sdk4)
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o client4 ../cmd/client4/client4.cpp $(SDKNAME4).so $(LDFLAGS)

.PHONY: server4
server4: sdk4 ## build server (sdk4)
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o server4 ../cmd/server4/server4.cpp $(SDKNAME4).so $(LDFLAGS)

.PHONY: test4
test4: sdk4 ## build tests (sdk4)
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o test4 ../sdk4/test.cpp $(SDKNAME4).so $(LDFLAGS)

# --------------------------------------------------------------

dist/$(SDKNAME5).so: dist
	@cd dist && $(CXX) $(CXX_FLAGS) -fPIC -I../sdk5/include -shared -o $(SDKNAME5).so ../sdk5/source/*.cpp $(LDFLAGS)

.PHONY: sdk5
sdk5: dist/$(SDKNAME5).so

.PHONY: client5
client5: sdk5 ## build client (sdk5)
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o client5 ../cmd/client5/client5.cpp $(SDKNAME5).so $(LDFLAGS)

.PHONY: server5
server5: sdk5 ## build server (sdk5)
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o server5 ../cmd/server5/server5.cpp $(SDKNAME5).so $(LDFLAGS)

.PHONY: test5
test5: sdk5 ## build tests (sdk5)
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o test5 ../sdk5/test.cpp $(SDKNAME5).so $(LDFLAGS)

# --------------------------------------------------------------

.PHONY: reference-relay
reference-relay: dist ## build reference relay
	@$(CXX) $(CXX_FLAGS) -o dist/reference_relay reference/relay/*.cpp $(LDFLAGS)

# --------------------------------------------------------------
