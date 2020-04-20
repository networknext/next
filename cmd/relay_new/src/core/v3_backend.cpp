#include "includes.h"
#include "v3_backend.hpp"

using namespace std::chrono_literals;

namespace core
{
  auto V3Backend::init() -> bool
  {
    return true;
  }

  auto V3Backend::updateCycle(const volatile bool& handle) -> bool {
    while (handle) {

      if (!update()) {
        return false;
      }

      std::this_thread::sleep_for(10s);
    }

    return true;
  }

  auto V3Backend::update() -> bool
  {
    return true;
  }
}  // namespace core
