#pragma once

namespace util
{
  template <typename T, int N>
  constexpr int array_length(T (&)[N])
  {
    return N;
  }
}  // namespace util
