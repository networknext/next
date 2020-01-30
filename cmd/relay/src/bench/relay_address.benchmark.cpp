#include "bench.hpp"
#include "relay/relay_address.hpp"

Bench(RelayAddress_vs_relay_address_t_address_parsing)
{
    Skip();
    relay::RelayAddress object;
    relay::relay_address_t structure;

    std::string straddr = "127.0.0.1:20000";
    Do(10000)
    {
        object.parse(straddr);
    }

    auto object_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 10000;
    std::cout << "parse() microseconds: " << object_elapsed << '\n';

    Do(10000)
    {
        relay::relay_address_parse(&structure, "127.0.0.1:20000");
    }

    auto structure_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 10000;
    std::cout << "relay_address_parse() microseconds: " << structure_elapsed << '\n';
}

Bench(Relay_vs_relay_address_t_stringify)
{
    //auto addr = "127.0.0.1:20000";
    auto addr = "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000";
    relay::RelayAddress object;
    relay::relay_address_t structure;
    object.parse(addr);
    relay::relay_address_parse(&structure, addr);

    std::string str;
    Do(1)
    {
        object.toString(str);
    }
    auto object_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 10000;
    std::cout << "toString() microseconds: " << object_elapsed << '\n';

    char buffer[64];
    Do(1)
    {
        relay::relay_address_to_string(&structure, buffer);
    }

    auto structure_elapsed = Timer.elapsed<benchmarking::Microsecond>() / 10000;
    std::cout << "relay_address_to_string() microseconds: " << structure_elapsed << '\n';

    std::cout << "class: " << str << '\n';
    std::cout << "struct: " << buffer << '\n';
}
