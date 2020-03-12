#ifndef TESTING_TEST_HPP
#define TESTING_TEST_HPP

#include "net/address.hpp"

#define TEST_BREAK "\n=============================================\n\n"
#define TEST_BREAK_WARNING "\n!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n\n"

#define TEST_CLASS_CREATOR(test_name, disabled)                        \
  class _test_##test_name##_: public testing::SpecTest                 \
  {                                                                    \
   public:                                                             \
    _test_##test_name##_(): testing::SpecTest(#test_name, disabled) {} \
    void body() override;                                              \
  };                                                                   \
  _test_##test_name##_ _test_var_##test_name##_;                       \
  void _test_##test_name##_::body()

#define TEST_CLASS_CREATOR_1_ARG(test_name) TEST_CLASS_CREATOR(test_name, false)
#define TEST_CLASS_CREATOR_2_ARG(test_name, disabled) TEST_CLASS_CREATOR(test_name, disabled)

#define GET_3RD_TEST_ARG(arg1, arg2, arg3, ...) arg3
#define TEST_MACRO_CHOOSER(...) GET_3RD_TEST_ARG(__VA_ARGS__, TEST_CLASS_CREATOR_2_ARG, TEST_CLASS_CREATOR_1_ARG)

/*
    Test macro. Takes two parameters, the second being optional when developing, required to be false when finishing

    THe first is the name of the test to run. It must be unique across the codebase however it can be the same name as a
   benchmark.

    The second is whether to disable it. If any one test is disabled regardless of if the others pass, then the program will
   exit with an error. So all written tests must pass.
*/

#define Test(...) TEST_MACRO_CHOOSER(__VA_ARGS__)(__VA_ARGS__)

#define check(condition) \
  testing::check_handler((condition), #condition, (const char*)__FUNCTION__, (const char*)__FILE__, __LINE__);

namespace testing
{
  class SpecTest
  {
   public:
    static bool Run();

    const char* TestName;
    const bool Disabled;

   protected:
    SpecTest(const char* name, bool disabled);

    virtual void body() = 0;
  };

  template <typename T>
  void check_handler(T result, const char* condition, const char* function, const char* file, int line);

  template <>
  inline void check_handler(bool result, const char* condition, const char* function, const char* file, int line)
  {
    if (!result) {
      printf("check failed: ( %s ), function %s, file %s, line %d\n", condition, function, file, line);
      fflush(stdout);
#ifndef NDEBUG
#if defined(__GNUC__)
      __builtin_trap();
#elif defined(_MSC_VER)
      __debugbreak();
#endif
#endif
      exit(1);
    }
  }

  template <typename T>
  std::enable_if_t<std::numeric_limits<T>::is_integer, T> Random();

  template <typename T>
  std::enable_if_t<std::is_floating_point<T>::value, T> RandomDecimal();

  net::Address RandomAddress();

  // slow but easy to write, use for tests only
  // valid return types are std string/vector
  template <class ReturnType>
  ReturnType ReadFile(std::string filename);

  template <typename T>
  std::enable_if_t<std::numeric_limits<T>::is_integer, T> Random()
  {
    static auto rand = std::bind(std::uniform_int_distribution<T>(), std::default_random_engine());
    return static_cast<T>(rand());
  }

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
