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
        std::string toString();

        bool operator==(const RelayAddress& other);
        bool equal(const RelayAddress& other);

        // mainly for debugging

        inline uint16_t port()
        {
            return mPort;
        }

       private:
        uint8_t mType;
        uint16_t mPort;

        union
        {
            std::array<uint8_t, 4> mIPv4;
            std::array<uint16_t, 8> mIPv6;
        };
    };

    // OLD
    struct relay_address_t
    {
        union
        {
            uint8_t ipv4[4];
            uint16_t ipv6[8];
        } data;
        uint16_t port;
        uint8_t type;
    };

    int relay_address_parse(relay_address_t* address, const char* address_string_in);

    const char* relay_address_to_string(const relay_address_t* address, char* buffer);

    int relay_address_equal(const relay_address_t* a, const relay_address_t* b);
}  // namespace relay
#endif