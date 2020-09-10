#pragma once

#include "macros.hpp"

namespace util
{
  class Console
  {
   public:
    Console(std::ostream& stream = std::cout);

    template <typename... Args>
    void write(Args&&... args);

    template <typename... Args>
    void write_line(Args&&... args);

    template <typename... Args>
    void log(Args&&... args);

    void flush();

   private:
    std::ostream& stream;
    std::mutex lock;

    static std::string stringify_time();
  };

  INLINE Console::Console(std::ostream& stream): stream(stream) {}

  template <typename... Args>
  INLINE void Console::write(Args&&... args)
  {
    std::lock_guard<std::mutex> lk(this->lock);
    ((this->stream << std::forward<Args>(args)), ...);
  }

  template <typename... Args>
  INLINE void Console::write_line(Args&&... args)
  {
    write(args..., '\n');
  }

  template <typename... Args>
  INLINE void Console::log(Args&&... args)
  {
    write_line('[', stringify_time(), "] ", args...);
  }

  INLINE void Console::flush()
  {
    this->stream.flush();
  }

  INLINE std::string Console::stringify_time()
  {
    std::array<char, 16> timebuff;
    auto t = time(nullptr);
    auto timestruct = localtime(&t);
    auto count = std::strftime(timebuff.data(), timebuff.size() * sizeof(char) - 1, "%I:%M:%S %P", timestruct);
    //auto count = std::strftime(timebuff.data(), timebuff.size() * sizeof(char) - 1, "%s", timestruct);
    return std::string(timebuff.begin(), timebuff.begin() + count);
  }
}  // namespace util
