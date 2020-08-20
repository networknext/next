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



  static bool threads_work = false;

  static relay::relay_platform_thread_return_t RELAY_PLATFORM_THREAD_FUNC test_thread_function(void*)
  {
    threads_work = true;
    RELAY_PLATFORM_THREAD_RETURN();
  }

  static void test_platform_thread()
  {
    relay::relay_platform_thread_t* thread = relay::relay_platform_thread_create(test_thread_function, NULL);
    check(thread != 0);
    relay::relay_platform_thread_join(thread);
    relay::relay_platform_thread_destroy(thread);
    check(threads_work);
  }

  static void test_platform_mutex()
  {
    relay::relay_platform_mutex_t* mutex = relay::relay_platform_mutex_create();
    check(mutex != 0);
    relay_platform_mutex_acquire(mutex);
    relay_platform_mutex_release(mutex);
    {
      relay::relay_mutex_guard(mutex);
      // ...
    }
    relay::relay_platform_mutex_destroy(mutex);
  }

  static void test_bandwidth_limiter()
  {
    relay::relay_bandwidth_limiter_t bandwidth_limiter;

    relay::relay_bandwidth_limiter_reset(&bandwidth_limiter);

    check(relay::relay_bandwidth_limiter_usage_kbps(&bandwidth_limiter, 0.0) == 0.0);

    // come in way under
    {
      const int kbps_allowed = 1000;
      const int packet_bits = 50;

      for (int i = 0; i < 10; ++i) {
        check(!relay::relay_bandwidth_limiter_add_packet(
         &bandwidth_limiter, i * (RELAY_BANDWIDTH_LIMITER_INTERVAL / 10.0), kbps_allowed, packet_bits));
      }
    }

    // get really close
    {
      relay::relay_bandwidth_limiter_reset(&bandwidth_limiter);

      const int kbps_allowed = 1000;
      const int packet_bits = kbps_allowed / 10 * 1000;

      for (int i = 0; i < 10; ++i) {
        check(!relay::relay_bandwidth_limiter_add_packet(
         &bandwidth_limiter, i * (RELAY_BANDWIDTH_LIMITER_INTERVAL / 10.0), kbps_allowed, packet_bits));
      }
    }

    // really close for several intervals
    {
      relay_bandwidth_limiter_reset(&bandwidth_limiter);

      const int kbps_allowed = 1000;
      const int packet_bits = kbps_allowed / 10 * 1000;

      for (int i = 0; i < 30; ++i) {
        check(!relay_bandwidth_limiter_add_packet(
         &bandwidth_limiter, i * (RELAY_BANDWIDTH_LIMITER_INTERVAL / 10.0), kbps_allowed, packet_bits));
      }
    }

    // go over budget
    {
      relay_bandwidth_limiter_reset(&bandwidth_limiter);

      const int kbps_allowed = 1000;
      const int packet_bits = kbps_allowed / 10 * 1000 * 1.01f;

      bool over_budget = false;

      for (int i = 0; i < 30; ++i) {
        over_budget |= relay_bandwidth_limiter_add_packet(
         &bandwidth_limiter, i * (RELAY_BANDWIDTH_LIMITER_INTERVAL / 10.0), kbps_allowed, packet_bits);
      }

      check(over_budget);
    }
  }

  static void test_header()
  {
    uint8_t private_key[crypto_box_SECRETKEYBYTES];

    legacy::relay_random_bytes(private_key, crypto_box_SECRETKEYBYTES);

    uint8_t buffer[RELAY_MTU];

    // client -> server
    {
      uint64_t sequence = 123123130131LL;
      uint64_t session_id = 0x12313131;
      uint8_t session_version = 0x12;

      check(
       relay::relay_write_header(
        RELAY_DIRECTION_CLIENT_TO_SERVER,
        core::packets::Type::ClientToServer,
        sequence,
        session_id,
        session_version,
        private_key,
        buffer,
        sizeof(buffer)) == RELAY_OK);

      core::packets::Type read_type = {};
      uint64_t read_sequence = 0;
      uint64_t read_session_id = 0;
      uint8_t read_session_version = 0;

      check(
       relay::relay_peek_header(
        RELAY_DIRECTION_CLIENT_TO_SERVER,
        &read_type,
        &read_sequence,
        &read_session_id,
        &read_session_version,
        buffer,
        sizeof(buffer)) == RELAY_OK);

      check(read_type == static_cast<uint8_t>(core::packets::Type::ClientToServer));
      check(read_sequence == sequence);
      check(read_session_id == session_id);
      check(read_session_version == session_version);

      check(relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, private_key, buffer, sizeof(buffer)) == RELAY_OK);
    }

    // server -> client
    {
      uint64_t sequence = 123123130131LL | (1ULL << 63);
      uint64_t session_id = 0x12313131;
      uint8_t session_version = 0x12;

      check(
       relay::relay_write_header(
        RELAY_DIRECTION_SERVER_TO_CLIENT,
        core::packets::Type::ServerToClient,
        sequence,
        session_id,
        session_version,
        private_key,
        buffer,
        sizeof(buffer)) == RELAY_OK);

      core::packets::Type read_type = {};
      uint64_t read_sequence = 0;
      uint64_t read_session_id = 0;
      uint8_t read_session_version = 0;

      check(
       relay::relay_peek_header(
        RELAY_DIRECTION_SERVER_TO_CLIENT,
        &read_type,
        &read_sequence,
        &read_session_id,
        &read_session_version,
        buffer,
        sizeof(buffer)) == RELAY_OK);

      check(read_type == static_cast<uint8_t>(core::packets::Type::ServerToClient));
      check(read_sequence == sequence);
      check(read_session_id == session_id);
      check(read_session_version == session_version);

      check(relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, private_key, buffer, sizeof(buffer)) == RELAY_OK);
    }
  }

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
