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
  encoding::write_address(buffer, index, a);
  check(index == net::Address::SIZE_OF);
  encoding::write_address(buffer, index, b);
  check(index == net::Address::SIZE_OF * 2);
  encoding::write_address(buffer, index, c);
  check(index == net::Address::SIZE_OF * 3);

  net::Address read_a, read_b, read_c;

  index = 0;
  encoding::read_address(buffer, index, read_a);
  check(index == net::Address::SIZE_OF);
  encoding::read_address(buffer, index, read_b);
  check(index == net::Address::SIZE_OF * 2);
  encoding::read_address(buffer, index, read_c);
  check(index == net::Address::SIZE_OF * 3);

  check(a == read_a);
  check(b == read_b);
  check(c == read_c);
}
