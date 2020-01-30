#################
### Functions ###
#################

rwildcard=$(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))

#################
### Variables ###
#################

CXX			:= g++
CXX_FLAGS 		:= -Wall -Wextra -std=c++17 -O3 -march=native -frename-registers -funroll-loops -fopenmp -D_GLIBCXX_PARALLEL

SRC			:= src
INCLUDE			:= $(SRC) # include path is the same as the src dir
SPEC			:= spec

INCLUDE_DIRS		:= -I$(INCLUDE)
STATIC_LIBS		:=
LIBRARY_DIRS		:= -L.
LIBRARIES		:= -lsodium -lcurl -lpthread -lm

EXE			:= relay

BIN			:= bin
OBJ			:= obj

SUB_DIRS		:= $(shell find src/* -type d -print)
OBJ_DIRS		:= $(patsubst $(SRC)/%, $(OBJ)/%, $(SUB_DIRS))

SRC_FILES	:= $(call rwildcard,$(SRC),*.cpp)
SRC_OBJ_FILES	:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.o, $(SRC_FILES))

DIST_DIR	:= ../../dist

################
### Targets  ###
################

all: $(BIN) $(OBJ) $(OBJ_DIRS) $(BIN)/$(EXE) $(DIST_DIR)/$(EXE)

clean:
	-rm -rf $(BIN) $(OBJ) $(DIST_DIR)/$(EXE)

force: clean all

run-benchmarks: all
	$(BIN)/$(EXE) bench

$(BIN)/$(EXE): $(SRC_OBJ_FILES)
	$(CXX) $(CXX_FLAGS) $(INCLUDE_DIRS) $(LIBRARY_DIRS) $^ $(STATIC_LIBS) -o $@ $(LIBRARIES)

$(OBJ)/%.o: $(SRC)/%.cpp
	$(CXX) $(CXX_FLAGS) -c $(INCLUDE_DIRS) $(LIBRARY_DIRS) $< -o $@ $(LIBRARIES)

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

