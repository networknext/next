#pragma once

#include "util/macros.hpp"

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
#define TEST(...) TEST_MACRO_CHOOSER(__VA_ARGS__)(__VA_ARGS__)

#define CHECK(cond) testing::CheckHandler((cond), #cond, __FILE__, __LINE__)

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
    SpecCheck(bool result, const char* condition, const char* file, int line);
    ~SpecCheck();

    void on_fail(std::function<void(void)> fail_func);

   private:
    bool result;
    const char* condition;
    const char* file;
    const int line;
    std::function<void(void)> on_fail_func;
  };

  INLINE SpecCheck::SpecCheck(bool result, const char* condition, const char* file, int line)
   : result(result), condition(condition), file(file), line(line)
  {}

  INLINE SpecCheck::~SpecCheck()
  {
    if (!this->result) {
      std::cout << "check failed: (" << this->condition << ")"
                << ", file " << this->file << " (" << this->line << ")\n";
      if (this->on_fail_func) {
        this->on_fail_func();
      }
      std::cout << std::flush;
      std::exit(1);
    }
  }

  INLINE void SpecCheck::on_fail(std::function<void(void)> fail_func)
  {
    this->on_fail_func = fail_func;
  }

  template <typename T>
  INLINE SpecCheck CheckHandler(T result, const char* condition, const char* file, int line);

  template <>
  INLINE SpecCheck CheckHandler(bool result, const char* condition, const char* file, int line)
  {
    return SpecCheck(result, condition, file, line);
  }
}  // namespace testing
