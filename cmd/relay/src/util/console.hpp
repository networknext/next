#pragma once
#include <iostream>
#include <string>
#include <ctime>
#include <strings.h>
#include <mutex>

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

	private:
		std::ostream& mStream;
		std::mutex mLock;

		static std::string StrTime();
	};

	inline Console::Console(std::ostream& stream): mStream(stream) {}

	template <typename... Args>
	inline void Console::write(Args&&... args)
	{
		mLock.lock();
		((mStream << std::forward<Args>(args)), ...);
		mLock.unlock();
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

	inline std::string Console::StrTime()
	{
		std::array<char, 16> timebuff;
		auto t = time(nullptr);
		auto timestruct = localtime(&t);
		auto count = std::strftime(timebuff.data(), timebuff.size() * sizeof(char) - 1, "%I:%M:%S %P", timestruct);
		return std::string(timebuff.begin(), timebuff.begin() + count);
	}
}  // namespace util
