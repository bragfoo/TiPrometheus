# Copyright 2021 The TiPrometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

include Makefile.common

.PHONY: docker
docker: npm_licenses common-docker

.PHONY: build
build: assets common-build

.PHONY: bench_tsdb
bench_tsdb: $(PROMU)
	@echo ">> building promtool"
	@GO111MODULE=$(GO111MODULE) $(PROMU) build --prefix $(PREFIX) promtool
	@echo ">> running benchmark, writing result to $(TSDB_BENCHMARK_OUTPUT_DIR)"
	@$(PROMTOOL) tsdb bench write --metrics=$(TSDB_BENCHMARK_NUM_METRICS) --out=$(TSDB_BENCHMARK_OUTPUT_DIR) $(TSDB_BENCHMARK_DATASET)
	@$(GO) tool pprof -svg $(PROMTOOL) $(TSDB_BENCHMARK_OUTPUT_DIR)/cpu.prof > $(TSDB_BENCHMARK_OUTPUT_DIR)/cpuprof.svg
	@$(GO) tool pprof --inuse_space -svg $(PROMTOOL) $(TSDB_BENCHMARK_OUTPUT_DIR)/mem.prof > $(TSDB_BENCHMARK_OUTPUT_DIR)/memprof.inuse.svg
	@$(GO) tool pprof --alloc_space -svg $(PROMTOOL) $(TSDB_BENCHMARK_OUTPUT_DIR)/mem.prof > $(TSDB_BENCHMARK_OUTPUT_DIR)/memprof.alloc.svg
	@$(GO) tool pprof -svg $(PROMTOOL) $(TSDB_BENCHMARK_OUTPUT_DIR)/block.prof > $(TSDB_BENCHMARK_OUTPUT_DIR)/blockprof.svg
	@$(GO) tool pprof -svg $(PROMTOOL) $(TSDB_BENCHMARK_OUTPUT_DIR)/mutex.prof > $(TSDB_BENCHMARK_OUTPUT_DIR)/mutexprof.svg
