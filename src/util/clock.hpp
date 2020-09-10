#pragma once

namespace util
{
#if defined _WIN32
  using Instant = std::chrono::steady_clock::time_point;
#elif defined __linux__
  using Instant = std::chrono::high_resolution_clock::time_point;
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

    /* Timestamps the clock */
    void reset();

    /* Get how much time as elapsed since starting */
    template <typename U>
    double elapsed() const;

    /* Check if a time duration has passed */
    template <typename U>
    bool elapsed(double value) const;

    /* Returns the number of seconds since the epoch */
    template <typename U>
    double unix_time() const;

   private:
    Instant now;
    size_t delta;

    template <typename T>
    double diff() const;
  };

  inline Clock::Clock()
  {
    reset();
  }

  inline void Clock::reset()
  {
    this->now = InternalClock::now();
  }

  template <>
  inline double Clock::elapsed<Nanosecond>() const
  {
    return diff<std::nano>();
  }

  template <>
  inline double Clock::elapsed<Microsecond>() const
  {
    return diff<std::micro>();
  }

  template <>
  inline double Clock::elapsed<Millisecond>() const
  {
    return diff<std::milli>();
  }

  template <>
  inline double Clock::elapsed<Second>() const
  {
    return diff<std::ratio<1>>();
  }

  template <typename U>
  inline bool Clock::elapsed(double value) const
  {
    return std::chrono::duration_cast<U>(InternalClock::now() - this->now).count() >= value;
  }

  template <typename T>
  inline double Clock::unix_time() const
  {
    const auto seconds = std::chrono::duration_cast<T>(InternalClock::now().time_since_epoch());
    return seconds.count();
  }

  template <typename T>
  inline double Clock::diff() const
  {
    return std::chrono::duration<double, T>(InternalClock::now() - this->now).count();
  }
}  // namespace util
