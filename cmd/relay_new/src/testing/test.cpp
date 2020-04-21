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

  bool SpecTest::Run()
  {
    std::cout << "Test count: " << gTests->size() << '\n';

    std::sort(gTests->begin(), gTests->end(), [](testing::SpecTest* a, testing::SpecTest* b) -> bool {
      return strcmp(a->TestName, b->TestName) > 0;
    });

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

  bool StubbedCurlWrapper::Success = true;
  std::string StubbedCurlWrapper::Request, StubbedCurlWrapper::Response, StubbedCurlWrapper::Hostname,
   StubbedCurlWrapper::Endpoint;
}  // namespace testing
