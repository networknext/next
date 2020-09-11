#pragma once

#include "util/macros.hpp"

namespace crypto
{
  template <typename T>
  INLINE auto random_bytes(T& buffer, size_t length) -> bool
  {
    if (buffer.size() < length) {
      return false;
    }
    randombytes_buf(buffer.data(), length);
    return true;
  }

  template <typename T>
  INLINE auto create_nonce_bytes(T& buffer) -> bool
  {
    return random_bytes(buffer, crypto_box_NONCEBYTES);
  }
}  // namespace crypto
