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
#include "relay/relay_platform.hpp"
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

    template <typename Stream>
    bool Serialize(Stream& stream)
    {
      const TestContext& context = *(const TestContext*)stream.GetContext();

      serialize_int(stream, data.a, context.min, context.max);
      serialize_int(stream, data.b, context.min, context.max);

      serialize_int(stream, data.c, -100, 10000);

      serialize_bits(stream, data.d, 6);
      serialize_bits(stream, data.e, 8);
      serialize_bits(stream, data.f, 7);

      serialize_align(stream);

      serialize_bool(stream, data.g);

      serialize_int(stream, data.numItems, 0, MaxItems - 1);
      for (int i = 0; i < data.numItems; ++i)
        serialize_bits(stream, data.items[i], 8);

      serialize_float(stream, data.float_value);

      serialize_double(stream, data.double_value);

      serialize_uint64(stream, data.uint64_value);

      serialize_bytes(stream, data.bytes, sizeof(data.bytes));

      serialize_string(stream, data.string, sizeof(data.string));

      serialize_address(stream, data.address_a);
      serialize_address(stream, data.address_b);
      serialize_address(stream, data.address_c);

      return true;
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

  static void test_endian()
  {
    uint32_t value = 0x11223344;

    const char* bytes = (const char*)&value;

#if RELAY_LITTLE_ENDIAN

    check(bytes[0] == 0x44);
    check(bytes[1] == 0x33);
    check(bytes[2] == 0x22);
    check(bytes[3] == 0x11);

#else  // #if RELAY_LITTLE_ENDIAN

    check(bytes[3] == 0x44);
    check(bytes[2] == 0x33);
    check(bytes[1] == 0x22);
    check(bytes[0] == 0x11);

#endif  // #if RELAY_LITTLE_ENDIAN
  }

  static void test_bitpacker()
  {
    const int BufferSize = 256;

    uint8_t buffer[BufferSize];

    encoding::BitWriter writer(buffer, BufferSize);

    check(writer.GetData() == buffer);
    check(writer.GetBitsWritten() == 0);
    check(writer.GetBytesWritten() == 0);
    check(writer.GetBitsAvailable() == BufferSize * 8);

    writer.WriteBits(0, 1);
    writer.WriteBits(1, 1);
    writer.WriteBits(10, 8);
    writer.WriteBits(255, 8);
    writer.WriteBits(1000, 10);
    writer.WriteBits(50000, 16);
    writer.WriteBits(9999999, 32);
    writer.FlushBits();

    const int bitsWritten = 1 + 1 + 8 + 8 + 10 + 16 + 32;

    check(writer.GetBytesWritten() == 10);
    check(writer.GetBitsWritten() == bitsWritten);
    check(writer.GetBitsAvailable() == BufferSize * 8 - bitsWritten);

    const int bytesWritten = writer.GetBytesWritten();

    check(bytesWritten == 10);

    memset(buffer + bytesWritten, 0, BufferSize - bytesWritten);

    encoding::BitReader reader(buffer, bytesWritten);

    check(reader.GetBitsRead() == 0);
    check(reader.GetBitsRemaining() == bytesWritten * 8);

    uint32_t a = reader.ReadBits(1);
    uint32_t b = reader.ReadBits(1);
    uint32_t c = reader.ReadBits(8);
    uint32_t d = reader.ReadBits(8);
    uint32_t e = reader.ReadBits(10);
    uint32_t f = reader.ReadBits(16);
    uint32_t g = reader.ReadBits(32);

    check(a == 0);
    check(b == 1);
    check(c == 10);
    check(d == 255);
    check(e == 1000);
    check(f == 50000);
    check(g == 9999999);

    check(reader.GetBitsRead() == bitsWritten);
    check(reader.GetBitsRemaining() == bytesWritten * 8 - bitsWritten);
  }

  static void test_stream()
  {
    const int BufferSize = 1024;

    uint8_t buffer[BufferSize];

    TestContext context;
    context.min = -10;
    context.max = +10;

    encoding::WriteStream writeStream(buffer, BufferSize);

    TestObject writeObject;
    writeObject.Init();
    writeStream.SetContext(&context);
    writeObject.Serialize(writeStream);
    writeStream.Flush();

    const int bytesWritten = writeStream.GetBytesProcessed();

    memset(buffer + bytesWritten, 0, BufferSize - bytesWritten);

    TestObject readObject;
    encoding::ReadStream readStream(buffer, bytesWritten);
    readStream.SetContext(&context);
    readObject.Serialize(readStream);

    check(readObject == writeObject);
  }

  static void test_random_bytes()
  {
    const int BufferSize = 64;
    uint8_t buffer[BufferSize];
    legacy::relay_random_bytes(buffer, BufferSize);
    for (int i = 0; i < 100; ++i) {
      uint8_t next_buffer[BufferSize];
      legacy::relay_random_bytes(next_buffer, BufferSize);
      check(memcmp(buffer, next_buffer, BufferSize) != 0);
      memcpy(buffer, next_buffer, BufferSize);
    }
  }

  static void test_crypto_box()
  {
#define CRYPTO_BOX_MESSAGE (const unsigned char*)"test"
#define CRYPTO_BOX_MESSAGE_LEN 4
#define CRYPTO_BOX_CIPHERTEXT_LEN (crypto_box_MACBYTES + CRYPTO_BOX_MESSAGE_LEN)

    unsigned char sender_publickey[crypto_box_PUBLICKEYBYTES];
    unsigned char sender_secretkey[crypto_box_SECRETKEYBYTES];
    crypto_box_keypair(sender_publickey, sender_secretkey);

    unsigned char receiver_publickey[crypto_box_PUBLICKEYBYTES];
    unsigned char receiver_secretkey[crypto_box_SECRETKEYBYTES];
    crypto_box_keypair(receiver_publickey, receiver_secretkey);

    unsigned char nonce[crypto_box_NONCEBYTES];
    unsigned char ciphertext[CRYPTO_BOX_CIPHERTEXT_LEN];
    legacy::relay_random_bytes(nonce, sizeof nonce);
    check(
     crypto_box_easy(ciphertext, CRYPTO_BOX_MESSAGE, CRYPTO_BOX_MESSAGE_LEN, nonce, receiver_publickey, sender_secretkey) == 0);

    unsigned char decrypted[CRYPTO_BOX_MESSAGE_LEN];
    check(
     crypto_box_open_easy(decrypted, ciphertext, CRYPTO_BOX_CIPHERTEXT_LEN, nonce, sender_publickey, receiver_secretkey) == 0);

    check(memcmp(decrypted, CRYPTO_BOX_MESSAGE, CRYPTO_BOX_MESSAGE_LEN) == 0);
  }

  static void test_crypto_secret_box()
  {
#define CRYPTO_SECRET_BOX_MESSAGE ((const unsigned char*)"test")
#define CRYPTO_SECRET_BOX_MESSAGE_LEN 4
#define CRYPTO_SECRET_BOX_CIPHERTEXT_LEN (crypto_secretbox_MACBYTES + CRYPTO_SECRET_BOX_MESSAGE_LEN)

    unsigned char key[crypto_secretbox_KEYBYTES];
    unsigned char nonce[crypto_secretbox_NONCEBYTES];
    unsigned char ciphertext[CRYPTO_SECRET_BOX_CIPHERTEXT_LEN];

    crypto_secretbox_keygen(key);
    randombytes_buf(nonce, crypto_secretbox_NONCEBYTES);
    crypto_secretbox_easy(ciphertext, CRYPTO_SECRET_BOX_MESSAGE, CRYPTO_SECRET_BOX_MESSAGE_LEN, nonce, key);

    unsigned char decrypted[CRYPTO_SECRET_BOX_MESSAGE_LEN];
    check(crypto_secretbox_open_easy(decrypted, ciphertext, CRYPTO_SECRET_BOX_CIPHERTEXT_LEN, nonce, key) == 0);
  }

  static void test_crypto_aead()
  {
#define CRYPTO_AEAD_MESSAGE (const unsigned char*)"test"
#define CRYPTO_AEAD_MESSAGE_LEN 4
#define CRYPTO_AEAD_ADDITIONAL_DATA (const unsigned char*)"123456"
#define CRYPTO_AEAD_ADDITIONAL_DATA_LEN 6

    unsigned char nonce[crypto_aead_chacha20poly1305_NPUBBYTES];
    unsigned char key[crypto_aead_chacha20poly1305_KEYBYTES];
    unsigned char ciphertext[CRYPTO_AEAD_MESSAGE_LEN + crypto_aead_chacha20poly1305_ABYTES];
    unsigned long long ciphertext_len;

    crypto_aead_chacha20poly1305_keygen(key);
    randombytes_buf(nonce, sizeof(nonce));

    crypto_aead_chacha20poly1305_encrypt(
     ciphertext,
     &ciphertext_len,
     CRYPTO_AEAD_MESSAGE,
     CRYPTO_AEAD_MESSAGE_LEN,
     CRYPTO_AEAD_ADDITIONAL_DATA,
     CRYPTO_AEAD_ADDITIONAL_DATA_LEN,
     NULL,
     nonce,
     key);

    unsigned char decrypted[CRYPTO_AEAD_MESSAGE_LEN];
    unsigned long long decrypted_len;
    check(
     crypto_aead_chacha20poly1305_decrypt(
      decrypted,
      &decrypted_len,
      NULL,
      ciphertext,
      ciphertext_len,
      CRYPTO_AEAD_ADDITIONAL_DATA,
      CRYPTO_AEAD_ADDITIONAL_DATA_LEN,
      nonce,
      key) == 0);
  }

  static void test_crypto_aead_ietf()
  {
#define CRYPTO_AEAD_IETF_MESSAGE (const unsigned char*)"test"
#define CRYPTO_AEAD_IETF_MESSAGE_LEN 4
#define CRYPTO_AEAD_IETF_ADDITIONAL_DATA (const unsigned char*)"123456"
#define CRYPTO_AEAD_IETF_ADDITIONAL_DATA_LEN 6

    unsigned char nonce[crypto_aead_xchacha20poly1305_ietf_NPUBBYTES];
    unsigned char key[crypto_aead_xchacha20poly1305_ietf_KEYBYTES];
    unsigned char ciphertext[CRYPTO_AEAD_IETF_MESSAGE_LEN + crypto_aead_xchacha20poly1305_ietf_ABYTES];
    unsigned long long ciphertext_len;

    crypto_aead_xchacha20poly1305_ietf_keygen(key);
    randombytes_buf(nonce, sizeof(nonce));

    crypto_aead_xchacha20poly1305_ietf_encrypt(
     ciphertext,
     &ciphertext_len,
     CRYPTO_AEAD_IETF_MESSAGE,
     CRYPTO_AEAD_IETF_MESSAGE_LEN,
     CRYPTO_AEAD_IETF_ADDITIONAL_DATA,
     CRYPTO_AEAD_IETF_ADDITIONAL_DATA_LEN,
     NULL,
     nonce,
     key);

    unsigned char decrypted[CRYPTO_AEAD_IETF_MESSAGE_LEN];
    unsigned long long decrypted_len;
    check(
     crypto_aead_xchacha20poly1305_ietf_decrypt(
      decrypted,
      &decrypted_len,
      NULL,
      ciphertext,
      ciphertext_len,
      CRYPTO_AEAD_IETF_ADDITIONAL_DATA,
      CRYPTO_AEAD_IETF_ADDITIONAL_DATA_LEN,
      nonce,
      key) == 0);
  }

  static void test_crypto_sign()
  {
#define CRYPTO_SIGN_MESSAGE (const unsigned char*)"test"
#define CRYPTO_SIGN_MESSAGE_LEN 4

    unsigned char public_key[crypto_sign_PUBLICKEYBYTES];
    unsigned char private_key[crypto_sign_SECRETKEYBYTES];
    crypto_sign_keypair(public_key, private_key);

    unsigned char signed_message[crypto_sign_BYTES + CRYPTO_SIGN_MESSAGE_LEN];
    unsigned long long signed_message_len;

    crypto_sign(signed_message, &signed_message_len, CRYPTO_SIGN_MESSAGE, CRYPTO_SIGN_MESSAGE_LEN, private_key);

    unsigned char unsigned_message[CRYPTO_SIGN_MESSAGE_LEN];
    unsigned long long unsigned_message_len;
    check(crypto_sign_open(unsigned_message, &unsigned_message_len, signed_message, signed_message_len, public_key) == 0);
  }

  static void test_crypto_sign_detached()
  {
#define MESSAGE_PART1 ((const unsigned char*)"Arbitrary data to hash")
#define MESSAGE_PART1_LEN 22

#define MESSAGE_PART2 ((const unsigned char*)"is longer than expected")
#define MESSAGE_PART2_LEN 23

    unsigned char public_key[crypto_sign_PUBLICKEYBYTES];
    unsigned char private_key[crypto_sign_SECRETKEYBYTES];
    crypto_sign_keypair(public_key, private_key);

    crypto_sign_state state;

    unsigned char signature[crypto_sign_BYTES];

    crypto_sign_init(&state);
    crypto_sign_update(&state, MESSAGE_PART1, MESSAGE_PART1_LEN);
    crypto_sign_update(&state, MESSAGE_PART2, MESSAGE_PART2_LEN);
    crypto_sign_final_create(&state, signature, NULL, private_key);

    crypto_sign_init(&state);
    crypto_sign_update(&state, MESSAGE_PART1, MESSAGE_PART1_LEN);
    crypto_sign_update(&state, MESSAGE_PART2, MESSAGE_PART2_LEN);
    check(crypto_sign_final_verify(&state, signature, public_key) == 0);
  }

  static void test_crypto_key_exchange()
  {
    uint8_t client_public_key[crypto_kx_PUBLICKEYBYTES];
    uint8_t client_private_key[crypto_kx_SECRETKEYBYTES];
    crypto_kx_keypair(client_public_key, client_private_key);

    uint8_t server_public_key[crypto_kx_PUBLICKEYBYTES];
    uint8_t server_private_key[crypto_kx_SECRETKEYBYTES];
    crypto_kx_keypair(server_public_key, server_private_key);

    uint8_t client_send_key[crypto_kx_SESSIONKEYBYTES];
    uint8_t client_receive_key[crypto_kx_SESSIONKEYBYTES];
    check(
     crypto_kx_client_session_keys(
      client_receive_key, client_send_key, client_public_key, client_private_key, server_public_key) == 0);

    uint8_t server_send_key[crypto_kx_SESSIONKEYBYTES];
    uint8_t server_receive_key[crypto_kx_SESSIONKEYBYTES];
    check(
     crypto_kx_server_session_keys(
      server_receive_key, server_send_key, server_public_key, server_private_key, client_public_key) == 0);

    check(memcmp(client_send_key, server_receive_key, crypto_kx_SESSIONKEYBYTES) == 0);
    check(memcmp(server_send_key, client_receive_key, crypto_kx_SESSIONKEYBYTES) == 0);
  }

  static void test_platform_socket()
  {
    // non-blocking socket (ipv4)
    {
      legacy::relay_address_t bind_address;
      legacy::relay_address_t local_address;
      legacy::relay_address_parse(&bind_address, "0.0.0.0");
      legacy::relay_address_parse(&local_address, "127.0.0.1");
      legacy::relay_platform_socket_t* socket =
       legacy::relay_platform_socket_create(&bind_address, RELAY_PLATFORM_SOCKET_NON_BLOCKING, 1.0, 64 * 1024, 64 * 1024);
      local_address.port = bind_address.port;
      check(socket != 0);
      uint8_t packet[256];
      memset(packet, 0, sizeof(packet));
      legacy::relay_platform_socket_send_packet(socket, &local_address, packet, sizeof(packet));
      legacy::relay_address_t from;
      while (legacy::relay_platform_socket_receive_packet(socket, &from, packet, sizeof(packet))) {
        check(relay_address_equal(&from, &local_address) != 0);
      }
      relay_platform_socket_destroy(socket);
    }

    // blocking socket with timeout (ipv4)
    {
      legacy::relay_address_t bind_address;
      legacy::relay_address_t local_address;
      legacy::relay_address_parse(&bind_address, "0.0.0.0");
      legacy::relay_address_parse(&local_address, "127.0.0.1");
      legacy::relay_platform_socket_t* socket =
       legacy::relay_platform_socket_create(&bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, 0.01f, 64 * 1024, 64 * 1024);
      local_address.port = bind_address.port;
      check(socket != 0);
      uint8_t packet[256];
      memset(packet, 0, sizeof(packet));
      legacy::relay_platform_socket_send_packet(socket, &local_address, packet, sizeof(packet));
      legacy::relay_address_t from;
      while (legacy::relay_platform_socket_receive_packet(socket, &from, packet, sizeof(packet))) {
        check(legacy::relay_address_equal(&from, &local_address) != 0);
      }
      legacy::relay_platform_socket_destroy(socket);
    }

    // blocking socket with no timeout (ipv4)
    {
      legacy::relay_address_t bind_address;
      legacy::relay_address_t local_address;
      legacy::relay_address_parse(&bind_address, "0.0.0.0");
      legacy::relay_address_parse(&local_address, "127.0.0.1");
      legacy::relay_platform_socket_t* socket =
       legacy::relay_platform_socket_create(&bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, -1.0f, 64 * 1024, 64 * 1024);
      local_address.port = bind_address.port;
      check(socket != 0);
      uint8_t packet[256];
      memset(packet, 0, sizeof(packet));
      legacy::relay_platform_socket_send_packet(socket, &local_address, packet, sizeof(packet));
      legacy::relay_address_t from;
      legacy::relay_platform_socket_receive_packet(socket, &from, packet, sizeof(packet));
      check(legacy::relay_address_equal(&from, &local_address) != 0);
      legacy::relay_platform_socket_destroy(socket);
    }

    // non-blocking socket (ipv6)
#if RELAY_PLATFORM_HAS_IPV6
    {
      legacy::relay_address_t bind_address;
      legacy::relay_address_t local_address;
      legacy::relay_address_parse(&bind_address, "[::]");
      legacy::relay_address_parse(&local_address, "[::1]");
      legacy::relay_platform_socket_t* socket =
       legacy::relay_platform_socket_create(&bind_address, RELAY_PLATFORM_SOCKET_NON_BLOCKING, 0, 64 * 1024, 64 * 1024);
      local_address.port = bind_address.port;
      check(socket != 0);
      uint8_t packet[256];
      memset(packet, 0, sizeof(packet));
      legacy::relay_platform_socket_send_packet(socket, &local_address, packet, sizeof(packet));
      legacy::relay_address_t from;
      while (legacy::relay_platform_socket_receive_packet(socket, &from, packet, sizeof(packet))) {
        check(legacy::relay_address_equal(&from, &local_address) != 0);
      }
      legacy::relay_platform_socket_destroy(socket);
    }

    // blocking socket with timeout (ipv6)
    {
      legacy::relay_address_t bind_address;
      legacy::relay_address_t local_address;
      legacy::relay_address_parse(&bind_address, "[::]");
      legacy::relay_address_parse(&local_address, "[::1]");
      legacy::relay_platform_socket_t* socket =
       legacy::relay_platform_socket_create(&bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, 0.01f, 64 * 1024, 64 * 1024);
      local_address.port = bind_address.port;
      check(socket != 0);
      uint8_t packet[256];
      memset(packet, 0, sizeof(packet));
      legacy::relay_platform_socket_send_packet(socket, &local_address, packet, sizeof(packet));
      legacy::relay_address_t from;
      while (legacy::relay_platform_socket_receive_packet(socket, &from, packet, sizeof(packet))) {
        check(legacy::relay_address_equal(&from, &local_address) != 0);
      }
      legacy::relay_platform_socket_destroy(socket);
    }

    // blocking socket with no timeout (ipv6)
    {
      legacy::relay_address_t bind_address;
      legacy::relay_address_t local_address;
      legacy::relay_address_parse(&bind_address, "[::]");
      legacy::relay_address_parse(&local_address, "[::1]");
      legacy::relay_platform_socket_t* socket =
       legacy::relay_platform_socket_create(&bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, -1.0f, 64 * 1024, 64 * 1024);
      local_address.port = bind_address.port;
      check(socket != 0);
      uint8_t packet[256];
      memset(packet, 0, sizeof(packet));
      legacy::relay_platform_socket_send_packet(socket, &local_address, packet, sizeof(packet));
      legacy::relay_address_t from;
      legacy::relay_platform_socket_receive_packet(socket, &from, packet, sizeof(packet));
      check(legacy::relay_address_equal(&from, &local_address) != 0);
      legacy::relay_platform_socket_destroy(socket);
    }
#endif
  }

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
