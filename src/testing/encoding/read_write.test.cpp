#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

TEST(ReadAndWritingAddresses)
{
  net::Address a, b, c;

  b.parse("127.0.0.1:50000");

  c.parse("[::1]:50000");

  std::array<uint8_t, 1024> buffer;

  size_t index = 0;
  encoding::write_address(buffer, index, a);
  CHECK(index == net::Address::SIZE_OF);
  encoding::write_address(buffer, index, b);
  CHECK(index == net::Address::SIZE_OF * 2);
  encoding::write_address(buffer, index, c);
  CHECK(index == net::Address::SIZE_OF * 3);

  net::Address read_a, read_b, read_c;

  index = 0;
  encoding::read_address(buffer, index, read_a);
  CHECK(index == net::Address::SIZE_OF);
  encoding::read_address(buffer, index, read_b);
  CHECK(index == net::Address::SIZE_OF * 2);
  encoding::read_address(buffer, index, read_c);
  CHECK(index == net::Address::SIZE_OF * 3);

  CHECK(a == read_a);
  CHECK(b == read_b);
  CHECK(c == read_c);
}
