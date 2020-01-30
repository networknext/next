#include "bench.hpp"
#include "relay/relay_address.hpp"

#include <algorithm>
#include <cstring>

// const auto REPS = 1;
const auto REPS = 1000000;

Bench(RelayAddress_vs_relay_address_t_address_parsing)
{
    // parse()
    {
        relay::RelayAddress object;
        std::string straddr = "127.0.0.1:20000";
        Do(REPS)
        {
            object.parse(straddr);
        }

        auto elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "parse() nanoseconds: " << elapsed << '\n';
    }

    // relay_address_parse()
    {
        relay::relay_address_t structure;
        Do(REPS)
        {
            relay::relay_address_parse(&structure, "127.0.0.1:20000");
        }

        auto elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "relay_address_parse() nanoseconds: " << elapsed << '\n';
    }
}

Bench(Relay_vs_relay_address_t_stringify_ipv4)
{
    auto addr = "127.0.0.1:20000";
    relay::RelayAddress object;
    relay::relay_address_t structure;
    object.parse(addr);
    relay::relay_address_parse(&structure, addr);

    Do(REPS)
    {
        std::string str;
        object.toString(str);
    }
    auto object_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
    std::cout << "toString() nanoseconds: " << object_elapsed << '\n';

    char buffer[64];
    Do(REPS)
    {
        relay::relay_address_to_string(&structure, buffer);
    }

    auto structure_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
    std::cout << "relay_address_to_string() nanoseconds: " << structure_elapsed << '\n';
}

Bench(Relay_vs_relay_address_t_stringify_ipv6)
{
    auto addr = "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000";
    relay::RelayAddress object;
    relay::relay_address_t structure;
    object.parse(addr);
    relay::relay_address_parse(&structure, addr);

    std::string str;
    Do(REPS)
    {
        object.toString(str);
    }
    auto object_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
    std::cout << "toString() nanoseconds: " << object_elapsed << '\n';

    char buffer[128];
    Do(REPS)
    {
        relay::relay_address_to_string(&structure, buffer);
    }

    auto structure_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
    std::cout << "relay_address_to_string() nanoseconds: " << structure_elapsed << '\n';

    std::cout << "\nclass: " << str << '\n';
    std::cout << "struct: " << buffer << '\n';
}
