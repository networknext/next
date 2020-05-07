#pragma once

namespace crypto
{
  // fnv1a 64
  class FNV
  {
   public:
    FNV(std::string val);

    const uint64_t Value;

   private:
    auto hash(const std::string& val) -> uint64_t;
  };  // namespace crypto

  inline FNV::FNV(std::string val): Value(hash(val)) {}

  inline auto FNV::hash(const std::string& val) -> uint64_t
  {
    uint64_t fnv = 0xCBF29CE484222325;
    for (const auto& chr : val) {
      fnv ^= chr;
      fnv *= 0x00000100000001B3;
    }
    return fnv;
  }
}  // namespace crypto