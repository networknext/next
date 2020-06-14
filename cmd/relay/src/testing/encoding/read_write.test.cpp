#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

Test(ReadAndWritingAddresses)
{
  net::Address a, b, c;

  b.parse("127.0.0.1:50000");

  c.parse("[::1]:50000");

  std::array<uint8_t, 1024> buffer;

  size_t index = 0;
  encoding::WriteAddress(buffer, index, a);
  check(index == net::Address::ByteSize);
  encoding::WriteAddress(buffer, index, b);
  check(index == net::Address::ByteSize * 2);
  encoding::WriteAddress(buffer, index, c);
  check(index == net::Address::ByteSize * 3);

  net::Address read_a, read_b, read_c;

  index = 0;
  encoding::ReadAddress(buffer, index, read_a);
  check(index == net::Address::ByteSize);
  encoding::ReadAddress(buffer, index, read_b);
  check(index == net::Address::ByteSize * 2);
  encoding::ReadAddress(buffer, index, read_c);
  check(index == net::Address::ByteSize * 3);

  check(a == read_a);
  check(b == read_b);
  check(c == read_c);
}

Test(legacy_test_basic_read_and_write)
{
  uint8_t buffer[1024];

  uint8_t* p = buffer;
  legacy::write_uint8(&p, 105);
  legacy::write_uint16(&p, 10512);
  legacy::write_uint32(&p, 105120000);
  legacy::write_uint64(&p, 105120000000000000LL);
  legacy::write_float32(&p, 100.0f);
  legacy::write_float64(&p, 100000000000000.0);
  legacy::write_bytes(&p, (uint8_t*)"hello", 6);
  legacy::write_string(&p, "hey ho, let's go!", 32);

  const uint8_t* q = buffer;

  uint8_t a = legacy::read_uint8(&q);
  uint16_t b = legacy::read_uint16(&q);
  uint32_t c = legacy::read_uint32(&q);
  uint64_t d = legacy::read_uint64(&q);
  float e = legacy::read_float32(&q);
  double f = legacy::read_float64(&q);
  uint8_t g[6];
  legacy::read_bytes(&q, g, 6);
  char string_buffer[32 + 1];
  memset(string_buffer, 0xFF, sizeof(string_buffer));
  legacy::read_string(&q, string_buffer, 32);
  check(strcmp(string_buffer, "hey ho, let's go!") == 0);

  check(a == 105);
  check(b == 10512);
  check(c == 105120000);
  check(d == 105120000000000000LL);
  check(e == 100.0f);
  check(f == 100000000000000.0);
  check(memcmp(g, "hello", 6) == 0);
}

Test(legacy_test_address_read_and_write)
{
  legacy::relay_address_t a, b, c;

  memset(&a, 0, sizeof(a));

  legacy::relay_address_parse(&b, "127.0.0.1:50000");

  legacy::relay_address_parse(&c, "[::1]:50000");

  uint8_t buffer[1024];

  uint8_t* p = buffer;

  legacy::write_address(&p, &a);
  legacy::write_address(&p, &b);
  legacy::write_address(&p, &c);

  struct legacy::relay_address_t read_a, read_b, read_c;

  const uint8_t* q = buffer;

  legacy::read_address(&q, &read_a);
  legacy::read_address(&q, &read_b);
  legacy::read_address(&q, &read_c);

  check(legacy::relay_address_equal(&a, &read_a) != 0);
  check(legacy::relay_address_equal(&b, &read_b) != 0);
  check(legacy::relay_address_equal(&c, &read_c) != 0);
}
