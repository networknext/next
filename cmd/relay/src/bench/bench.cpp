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
                  << "\n";

        for (auto benchmark : mBenchmarks) {
            std::cout << "\nRunning '\x1b[35m" << benchmark->BenchmarkName << "\x1b[m'\n\n";
            benchmark->body();
        }
    }
}  // namespace benchmarking