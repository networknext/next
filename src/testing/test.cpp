#include "includes.h"
#include "test.hpp"

#include "crypto/bytes.hpp"

namespace
{
  bool gTestInit = false;
  std::unique_ptr<std::deque<testing::SpecTest*>> gTests;
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

  bool SpecTest::Run(int argc, const char* argv[])
  {
    if (argc > 1) {
      gTests->erase(
       std::remove_if(
        gTests->begin(),
        gTests->end(),
        [argc, argv](auto test) -> bool {
          for (int i = 1; i < argc; i++) {
            if (std::string(argv[i]) == test->TestName) {
              return false;
            }
          }

          return true;
        }),
       gTests->end());
    }

    std::sort(gTests->begin(), gTests->end(), [](testing::SpecTest* a, testing::SpecTest* b) -> bool {
      auto capitalize = [](std::string& str) {
        for (char& c : str) {
          c &= 0xDF;
        }
      };

      std::string aName = a->TestName;
      std::string bName = b->TestName;

      capitalize(aName);
      capitalize(bName);

      return aName.compare(bName) > 0;
    });

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

  net::Address RandomAddress()
  {
    net::Address retval;
    if (crypto::Random<uint8_t>() & 1) {
      retval.Type = net::AddressType::IPv4;
      for (auto& ip : retval.IPv4) {
        ip = crypto::Random<uint8_t>();
      }
      retval.Port = crypto::Random<uint16_t>();
    } else {
      retval.Type = net::AddressType::IPv6;
      for (auto& ip : retval.IPv6) {
        ip = crypto::Random<uint16_t>();
      }
      retval.Port = crypto::Random<uint16_t>();
    }
    return retval;
  }
}  // namespace testing
