#include "includes.h"
#include "bench.hpp"

util::Clock Timer;

namespace
{
  struct
  {
    std::unique_ptr<std::deque<bench::Benchmark*>> benchmarks;
  } Globals;

}  // namespace

namespace bench
{
  Benchmark::Benchmark(const char* name, bool enabled): benchmark_name(name), enabled(enabled)
  {
    // to prevent static initialization fiasco
    if (Globals.benchmarks == nullptr) {
      Globals.benchmarks = std::make_unique<std::deque<Benchmark*>>();
    }
    Globals.benchmarks->push_back(this);
  }

  void Benchmark::run()
  {
    std::cout << "Benchmark Count: " << Globals.benchmarks->size() << '\n';

    for (auto benchmark : *Globals.benchmarks) {
      if (benchmark->enabled) {
        std::cout << BENCH_BREAK << "Running '\x1b[35m" << benchmark->benchmark_name << "\x1b[m'\n";
        benchmark->body();
      }
    }
    std::cout << BENCH_BREAK;
  }
}  // namespace benchmarking
