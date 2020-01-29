#include "bench.hpp"

benchmarking::Clock Timer;

namespace benchmarking
{
    // to help prevent static initialization fiasco
    std::deque<Benchmark*> Benchmark::mBenchmarks = []() -> std::deque<Benchmark*> { return std::deque<Benchmark*>(); }();

    Benchmark::Benchmark(const char* name) : BenchmarkName(name)
    {
        mBenchmarks.push_back(this);
    }

    void Benchmark::Run()
    {
        std::cout << "Running " << mBenchmarks.size() << " benchmarks"
                  << "\n\n";

        for (auto benchmark : mBenchmarks) {
            std::cout << "Running '" << benchmark->BenchmarkName << "'\n\n";
            benchmark->body();
        }
    }
}  // namespace benchmarking