#pragma once

namespace util
{
  class Console
  {
   public:
    Console(std::ostream& stream = std::cout);

    template <typename... Args>
    void write(Args&&... args);

    template <typename... Args>
    void writeLine(Args&&... args);

    template <typename... Args>
    void log(Args&&... args);

    void flush();

   private:
    std::ostream& mStream;
    std::mutex mLock;

    static std::string StrTime();
  };

  inline Console::Console(std::ostream& stream): mStream(stream) {}

  template <typename... Args>
  inline void Console::write(Args&&... args)
  {
    std::lock_guard<std::mutex> lk(mLock); // prevents log messages from merging into unreadable messes
    ((mStream << std::forward<Args>(args)), ...);
  }

  template <typename... Args>
  inline void Console::writeLine(Args&&... args)
  {
    write(args..., '\n');
  }

  template <typename... Args>
  inline void Console::log(Args&&... args)
  {
    writeLine('[', StrTime(), "] ", args...);
  }

  inline void Console::flush()
  {
    mStream.flush();
  }

  inline std::string Console::StrTime()
  {
    std::array<char, 16> timebuff;
    auto t = time(nullptr);
    auto timestruct = localtime(&t);
    auto count = std::strftime(timebuff.data(), timebuff.size() * sizeof(char) - 1, "%I:%M:%S %P", timestruct);
    //auto count = std::strftime(timebuff.data(), timebuff.size() * sizeof(char) - 1, "%s", timestruct);
    return std::string(timebuff.begin(), timebuff.begin() + count);
  }
}  // namespace util
