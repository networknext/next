#include "bench/bench.hpp"

#include "encoding/write.hpp"
#include "encoding/read.hpp"

const auto REPS = 10000000;

Bench(ReadAddress_vs_read_address_ipv4, true)
{
    // ReadAddress()
    {
        net::Address addr;
        std::array<uint8_t, RELAY_ADDRESS_BYTES> buff;
        addr.parse("127.0.0.1:51034");
        size_t i = 0;
        encoding::WriteAddress(buff, i, addr);

        Do(REPS)
        {
            size_t index = 0;
            encoding::ReadAddress(buff, index, addr);
        }

        auto elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "ReadAddress() nanoseconds: " << elapsed << '\n';
    }

    // read_address()
    {
        legacy::relay_address_t addr;
        uint8_t buff[RELAY_ADDRESS_BYTES];

        Do(REPS)
        {
            const uint8_t* ptr = &buff[0];
            encoding::read_address(&ptr, &addr);
        }

        auto elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "read_address() nanoseconds: " << elapsed << '\n';
    }
}