#include "test.hpp"
#include <iostream>
#include <memory>

namespace
{
  bool gTestInit = false;
  std::unique_ptr<std::deque<testing::SpecTest*>> gTests;
  ;
}  // namespace

namespace testing
{
  SpecTest::SpecTest(const char* name, bool disabled): TestName(name), Disabled(disabled)
  {
    if (!gTestInit) {
      gTests = std::make_unique<std::deque<testing::SpecTest*>>();
      gTestInit = true;
    }

    gTests->push_back(this);
  }

  bool SpecTest::Run()
  {
    std::cout << "Test count: " << gTests->size() << '\n';

    bool noTestsSkipped = true;
    for (auto test : *gTests) {
      if (!test->Disabled) {
        std::cout << TEST_BREAK << "Running test '\x1b[35m" << test->TestName << "\x1b[m'\n";
        test->body();
      } else {
        std::cout << TEST_BREAK_WARNING << "Skipping test '\x1b[36m" << test->TestName << "\x1b[m'\n";
        noTestsSkipped = false;
      }
    }
    std::cout << TEST_BREAK;
    return noTestsSkipped;
  }
}  // namespace testing