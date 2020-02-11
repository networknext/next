#pragma once

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
    Clock() = default;

    /* Timestamps the clock */
    void reset();

    /* Get how much time as elasped since starting */
    template <typename UnitOfTime>
    auto elapsed() const -> double;

    /* Check if a time duration has passed */
    template <typename UnitOfTime>
    inline auto elapsed(double value) const -> bool;

   private:
    Instant mNow;
    size_t mDelta;

    template <typename T>
    inline auto diff() const -> double;
  };

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

  inline void Clock::reset()
  {
    mNow = InternalClock::now();
  }

  template <typename UnitOfTime>
  inline auto Clock::elapsed(double value) const -> bool
  {
    return std::chrono::duration_cast<UnitOfTime>(InternalClock::now() - mNow).count() >= value;
  }

  template <typename T>
  inline auto Clock::diff() const -> double
  {
    return std::chrono::duration<double, T>(InternalClock::now() - mNow).count();
  }
}  // namespace util
