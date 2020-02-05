#include "bench.hpp"

benchmarking::Clock Timer;

namespace benchmarking
{
  // to help prevent static initialization fiasco
  std::deque<Benchmark*> Benchmark::mBenchmarks = []() -> std::deque<Benchmark*> {
    return std::deque<Benchmark*>();
  }();

  Benchmark::Benchmark(const char* name, bool enabled): BenchmarkName(name), Enabled(enabled)
  {
    mBenchmarks.push_back(this);
  }

  void Benchmark::Run()
  {
    for (auto benchmark : mBenchmarks) {
      if (benchmark->Enabled) {
        std::cout << BENCH_BREAK << "Running '\x1b[35m" << benchmark->BenchmarkName << "\x1b[m'\n\n";
        benchmark->body();
      }
    }
    std::cout << BENCH_BREAK;
  }
}  // namespace benchmarking