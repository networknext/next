#include "bench.hpp"
#include "relay/relay_address.hpp"

Bench(relay_RelayAddress)
{
    auto addr = "127.0.0.1:1234";
    relay::RelayAddress object;
    relay::relay_address_t structure;

    Timer.reset();
    object.parse(addr);
    auto object_elapsed = Timer.elapsed<benchmarking::Microsecond>();

    Timer.reset();
    relay::relay_address_parse(&structure, addr);
    auto structure_elapsed = Timer.elapsed<benchmarking::Microsecond>();

    std::cout << "New way microseconds: " << object_elapsed << '\n';
    std::cout << "Old way microseconds: " << structure_elapsed << '\n';
}
