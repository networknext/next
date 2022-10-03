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

# Build most golang services

dist/%: cmd/%/*.go $(shell find modules -name '*.go')
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
build: dist/client4 dist/server4 dist/test4 dist/client5 dist/server5 dist/test5 $(shell ./scripts/all_commands.sh) ## build everything

.PHONY: rebuild
rebuild: clean build ## rebuild everything

# Build sdk4

dist/$(SDKNAME4).so: $(shell find sdk4 -type f)
	@cd dist && $(CXX) $(CXX_FLAGS) -fPIC -I../sdk4/include -shared -o $(SDKNAME4).so ../sdk4/source/*.cpp $(LDFLAGS)

dist/client4: dist/$(SDKNAME4).so cmd/client4/client4.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o client4 ../cmd/client4/client4.cpp $(SDKNAME4).so $(LDFLAGS)

dist/server4: dist/$(SDKNAME4).so cmd/server4/server4.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o server4 ../cmd/server4/server4.cpp $(SDKNAME4).so $(LDFLAGS)

dist/test4: dist/$(SDKNAME4).so sdk4/test.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o test4 ../sdk4/test.cpp $(SDKNAME4).so $(LDFLAGS)

# Build sdk5

dist/$(SDKNAME5).so: $(shell find sdk5 -type f)
	@cd dist && $(CXX) $(CXX_FLAGS) -fPIC -I../sdk5/include -shared -o $(SDKNAME5).so ../sdk5/source/*.cpp $(LDFLAGS)

dist/client5: dist/$(SDKNAME5).so cmd/client5/client5.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o client5 ../cmd/client5/client5.cpp $(SDKNAME5).so $(LDFLAGS)

dist/server5: dist/$(SDKNAME5).so cmd/server5/server5.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o server5 ../cmd/server5/server5.cpp $(SDKNAME5).so $(LDFLAGS)

dist/test5: dist/$(SDKNAME5).so sdk5/test.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o test5 ../sdk5/test.cpp $(SDKNAME5).so $(LDFLAGS)

# Build reference relay

dist/reference_relay: reference/relay/*
	@$(CXX) $(CXX_FLAGS) -o dist/reference_relay reference/relay/*.cpp $(LDFLAGS)

# Functional tests (sdk4)

dist/func_server4: dist/$(SDKNAME4).so dist/cmd/func_server4/*
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o func_server4 ../cmd/func_server4/func_server4.cpp $(SDKNAME4).so $(LDFLAGS)

dist/func_client4: dist/$(SDKNAME4).so dist/cmd/func_client4/*
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o func_client4 ../cmd/func_client4/func_client4.cpp $(SDKNAME4).so $(LDFLAGS)
