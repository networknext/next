#pragma once

namespace core
{
  // Legacy support for the old v3 backend
  class V3Backend
  {
   public:
    V3Backend() = default;
    ~V3Backend() = default;

    auto init() -> bool;

    auto updateCycle(const volatile bool& handle) -> bool;

   private:
    auto update() -> bool;
  };
}  // namespace core