CXX			:= g++
CXX_FLAGS 		:= -Wall -Wextra -std=c++17

SRC			:= src
INCLUDE			:= include
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

rwildcard=$(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))

SRC_FILES	:= $(call rwildcard,$(SRC),*.cpp)
SRC_OBJ_FILES	:= $(patsubst $(SRC)/%.cpp, $(OBJ)/%.o, $(SRC_FILES))

all: make_dirs $(BIN)/$(EXE) $(BIN)/$(EXE_SPEC)

$(BIN)/$(EXE): $(SRC_OBJ_FILES)
	$(CXX) $(CXX_FLAGS) $(INCLUDE_DIRS) $(LIBRARY_DIRS) $^ $(STATIC_LIBS) -o $@ $(LIBRARIES)

$(OBJ)/%.o: $(SRC)/%.cpp
	$(CXX) $(CXX_FLAGS) -c $(INCLUDE_DIRS) $(LIBRARY_DIRS) $< -o $@ $(LIBRARIES)

make_dirs:
	@mkdir -p $(BIN) $(OBJ) $(OBJ_DIRS)

clean:
	-rm $(BIN)/* $(OBJ)/*

FORCE:
