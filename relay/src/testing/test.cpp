#include "includes.h"
#include "test.hpp"

namespace
{
  struct
  {
    std::unique_ptr<std::deque<testing::SpecTest*>> tests;
  }  // namespace struct
  Globals;
}  // namespace

namespace testing
{
  SpecTest::SpecTest(const char* name, bool disabled): TestName(name), Disabled(disabled)
  {
    if (Globals.tests == nullptr) {
      Globals.tests = std::make_unique<std::deque<testing::SpecTest*>>();
    }

    Globals.tests->push_back(this);
  }

  bool SpecTest::Run(int argc, const char* argv[])
  {
    if (argc > 1) {
      Globals.tests->erase(
       std::remove_if(
        Globals.tests->begin(),
        Globals.tests->end(),
        [argc, argv](auto test) -> bool {
          for (int i = 1; i < argc; i++) {
            if (strstr(test->TestName, argv[i]) != nullptr) {
              return false;
            }
          }

          return true;
        }),
       Globals.tests->end());
    }

    std::sort(Globals.tests->begin(), Globals.tests->end(), [](testing::SpecTest* a, testing::SpecTest* b) -> bool {
      auto upcase = [](std::string& str) {
        for (char& c : str) {
          c &= 0xDF;
        }
      };

      std::string a_name = a->TestName;
      std::string b_name = b->TestName;

      upcase(a_name);
      upcase(b_name);

      return a_name.compare(b_name) > 0;
    });

    std::cout << "Test count: " << Globals.tests->size() << '\n';

    bool no_tests_skipped = true;
    for (auto test : *Globals.tests) {
      if (!test->Disabled) {
        std::cout << TEST_BREAK << "Running test '\x1b[35m" << test->TestName << "\x1b[m'\n";
        test->body();
      } else {
        std::cout << TEST_BREAK_WARNING << "Skipping test '\x1b[36m" << test->TestName << "\x1b[m'\n";
        no_tests_skipped = false;
      }
    }
    std::cout << TEST_BREAK;
    return no_tests_skipped;
  }

}  // namespace testing
