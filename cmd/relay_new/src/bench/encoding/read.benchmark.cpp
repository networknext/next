#include "includes.h"
#include "bench/bench.hpp"

#include "encoding/write.hpp"
#include "encoding/read.hpp"

const auto REPS = 10000000000;

Bench(ReadUint16_vs_read_uint16)
{
  {
    std::array<uint8_t, 2> buff;

    Do(REPS)
    {
      size_t index = 0;
      encoding::ReadUint16(buff, index);
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "ReadUint16() nanoseconds: " << elapsed << '\n';
  }

  {
    uint8_t buff[2];

    Do(REPS)
    {
      const uint8_t* ptr = &buff[0];
      legacy::read_uint16(&ptr);
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "read_uint16() nanoseconds: " << elapsed << '\n';
  }
}

Bench(ReadAddress_vs_read_address_ipv4)
{
  // ReadAddress()
  {
    net::Address addr;
    std::array<uint8_t, net::Address::ByteSize> buff;
    addr.parse("127.0.0.1:51034");
    size_t i = 0;
    encoding::WriteAddress(buff, i, addr);

    Do(REPS)
    {
      size_t index = 0;
      encoding::ReadAddress(buff, index, addr);
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "ReadAddress() nanoseconds: " << elapsed << '\n';
  }

  // read_address()
  {
    legacy::relay_address_t addr;
    uint8_t buff[net::Address::ByteSize];
    legacy::relay_address_parse(&addr, "127.0.0.1:51034");
    auto p = &buff[0];
    legacy::write_address(&p, &addr);

    Do(REPS)
    {
      const uint8_t* ptr = &buff[0];
      legacy::read_address(&ptr, &addr);
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "read_address() nanoseconds: " << elapsed << '\n';
  }
}
