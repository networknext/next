#include "includes.h"
#include "bench/bench.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

const auto REPS = 100000000000;

Bench(WriteAddress_vs_write_address_ipv4)
{
  // ReadAddress()
  {
    net::Address addr;
    std::array<uint8_t, net::Address::ByteSize> buff;
    addr.parse("127.0.0.1:51034");

    Do(REPS)
    {
      size_t index = 0;
      encoding::WriteAddress(buff, index, addr);
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "WriteAddress() nanoseconds: " << elapsed << '\n';
  }

  // read_address()
  {
    legacy::relay_address_t addr;
    uint8_t buff[net::Address::ByteSize];
    legacy::relay_address_parse(&addr, "127.0.0.1:51034");

    Do(REPS)
    {
      uint8_t* ptr = &buff[0];
      legacy::write_address(&ptr, &addr);
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "read_address() nanoseconds: " << elapsed << '\n';
  }
}
