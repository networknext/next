#ifndef CRYPTO_BYTES_HPP
#define CRYPTO_BYTES_HPP
namespace crypto
{
  template <typename T>
  std::enable_if_t<std::numeric_limits<T>::is_integer, T> Random()
  {
    static auto rand = std::bind(std::uniform_int_distribution<T>(), std::default_random_engine());
    return rand();
  }

  template <typename T>
  [[gnu::always_inline]] inline void RandomBytes(T& buffer, int bytes)
  {
    randombytes_buf(buffer.data(), bytes);
  }

  template <typename T>
  [[gnu::always_inline]] inline void CreateNonceBytes(T& buffer)
  {
    RandomBytes(buffer, crypto_box_NONCEBYTES);
  }
}  // namespace crypto

namespace legacy
{
  inline void relay_random_bytes(uint8_t* buffer, int bytes)
  {
    randombytes_buf(buffer, bytes);
  }
}  // namespace legacy
#endif