#pragma once

/*
  Uses of this class:
  - As the relay system clock, defined by the variable in main "relayClock"
 */

namespace util
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
    Clock();
    ~Clock() = default;

    /* Timestamps the clock */
    void reset();

    /* Get how much time as elasped since starting */
    template <typename UnitOfTime>
    auto elapsed() const -> double;

   private:
    Instant mTimestamp;
    size_t mDelta;

    template <typename T>
    inline auto diff() const -> double;
  };

  inline Clock::Clock()
  {
    reset();
  }

  inline void Clock::reset()
  {
    mTimestamp = InternalClock::now();
  }

  template <>
  inline auto Clock::elapsed<Nanosecond>() const -> double
  {
    return diff<std::nano>();
  }

  template <>
  inline auto Clock::elapsed<Microsecond>() const -> double
  {
    return diff<std::micro>();
  }

  template <>
  inline auto Clock::elapsed<Millisecond>() const -> double
  {
    return diff<std::milli>();
  }

  template <>
  inline auto Clock::elapsed<Second>() const -> double
  {
    return diff<std::ratio<1>>();
  }

  template <typename T>
  inline auto Clock::diff() const -> double
  {
    return std::chrono::duration<double, T>(InternalClock::now() - mTimestamp).count();
  }
}  // namespace util
