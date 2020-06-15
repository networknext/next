#include "includes.h"
#include "bench.hpp"

util::Clock Timer;

namespace
{
  bool gBenchmarkInit = false;
  std::unique_ptr<std::deque<benchmarking::Benchmark*>> gBenchmarks;
}  // namespace

namespace benchmarking
{
  Benchmark::Benchmark(const char* name, bool enabled): BenchmarkName(name), Enabled(enabled)
  {
    // to prevent static initialization fiasco
    if (!gBenchmarkInit) {
      gBenchmarks = std::make_unique<std::deque<Benchmark*>>();
      gBenchmarkInit = true;
    }
    gBenchmarks->push_back(this);
  }

  void Benchmark::Run()
  {
    std::cout << "Benchmark Count: " << gBenchmarks->size() << '\n';

    for (auto benchmark : *gBenchmarks) {
      if (benchmark->Enabled) {
        std::cout << BENCH_BREAK << "Running '\x1b[35m" << benchmark->BenchmarkName << "\x1b[m'\n";
        benchmark->body();
      }
    }
    std::cout << BENCH_BREAK;
  }
}  // namespace benchmarking
