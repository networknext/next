#include "includes.h"
#include "testing/test.hpp"

#include "crypto/bytes.hpp"
#include "encoding/base64.hpp"
#include "encoding/binary.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "encoding/bit_reader.hpp"
#include "encoding/bit_writer.hpp"
#include "encoding/write_stream.hpp"
#include "encoding/read_stream.hpp"

#include "relay/relay.hpp"
#include "net/address.hpp"
#include "relay/relay_bandwidth_limiter.hpp"
#include "core/packets/types.hpp"

namespace
{
  const int MaxItems = 11;
  struct TestData
  {
    TestData()
    {
      memset(this, 0, sizeof(TestData));
    }

    int a, b, c;
    uint32_t d : 8;
    uint32_t e : 8;
    uint32_t f : 8;
    bool g;
    int numItems;
    int items[MaxItems];
    float float_value;
    double double_value;
    uint64_t uint64_value;
    uint8_t bytes[17];
    char string[256];
    legacy::relay_address_t address_a, address_b, address_c;
  };

  struct TestContext
  {
    int min;
    int max;
  };

  struct TestObject
  {
    TestData data;

    void Init()
    {
      data.a = 1;
      data.b = -2;
      data.c = 150;
      data.d = 55;
      data.e = 255;
      data.f = 127;
      data.g = true;

      data.numItems = MaxItems / 2;
      for (int i = 0; i < data.numItems; ++i)
        data.items[i] = i + 10;

      data.float_value = 3.1415926f;
      data.double_value = 1 / 3.0;
      data.uint64_value = 0x1234567898765432L;

      for (int i = 0; i < (int)sizeof(data.bytes); ++i)
        data.bytes[i] = (i * 37) % 255;

      strcpy(data.string, "hello world!");

      memset(&data.address_a, 0, sizeof(legacy::relay_address_t));

      relay_address_parse(&data.address_b, "127.0.0.1:50000");

      relay_address_parse(&data.address_c, "[::1]:50000");
    }


    bool operator==(const TestObject& other) const
    {
      return memcmp(&data, &other.data, sizeof(TestData)) == 0;
    }

    bool operator!=(const TestObject& other) const
    {
      return !(*this == other);
    }
  };

  static void test_base64()
  {
    const char* input = "a test string. let's see if it works properly";
    char encoded[1024];
    char decoded[1024];
    check(legacy::base64_encode_string(input, encoded, sizeof(encoded)) > 0);
    check(legacy::base64_decode_string(encoded, decoded, sizeof(decoded)) > 0);
    check(strcmp(decoded, input) == 0);
    check(legacy::base64_decode_string(encoded, decoded, 10) == 0);
  }

}  // namespace

Test(Legacy)
{
  check(relay::relay_initialize() == RELAY_OK);
  test_endian();
  test_bitpacker();
  test_stream();
  test_random_bytes();
  test_crypto_box();
  test_crypto_secret_box();
  test_crypto_aead();
  test_crypto_aead_ietf();
  test_crypto_sign();
  test_crypto_sign_detached();
  test_crypto_key_exchange();
  test_platform_socket();
  test_platform_thread();
  test_platform_mutex();
  test_bandwidth_limiter();
  test_header();
  test_base64();

  relay::relay_term();
}
