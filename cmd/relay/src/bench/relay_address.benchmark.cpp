#include "bench.hpp"
#include "relay/relay_address.hpp"

Bench(RelayAddress_vs_relay_address_t_address_parsing)
{
    auto addr = "127.0.0.1:1234";
    relay::RelayAddress object;
    relay::relay_address_t structure;

    Do(100)
    {
        object.parse(addr);
    }
    auto object_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 100;
    std::cout << "parse() microseconds: " << object_elapsed << '\n';

    Do(100)
    {
        relay::relay_address_parse(&structure, addr);
    }

    auto structure_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 100;
    std::cout << "relay_address_parse() microseconds: " << structure_elapsed << '\n';
}

Bench(Relay_vs_relay_address_t_stringify)
{
    auto addr = "127.0.0.1:1234";
    relay::RelayAddress object;
    relay::relay_address_t structure;
    object.parse(addr);
    relay::relay_address_parse(&structure, addr);

    std::string str;
    Do(100)
    {
        object.toString(str);
    }
    auto object_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 100;
    std::cout << "toString() microseconds: " << object_elapsed << '\n';

    char buffer[15];
    Do(100)
    {
        relay::relay_address_to_string(&structure, buffer);
    }

    auto structure_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 100;
    std::cout << "relay_address_to_string() microseconds: " << structure_elapsed << '\n';
}
