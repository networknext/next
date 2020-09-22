#pragma once

#include "util/macros.hpp"

namespace crypto
{
  using Nonce = std::array<uint8_t, crypto_box_NONCEBYTES>;

  template <typename T>
  INLINE auto random_bytes(T& buffer, size_t length) -> bool
  {
    if (buffer.size() < length) {
      return false;
    }
    randombytes_buf(buffer.data(), length);
    return true;
  }

  INLINE auto make_nonce(Nonce& buffer) -> bool
  {
    return random_bytes(buffer, buffer.size());
  }
}  // namespace crypto
