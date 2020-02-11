#ifndef BENCH_BENCH_HPP
#define BENCH_BENCH_HPP

#include "util/clock.hpp"

#define BENCH_BREAK "\n=============================================\n\n"

#define BENCHMARK_CLASS_CREATOR(benchmark_name, enabled)                               \
  class _bench_##benchmark_name##_: public benchmarking::Benchmark                     \
  {                                                                                    \
   public:                                                                             \
    _bench_##benchmark_name##_(): benchmarking::Benchmark(#benchmark_name, enabled) {} \
    void body() override;                                                              \
  };                                                                                   \
  _bench_##benchmark_name##_ _bench_var_##benchmark_name##_;                                 \
  void _bench_##benchmark_name##_::body()

#define BENCHMARK_CLASS_CREATOR_1_ARG(benchmark_name) BENCHMARK_CLASS_CREATOR(benchmark_name, false)
#define BENCHMARK_CLASS_CREATOR_2_ARG(benchmark_name, enabled) BENCHMARK_CLASS_CREATOR(benchmark_name, enabled)

#define GET_3RD_BENCH_ARG(arg1, arg2, arg3, ...) arg3
#define BENCHMARK_MACRO_CHOOSER(...) GET_3RD_BENCH_ARG(__VA_ARGS__, BENCHMARK_CLASS_CREATOR_2_ARG, BENCHMARK_CLASS_CREATOR_1_ARG)

/*
    Benchmark macro. Takes two parameters, and with preprocessor magic the second is optional

    The first parameter is the name of the benchmark to run. It must be unique across the codebase since it is transformed into
    a class. However it can be the same name as a test.

    The second is wheter to enable it. False by default because there will likely be more benchmarks than desired
*/

#define Bench(...) BENCHMARK_MACRO_CHOOSER(__VA_ARGS__)(__VA_ARGS__)

#define Do(times) \
  Timer.reset();  \
  for (size_t i = 0; i < (times); i++)

// Just for readability
#define Skip()                                          \
  std::cout << BENCH_BREAK;                             \
  std::cout << "Skipping the rest of this benchmark\n"; \
  return

extern util::Clock Timer;

namespace benchmarking
{
  class Benchmark
  {
   public:
    static void Run();

    const char* BenchmarkName;
    const bool Enabled;

   protected:
    Benchmark(const char* name, bool enabled);

    virtual void body() = 0;
  };
}  // namespace benchmarking
#endif
