#include "bench.hpp"
#include "relay/relay_address.hpp"

Bench(RelayAddress_vs_relay_address_t_address_parsing)
{
    auto addr = "127.0.0.1:20000";

    relay::RelayAddress object;
    relay::relay_address_t structure;

    Do(1000)
    {
        object.parse(addr);
    }
    auto object_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 1000;
    std::cout << "parse() microseconds: " << object_elapsed << '\n';

    Do(1000)
    {
        relay::relay_address_parse(&structure, addr);
    }

    auto structure_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 1000;
    std::cout << "relay_address_parse() microseconds: " << structure_elapsed << '\n';

    std::cout << "object string: '" << object.toString() << "'\n";
    char buff[256];
    relay::relay_address_parse(&structure, buff);
    std::cout << "structure string: '" << buff << "'\n";
}

Bench(Relay_vs_relay_address_t_stringify)
{
    Skip();
    auto addr = "127.0.0.1:20000";
    relay::RelayAddress object;
    relay::relay_address_t structure;
    object.parse(addr);
    relay::relay_address_parse(&structure, addr);

    std::string str;
    Do(1000)
    {
        object.toString(str);
    }
    auto object_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 1000;
    std::cout << "toString() microseconds: " << object_elapsed << '\n';

    char buffer[64];
    Do(1000)
    {
        relay::relay_address_to_string(&structure, buffer);
    }

    auto structure_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 1000;
    std::cout << "relay_address_to_string() microseconds: " << structure_elapsed << '\n';
}
