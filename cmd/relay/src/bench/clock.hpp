#pragma once
#include <chrono>

namespace benchmarking
{
#if defined _WIN32
    using Instant = std::chrono::steady_clock::time_point;
#elif defined __linux__
    using Instant = std::chrono::system_clock::time_point;
#endif

    using InternalClock = std::chrono::high_resolution_clock;
    using Nanosecond = std::chrono::nanoseconds;
    using Microsecond = std::chrono::microseconds;
    using Millisecond = std::chrono::milliseconds;
    using Second = std::chrono::seconds;

    class Clock
    {
       public:
        Clock() = default;

        /* Timestamps the clock */
        void reset();

        /* Get how much time as elasped since starting */
        template <typename UnitOfTime>
        double elapsed();

        /* Check if a time duration has passed */
        template <typename UnitOfTime>
        bool elapsed(double value)
        {
            return std::chrono::duration_cast<UnitOfTime>(InternalClock::now() - mNow).count() >= value;
        }

       private:
        Instant mNow;
        size_t mDelta;

        template <typename T>
        double diff()
        {
            return std::chrono::duration<double, T>(InternalClock::now() - mNow).count();
        }
    };

    template <>
    inline double Clock::elapsed<Nanosecond>()
    {
        return diff<std::nano>();
    }

    template <>
    inline double Clock::elapsed<Microsecond>()
    {
        return diff<std::micro>();
    }

    template <>
    inline double Clock::elapsed<Millisecond>()
    {
        return diff<std::milli>();
    }

    template <>
    inline double Clock::elapsed<Second>()
    {
        return diff<std::ratio<1>>();
    }

    inline void Clock::reset()
    {
        mNow = InternalClock::now();
    }
}  // namespace epoch
