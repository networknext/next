#ifndef RELAY_RELAY_ADDRESS
#define RELAY_RELAY_ADDRESS

#include <array>
#include <string>
#include <cinttypes>

namespace relay
{
    // NEW
    class RelayAddress
    {
       public:
        RelayAddress();
        ~RelayAddress() = default;

        bool parse(const std::string& address_string_in);

        void toString(std::string& buffer);

        bool operator==(const RelayAddress& other);

       private:
        uint8_t mType;
        uint16_t mPort;

        union
        {
            std::array<uint8_t, 4> mIPv4;
            std::array<uint16_t, 8> mIPv6;
        };
    };
}  // namespace relay
#endif