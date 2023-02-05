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

BUILD_TIME ?= $(shell date -u +'%Y-%m-%d|%H:%M:%S')
COMMIT_MESSAGE ?= $(shell git log -1 --pretty=%B | tr "\n" " " | tr \' '*')
COMMIT_HASH ?= $(shell git rev-parse --short HEAD) 

# Clean, build and rebuild

.PHONY: build
build:
	@make -s build-fast -j

.PHONY: build-fast
build-fast: dist/$(SDKNAME4).so dist/$(SDKNAME5).so dist/reference_relay dist/reference_backend4 dist/reference_backend5 dist/client4 dist/server4 dist/test4 dist/client5 dist/server5 dist/test5 func-test-sdk5 func-test-sdk4 dist/raspberry_server dist/raspberry_client $(shell ./scripts/all_commands.sh)

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

dist/%.dev.tar.gz: dist/%
	@go run tools/artifact/artifact.go $@ dev
	@echo $@

dist/%.prod.tar.gz: dist/%
	@go run tools/artifact/artifact.go $@ prod
	@echo $@

# Format code

.PHONY: format
format:
	@gofmt -s -w .
	@./scripts/tabs2spaces.sh

# Build sdk4

dist/$(SDKNAME4).so: $(shell find sdk4 -type f)
	@cd dist && $(CXX) $(CXX_FLAGS) -fPIC -I../sdk4/include -shared -o $(SDKNAME4).so ../sdk4/source/*.cpp $(LDFLAGS)
	@echo $@

dist/client4: dist/$(SDKNAME4).so cmd/client4/client4.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o client4 ../cmd/client4/client4.cpp $(SDKNAME4).so $(LDFLAGS)
	@echo $@

dist/server4: dist/$(SDKNAME4).so cmd/server4/server4.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o server4 ../cmd/server4/server4.cpp $(SDKNAME4).so $(LDFLAGS)
	@echo $@

dist/test4: dist/$(SDKNAME4).so sdk4/test.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o test4 ../sdk4/test.cpp $(SDKNAME4).so $(LDFLAGS)
	@echo $@

# Build sdk5

dist/$(SDKNAME5).so: $(shell find sdk5 -type f)
	@cd dist && $(CXX) $(CXX_FLAGS) -fPIC -I../sdk5/include -shared -o $(SDKNAME5).so ../sdk5/source/*.cpp $(LDFLAGS)
	@echo $@

dist/client5: dist/$(SDKNAME5).so cmd/client5/client5.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o client5 ../cmd/client5/client5.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

dist/server5: dist/$(SDKNAME5).so cmd/server5/server5.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o server5 ../cmd/server5/server5.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

dist/test5: dist/$(SDKNAME5).so sdk5/test.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o test5 ../sdk5/test.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

# Build reference binaries

dist/reference_relay: reference/reference_relay/*
	@$(CXX) $(CXX_FLAGS) -o dist/reference_relay reference/reference_relay/*.cpp $(LDFLAGS)
	@echo $@

dist/reference_backend4: reference/reference_backend4/*.go
	@go build -o $@ reference/reference_backend4/*.go
	@echo $@

dist/reference_backend5: reference/reference_backend5/*.go
	@go build -o $@ reference/reference_backend5/*.go
	@echo $@

# Functional tests (sdk4)

dist/func_server4: dist/$(SDKNAME4).so cmd/func_server4/*
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o func_server4 ../cmd/func_server4/func_server4.cpp $(SDKNAME4).so $(LDFLAGS)
	@echo $@

dist/func_client4: dist/$(SDKNAME4).so cmd/func_client4/*
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk4/include -o func_client4 ../cmd/func_client4/func_client4.cpp $(SDKNAME4).so $(LDFLAGS)
	@echo $@

.PHONY: func-test-sdk4
func-test-sdk4: dist/reference_relay dist/func_server4 dist/func_client4 dist/func_backend4 dist/func_tests_sdk4

# Functional tests (sdk5)

dist/func_server5: dist/$(SDKNAME5).so cmd/func_server5/*
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o func_server5 ../cmd/func_server5/func_server5.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

dist/func_client5: dist/$(SDKNAME5).so cmd/func_client5/*
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o func_client5 ../cmd/func_client5/func_client5.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

.PHONY: func-test-sdk5
func-test-sdk5: dist/reference_relay dist/func_server5 dist/func_client5 dist/func_backend5 dist/func_tests_sdk5

# Raspberry

dist/raspberry_client: dist/$(SDKNAME5).so cmd/raspberry_client/raspberry_client.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o raspberry_client ../cmd/raspberry_client/raspberry_client.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@

dist/raspberry_server: dist/$(SDKNAME5).so cmd/raspberry_server/raspberry_server.cpp
	@cd dist && $(CXX) $(CXX_FLAGS) -I../sdk5/include -o raspberry_server ../cmd/raspberry_server/raspberry_server.cpp $(SDKNAME5).so $(LDFLAGS)
	@echo $@
