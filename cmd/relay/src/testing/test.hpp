#ifndef TESTING_TEST_HPP
#define TESTING_TEST_HPP

#include <deque>

#define TEST_CLASS_CREATOR(test_name, disabled)     \
	class _test_##test_name##_: public testing::Test   \
	{                                                  \
	public:                                            \
		_test_##test_name##_(): testing::Test(#test_name) \
	};
namespace testing
{
	class Test
	{
	public:
		static void Run();

		const char* TestName;
		const bool Disabled;

	protected:
		Test(const char* name, bool disabled);

		virtual void body() = 0;

	private:
		static std::deque<Test*> mTests;
	};

	void relay_test();
}  // namespace testing
#endif