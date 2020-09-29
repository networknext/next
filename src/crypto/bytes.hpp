#pragma once

#include "util/macros.hpp"

namespace crypto
{
  template <typename T>
  INLINE std::enable_if_t<std::numeric_limits<T>::is_integer, T> Random()
  {
    static auto rand = std::bind(std::uniform_int_distribution<T>(), std::default_random_engine());
    return rand();
  }

  template <typename T>
  INLINE auto RandomBytes(T& buffer, size_t length) -> bool
  {
    if (buffer.size() < length) {
      return false;
    }
    randombytes_buf(buffer.data(), length);
    return true;
  }

  template <typename T>
  INLINE auto CreateNonceBytes(T& buffer) -> bool
  {
    return RandomBytes(buffer, crypto_box_NONCEBYTES);
  }
}  // namespace crypto
