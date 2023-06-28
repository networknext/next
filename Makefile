# Network Next Makefile

CXX_FLAGS := -g -Wall -Wextra

OS := $(shell uname -s | tr A-Z a-z)
ifeq ($(OS),darwin)
	LDFLAGS = -lsodium -lcurl -lpthread -lm -framework CoreFoundation -framework SystemConfiguration
	CXX = g++
else
	LDFLAGS = -lsodium -lcurl -lpthread -lm
	CXX = g++
endif

SDKNAME5 = libnext5

MODULE ?= "github.com/networknext/accelerate/modules/common"

BUILD_TIME ?= $(shell date -u +'%Y-%m-%d|%H:%M:%S')
COMMIT_MESSAGE ?= $(shell git log -1 --pretty=%B | tr "\n" " " | tr \' '*')
COMMIT_HASH ?= $(shell git rev-parse --short HEAD) 

# Build and run tests by default

.PHONY: test
test: build
	./run test

# Clean, build and rebuild

.PHONY: build
build:
	@make -s build-fast -j

.PHONY: build-fast
build-fast: dist/$(SDKNAME5).so dist/relay-debug dist/relay-release dist/client dist/server dist/test dist/raspberry_server dist/raspberry_client dist/func_server dist/func_client $(shell ./scripts/all_commands.sh)

.PHONY: rebuild
rebuild: clean ## rebuild everything
	@echo rebuilding...
	@make build -j

.PHONY: clean
clean: ## clean everything
	@rm -rf dist
	@rm -rf logs
	@mkdir -p dist

# Build most golang services

dist/%: cmd/%/*.go $(shell find modules -name '*.go')
	@go build -ldflags "-s -w -X $(MODULE).buildTime=$(BUILD_TIME) -X \"$(MODULE).commitMessage=$(COMMIT_MESSAGE)\" -X $(MODULE).commitHash=$(COMMIT_HASH)" -o $@ $(<D)/*.go
	@echo $@

# Build artifacts

dist/%.tar.gz: dist/%
	@go run tools/artifact/artifact.go $@
	@echo $@

# Format code

.PHONY: format
format:
	@gofmt -s -w .
	@./scripts/tabs2spaces.sh

# Build sdk5

SDK_FLAGS := -DNEXT_DEVELOPMENT=1 -DNEXT_COMPILE_WITH_TESTS=1 

dist/$(SDKNAME5).so: $(shell find sdk5 -type f)
	@cd dist && $(CXX) $(CXX_FLAGS) $(SDK_FLAGS) -fPIC -I../sdk5/include -shared -o $(SDKNAME5).so ../sdk5/source/*.cpp $(LDFLAGS)
	@echo $@

dist/client: dist/$(SDKNAME5).so cmd/client/client.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) $(SDK_FLAGS) -I../sdk5/include -o client ../cmd/client/client.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

dist/server: dist/$(SDKNAME5).so cmd/server/server.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) $(SDK_FLAGS) -I../sdk5/include -o server ../cmd/server/server.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

dist/test: dist/$(SDKNAME5).so sdk5/test.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) $(SDK_FLAGS) -I../sdk5/include -o test ../sdk5/test.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

# Build relay

dist/relay-debug: relay/*
	@$(CXX) $(CXX_FLAGS) -DRELAY_DEBUG=1 -o dist/relay-debug relay/*.cpp $(LDFLAGS)
	@echo $@

dist/relay-release: relay/*
	@$(CXX) $(CXX_FLAGS) -O3 -DNDEBUG -o dist/relay-release relay/*.cpp $(LDFLAGS)
	@echo $@

# Functional tests (sdk5)

dist/func_server: dist/$(SDKNAME5).so cmd/func_server/*
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o func_server ../cmd/func_server/func_server.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

dist/func_client: dist/$(SDKNAME5).so cmd/func_client/*
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o func_client ../cmd/func_client/func_client.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

# Raspberry

dist/raspberry_client: dist/$(SDKNAME5).so cmd/raspberry_client/raspberry_client.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o raspberry_client ../cmd/raspberry_client/raspberry_client.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

dist/raspberry_server: dist/$(SDKNAME5).so cmd/raspberry_server/raspberry_server.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o raspberry_server ../cmd/raspberry_server/raspberry_server.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@
