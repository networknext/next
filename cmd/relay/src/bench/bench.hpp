#ifndef BENCH_BENCH_HPP
#define BENCH_BENCH_HPP

#include <deque>
#include <iostream>

#include "clock.hpp"

#define Bench(benchmark_name)                                         \
    class _bench_##benchmark_name##_ : public benchmarking::Benchmark \
    {                                                                 \
       public:                                                        \
        _bench_##benchmark_name##_() : benchmarking::Benchmark(#benchmark_name) \
        {}                                                            \
        void body() override;                                         \
    };                                                                \
    _bench_##benchmark_name##_ _var_##benchmark_name##_;              \
    void _bench_##benchmark_name##_::body()

extern benchmarking::Clock Timer;

namespace benchmarking
{
    class Benchmark
    {
       public:
        static void Run();

        const char* BenchmarkName;

       protected:
        Benchmark(const char* name);

        virtual void body() = 0;

       private:
        static std::deque<Benchmark*> mBenchmarks;
    };
}  // namespace benchmarking
#endif