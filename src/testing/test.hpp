#ifndef TESTING_TEST_HPP
#define TESTING_TEST_HPP

#include "net/address.hpp"

#define TEST_BREAK "\n=============================================\n\n"
#define TEST_BREAK_WARNING "\n!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n\n"

#define TEST_CLASS_CREATOR(test_name, disabled)                          \
  namespace testing                                                      \
  {                                                                      \
    class _test_##test_name##_: public testing::SpecTest                 \
    {                                                                    \
     public:                                                             \
      _test_##test_name##_(): testing::SpecTest(#test_name, disabled) {} \
      void body() override;                                              \
    };                                                                   \
    _test_##test_name##_ _test_var_##test_name##_;                       \
  }                                                                      \
  void testing::_test_##test_name##_::body()

#define TEST_CLASS_CREATOR_1_ARG(test_name) TEST_CLASS_CREATOR(test_name, false)
#define TEST_CLASS_CREATOR_2_ARG(test_name, disabled) TEST_CLASS_CREATOR(test_name, disabled)

#define GET_3RD_TEST_ARG(arg1, arg2, arg3, ...) arg3
#define TEST_MACRO_CHOOSER(...) GET_3RD_TEST_ARG(__VA_ARGS__, TEST_CLASS_CREATOR_2_ARG, TEST_CLASS_CREATOR_1_ARG)

/*
 * Test macro. Takes two parameters, the second being optional when developing, required to be false when finishing
 * The first is the name of the test to run. It must be unique across the codebase however it can be the same name as a
 * benchmark.
 * The second is whether to disable it. If any one test is disabled regardless of if the others pass, then the program will
 * exit with an error. So all written tests must pass.
 *
 * The above macros result in the creation of a class with the name being the name of the test prefixed by "_test_" and
 * postfixed with a single '_' Because of that you can test private functions of regular classes in the code base.
 *
 * To do so first forward declare the complete name of the test (with the pre & postfix) within the testing namespace.
 * Then simply use the friend keyword within the class you'd like to test the private functions of.
 */
#define Test(...) TEST_MACRO_CHOOSER(__VA_ARGS__)(__VA_ARGS__)

#define check(cond) testing::CheckHandler((cond), #cond, __FUNCTION__, __FILE__, __LINE__)

namespace testing
{
  class SpecTest
  {
   public:
    static bool Run(int argc, const char* argv[]);

    const char* TestName;
    const bool Disabled;

    virtual void body() = 0;

   protected:
    SpecTest(const char* name, bool disabled);
  };

  class SpecCheck
  {
   public:
    SpecCheck(bool result, const char* condition, const char* function, const char* file, int line);
    ~SpecCheck();

    void onFail(std::function<void(void)> failFunc);

   private:
    bool mResult;
    const char* mCondition;
    const char* mFunction;
    const char* mFile;
    const int mLine;
    std::function<void(void)> mOnFail;
  };

  inline SpecCheck::SpecCheck(bool result, const char* condition, const char* function, const char* file, int line)
   : mResult(result), mCondition(condition), mFunction(function), mFile(file), mLine(line)
  {}

  inline SpecCheck::~SpecCheck()
  {
    if (!mResult) {
      std::cout << "check failed: ( " << mCondition << " ), function " << mFunction << ", file " << mFile << ", line  " << mLine
                << std::endl;
      if (mOnFail) {
        mOnFail();
      }
#if defined(__GNUC__)
      __builtin_trap();
#elif defined(_MSC_VER)
      __debugbreak();
#endif
      std::exit(1);
    }
  }

  inline void SpecCheck::onFail(std::function<void(void)> failFunc)
  {
    mOnFail = failFunc;
  }

  template <typename T>
  SpecCheck CheckHandler(T result, const char* condition, const char* function, const char* file, int line);

  template <>
  inline SpecCheck CheckHandler(bool result, const char* condition, const char* function, const char* file, int line)
  {
    return SpecCheck(result, condition, function, file, line);
  }

  template <typename T>
  std::enable_if_t<std::is_floating_point<T>::value, T> RandomDecimal();

  net::Address RandomAddress();

  // slow but easy to write, use for tests only
  // valid return types are std string/vector
  template <class ReturnType>
  ReturnType ReadFile(std::string filename);

  template <typename T>
  std::enable_if_t<std::is_floating_point<T>::value, T> RandomDecimal()
  {
    static auto rand = std::bind(std::uniform_real_distribution<T>(), std::default_random_engine());
    return static_cast<T>(rand());
  }

  template <class ReturnType>
  ReturnType ReadFile(std::string filename)
  {
    ReturnType retval;

    if (!filename.empty()) {
      std::ifstream stream;

      stream.open(filename, std::ios::binary);

      if (stream) {
        std::stringstream data;
        data << stream.rdbuf();
        auto str = data.str();
        retval.assign(str.begin(), str.end());
      }

      stream.close();
    }

    return retval;
  }
}  // namespace testing
#endif
