#pragma once

namespace crypto
{
  // fnv1a 64
  [[gnu::always_inline]] inline auto FNV(const std::string& str) -> uint64_t
  {
    uint64_t fnv = 0xCBF29CE484222325;
    for (const auto& chr : str) {
      fnv ^= chr;
      fnv *= 0x00000100000001B3;
    }
    return fnv;
  }
}  // namespace crypto