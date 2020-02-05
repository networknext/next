#include "bench/bench.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

const auto REPS = 10000000000;

Bench(WriteUint16_vs_read_uint16, true)
{
  {
    std::array<uint8_t, 2> buff;

    Do(REPS)
    {
      size_t index = 0;
      encoding::WriteUint16(buff, index, 533);
    }

    auto elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
    std::cout << "WriteUint16() nanoseconds: " << elapsed << '\n';
  }

  {
    uint8_t buff[2];

    Do(REPS)
    {
      uint8_t* ptr = &buff[0];
      encoding::write_uint16(&ptr, 533);
    }

    auto elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
    std::cout << "write_uint16() nanoseconds: " << elapsed << '\n';
  }
}

Bench(WriteAddress_vs_read_address_ipv4, true)
{
  // ReadAddress()
  {
    net::Address addr;
    std::array<uint8_t, RELAY_ADDRESS_BYTES> buff;
    addr.parse("127.0.0.1:51034");

    Do(REPS)
    {
      size_t index = 0;
      encoding::WriteAddress(buff, index, addr);
    }

    auto elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
    std::cout << "WriteAddress() nanoseconds: " << elapsed << '\n';
  }

  // read_address()
  {
    legacy::relay_address_t addr;
    uint8_t buff[RELAY_ADDRESS_BYTES];
    legacy::relay_address_parse(&addr, "127.0.0.1:51034");

    Do(REPS)
    {
      uint8_t* ptr = &buff[0];
      encoding::write_address(&ptr, &addr);
    }

    auto elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
    std::cout << "read_address() nanoseconds: " << elapsed << '\n';
  }
}