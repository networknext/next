#include "bench.hpp"
#include "relay/relay_address.hpp"

#include <algorithm>
#include <cstring>

// const auto REPS = 1;
const auto REPS = 1000000;

Bench(RelayAddress_vs_relay_address_t_address_parsing)
{
    Skip();
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
        legacy::relay_address_t structure;
        Do(REPS)
        {
            legacy::relay_address_parse(&structure, "127.0.0.1:20000");
        }

        auto elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "relay_address_parse() nanoseconds: " << elapsed << '\n';
    }
}

Bench(Relay_vs_relay_address_t_stringify_ipv4)
{
    Skip();
    auto addr = "127.0.0.1:20000";

    // toString(string)
    {
        relay::RelayAddress object;
        object.parse(addr);
        std::string str;
        Do(REPS)
        {
            object.toString(str);
        }
        auto object_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "toString(string) nanoseconds: " << object_elapsed << '\n';
        std::cout << "Address: " << str << '\n';
    }

    // relay_address_to_string()
    {
        legacy::relay_address_t structure;
        legacy::relay_address_parse(&structure, addr);
        char buffer[64];
        Do(REPS)
        {
            legacy::relay_address_to_string(&structure, buffer);
        }
        auto structure_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "relay_address_to_string() nanoseconds: " << structure_elapsed << '\n';
        std::cout << "Address: " << buffer << '\n';
    }
}

Bench(Relay_vs_relay_address_t_stringify_ipv6_with_braces)
{
    Skip();
    auto addr = "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000";

    {
        relay::RelayAddress object;
        object.parse(addr);
        std::string str;
        Do(REPS)
        {
            object.toString(str);
        }
        auto object_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "toString() nanoseconds: " << object_elapsed << '\n';
        std::cout << "Address: " << str << '\n';
    }

    {
        legacy::relay_address_t structure;
        legacy::relay_address_parse(&structure, addr);
        char buffer[128];
        Do(REPS)
        {
            legacy::relay_address_to_string(&structure, buffer);
        }

        auto structure_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "relay_address_to_string() nanoseconds: " << structure_elapsed << '\n';
        std::cout << "Address: " << buffer << '\n';
    }
}

Bench(Relay_vs_relay_address_t_stringify_ipv6_without_braces)
{
    Skip();
    auto addr = "2001:0db8:85a3:0000:0000:8a2e:0370:7334";

    {
        relay::RelayAddress object;
        object.parse(addr);
        std::string str;
        Do(REPS)
        {
            object.toString(str);
        }
        auto object_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "toString() nanoseconds: " << object_elapsed << '\n';
        std::cout << "Address: " << str << '\n';
    }

    {
        legacy::relay_address_t structure;
        legacy::relay_address_parse(&structure, addr);
        char buffer[368];
        Do(REPS)
        {
            legacy::relay_address_to_string(&structure, buffer);
        }

        auto structure_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "relay_address_to_string() nanoseconds: " << structure_elapsed << '\n';
        std::cout << "Address: " << buffer << '\n';
    }
}

Bench(Relay_vs_relay_address_t_stringify_invalid)
{
    Skip();
    auto addr = "invalid-ip";

    {
        relay::RelayAddress object;
        object.parse(addr);
        std::string str;
        Do(REPS)
        {
            object.toString(str);
        }
        auto object_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "toString() nanoseconds: " << object_elapsed << '\n';
        std::cout << "Address: " << str << '\n';
    }

    {
        legacy::relay_address_t structure;
        legacy::relay_address_parse(&structure, addr);
        char buffer[368];
        Do(REPS)
        {
            legacy::relay_address_to_string(&structure, buffer);
        }

        auto structure_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "relay_address_to_string() nanoseconds: " << structure_elapsed << '\n';
        std::cout << "Address: " << buffer << '\n';
    }
}

Bench(Relay_vs_relay_address_t_equal_ipv4)
{
    auto addr = "127.0.0.1:20000";

    {
        relay::RelayAddress a, b;
        a.parse(addr);
        b.parse(addr);
        bool result;
        Do(REPS)
        {
            result = a == b;
        }
        auto object_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "operator==() nanoseconds: " << object_elapsed << '\n';
        std::cout << "equal: " << result << '\n';
    }

    {
        legacy::relay_address_t a, b;
        legacy::relay_address_parse(&a, addr);
        legacy::relay_address_parse(&b, addr);
        int result;
        Do(REPS)
        {
            result = legacy::relay_address_equal(&a, &b);
        }

        auto structure_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "relay_address_equal() nanoseconds: " << structure_elapsed << '\n';
        std::cout << "equal: " << result << '\n';
    }
}

Bench(Relay_vs_relay_address_t_equal_ipv6_with_braces)
{
    auto addr = "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000";

    {
        relay::RelayAddress a, b;
        a.parse(addr);
        b.parse(addr);
        bool result;
        Do(REPS)
        {
            result = a == b;
        }
        auto object_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "operator==() nanoseconds: " << object_elapsed << '\n';
        std::cout << "equal: " << result << '\n';
    }

    {
        legacy::relay_address_t a, b;
        legacy::relay_address_parse(&a, addr);
        legacy::relay_address_parse(&b, addr);
        int result;
        Do(REPS)
        {
            result = legacy::relay_address_equal(&a, &b);
        }

        auto structure_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "relay_address_equal() nanoseconds: " << structure_elapsed << '\n';
        std::cout << "equal: " << result << '\n';
    }
}

Bench(Relay_vs_relay_address_t_equal_ipv6_without_braces)
{
    auto addr = "2001:0db8:85a3:0000:0000:8a2e:0370:7334";

    {
        relay::RelayAddress a, b;
        a.parse(addr);
        b.parse(addr);
        bool result;
        Do(REPS)
        {
            result = a == b;
        }
        auto object_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "operator==() nanoseconds: " << object_elapsed << '\n';
        std::cout << "equal: " << result << '\n';
    }

    {
        legacy::relay_address_t a, b;
        legacy::relay_address_parse(&a, addr);
        legacy::relay_address_parse(&b, addr);
        int result;
        Do(REPS)
        {
            result = legacy::relay_address_equal(&a, &b);
        }

        auto structure_elapsed = Timer.elapsed<benchmarking::Nanosecond>() / REPS;
        std::cout << "relay_address_equal() nanoseconds: " << structure_elapsed << '\n';
        std::cout << "equal: " << result << '\n';
    }
}
