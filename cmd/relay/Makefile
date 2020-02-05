#################
### Functions ###
#################

rwildcard=$(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))

#################
### Variables ###
#################

SHELL				:= /bin/bash

CXX					:= g++
CXX_FLAGS 			:= -Wall -Wextra -std=c++17 
CXX_DBG_FLAGS		:= -g -O0
CXX_PERF_FLAGS		:= -O3 -march=native -frename-registers -funroll-loops -fopenmp -D_GLIBCXX_PARALLEL

SRC					:= src
INCLUDE				:= $(SRC) # include path is the same as the src dir
SPEC				:= spec

INCLUDE_DIRS		:= -I$(INCLUDE)
STATIC_LIBS			:=
LIBRARY_DIRS		:= -L.
LIBRARIES			:= -lsodium -lcurl -lpthread -lm

EXE					:= relay
EXE_DBG				:= relay.dbg
EXE_TEST			:= relay.test
EXE_BENCH			:= relay.bench

BIN					:= bin
OBJ					:= obj

SUB_DIRS			:= $(shell find src/* -type d -print)
OBJ_DIRS			:= $(patsubst $(SRC)/%, $(OBJ)/%, $(SUB_DIRS))

SRC_FILES			:= $(call rwildcard,$(SRC),*.cpp)

SRC_OBJ_FILES		:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.o, $(SRC_FILES))
SRC_DEP_FILES		:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.d, $(SRC_FILES))

DBG_OBJ_FILES		:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.dbg.o, $(SRC_FILES))
DBG_DEP_FILES		:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.dbg.d, $(SRC_FILES))

TEST_OBJ_FILES		:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.test.o, $(SRC_FILES))
TEST_DEP_FILES		:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.test.d, $(SRC_FILES))

BENCH_OBJ_FILES		:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.bench.o, $(SRC_FILES))
BENCH_DEP_FILES		:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.bench.d, $(SRC_FILES))

DIST_DIR			:= ../../dist

OBJECTS				:= $(SRC_OBJ_FILES) $(DBG_OBJ_FILES) $(TEST_OBJ_FILES) $(BENCH_OBJ_FILES)
DEPENDENCIES		:= $(SRC_DEP_FILES) $(DBG_DEP_FILES) $(TEST_DEP_FILES) $(BENCH_DEP_FILES)
EXECUTABLES			:= $(BIN)/$(EXE) $(BIN)/$(EXE_DBG) $(BIN)/$(EXE_TEST) $(BIN)/$(EXE_BENCH) $(DIST_DIR)/$(EXE)

################
### Targets  ###
################

.PHONY: relay 
relay: $(BIN) $(OBJ) $(OBJ_DIRS) $(BIN)/$(EXE) $(DIST_DIR)/$(EXE)

.PHONY: relay-dbg
relay-dbg: $(BIN) $(OBJ) $(OBJ_DIRS) $(BIN)/$(EXE_DBG)

.PHONY: relay-test
relay-test: $(BIN) $(OBJ) $(OBJ_DIRS) $(BIN)/$(EXE_TEST)

.PHONY: relay-benchmarks
relay-benchmarks: $(BIN) $(OBJ) $(OBJ_DIRS) $(BIN)/$(EXE_BENCH)

.PHONY: make-dbg
make-dbg: relay-dbg

.PHONY:
run-dbg: make-dbg
	@$(BIN)/$(EXE_DBG)

.PHONY: make-test
make-test: relay-test

.PHONY: run-tests
run-tests: make-test
	@$(BIN)/$(EXE_TEST) test

.PHONY: make-benchmarks
make-benchmarks: relay-benchmarks

.PHONY: run-benchmarks
run-benchmarks: make-benchmarks
	@$(BIN)/$(EXE_BENCH) bench

.PHONY: clean
clean:
	-rm -f $(OBJECTS) $(DEPENDENCIES) $(EXECUTABLES)

.PHONY: force
force: clean relay relay-test relay-benchmarks

# This must be after "relay" otherwise the first target that comes out of this will be the one make defaults to
-include $(DEPENDENCIES)

##############
### Builds ###
##############

### Release ###

$(BIN)/$(EXE): $(SRC_OBJ_FILES)
	$(CXX) $(CXX_FLAGS) $(CXX_PERF_FLAGS) $(INCLUDE_DIRS) $(LIBRARY_DIRS) $^ $(STATIC_LIBS) -o $@ $(LIBRARIES)

$(OBJ)/%.o: $(SRC)/%.cpp Makefile
	$(CXX) $(CXX_FLAGS) $(CXX_PERF_FLAGS) -c -MMD -MP -DNDEBUG $(INCLUDE_DIRS) $(LIBRARY_DIRS) $< -o $@ $(LIBRARIES)

### Debug ###

$(BIN)/$(EXE_DBG): $(SRC_OBJ_FILES)
	$(CXX) $(CXX_FLAGS) $(CXX_DBG_FLAGS) $(INCLUDE_DIRS) $(LIBRARY_DIRS) $^ $(STATIC_LIBS) -o $@ $(LIBRARIES)

$(OBJ)/%.dbg.o: $(SRC)/%.cpp Makefile
	$(CXX) $(CXX_FLAGS) $(CXX_DBG_FLAGS) -c -MMD -MP $(INCLUDE_DIRS) $(LIBRARY_DIRS) $< -o $@ $(LIBRARIES)

### Test ###

$(BIN)/$(EXE_TEST): $(TEST_OBJ_FILES)
	$(CXX) $(CXX_FLAGS) $(CXX_DBG_FLAGS) $(INCLUDE_DIRS) $(LIBRARY_DIRS) $^ $(STATIC_LIBS) -o $@ $(LIBRARIES)

$(OBJ)/%.test.o: $(SRC)/%.cpp Makefile
	$(CXX) $(CXX_FLAGS) $(CXX_DBG_FLAGS) -c -MMD -MP -DTEST_BUILD $(INCLUDE_DIRS) $(LIBRARY_DIRS) $< -o $@ $(LIBRARIES)

### Benchmark ###

$(BIN)/$(EXE_BENCH): $(BENCH_OBJ_FILES)
	$(CXX) $(CXX_FLAGS) $(CXX_PERF_FLAGS) $(INCLUDE_DIRS) $(LIBRARY_DIRS) $^ $(STATIC_LIBS) -o $@ $(LIBRARIES)

$(OBJ)/%.bench.o: $(SRC)/%.cpp Makefile
	$(CXX) $(CXX_FLAGS) $(CXX_PERF_FLAGS) -c -MMD -MP -DNDEBUG -DBENCH_BUILD $(INCLUDE_DIRS) $(LIBRARY_DIRS) $< -o $@ $(LIBRARIES)

############
### Dirs ###
############

$(BIN):
	@echo "Creating bin directory"
	-@mkdir $@

$(OBJ):
	@echo "Creating obj directory"
	-@mkdir $@

$(OBJ_DIRS):
	@echo "creating obj directory: $@"
	-@mkdir -p $@

$(DIST_DIR)/$(EXE): $(BIN)/$(EXE)
	@cp $< $@
