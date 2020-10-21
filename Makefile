MAKEFILE_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

include $(MAKEFILE_DIR)/Makefile.common

################
### Targets  ###
################

.PHONY: release
release: $(BIN) $(OBJ)
	$(MAKE) -f $(RELEASE_MAKEFILE)

.PHONY: debug
debug: $(BIN) $(OBJ)
	$(MAKE) -f $(DEBUG_MAKEFILE)

.PHONY: test
test: $(BIN) $(OBJ)
	$(MAKE) -f $(TEST_MAKEFILE)

.PHONY: benchmark
benchmark: $(BIN) $(OBJ)
	$(MAKE) -f $(BENCHMARK_MAKEFILE)

.PHONY: run-release
run-release: release
	$(MAKE) -f $(RELEASE_MAKEFILE) run

.PHONY: run-debug
run-debug: debug
	$(MAKE) -f $(DEBUG_MAKEFILE) run

.PHONY: run-tests
run-tests: test
	$(MAKE) -f $(TEST_MAKEFILE) run

.PHONY: run-benchmarks
run-benchmarks: benchmark
	$(MAKE) -f $(BENCHMARK_MAKEFILE) run

.PHONY: clean
clean:
	$(MAKE) -f $(RELEASE_MAKEFILE) 		clean
	$(MAKE) -f $(DEBUG_MAKEFILE) 			clean
	$(MAKE) -f $(TEST_MAKEFILE) 			clean
	$(MAKE) -f $(BENCHMARK_MAKEFILE) 	clean

.PHONY: reset-pch
reset-pch:
	rm $(GCH)

############
### Dirs ###
############

$(BIN):
	@echo "Creating bin directory"
	-@mkdir $@

$(OBJ): $(SRC_OBJ_DIRS) $(LIB_OBJ_DIRS)

$(SRC_OBJ_DIRS):
	@echo "creating obj directory: $@"
	-@mkdir -p $@

$(LIB_OBJ_DIRS):
	@echo "creating obj directory: $@"
	-@mkdir -p $@
