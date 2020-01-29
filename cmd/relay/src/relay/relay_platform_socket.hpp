#ifndef RELAY_RELAY_PLATFORM_SOCKET
#define RELAY_RELAY_PLATFORM_SOCKET

#include "relay_address.hpp"

namespace relay
{
    typedef int relay_platform_socket_handle_t;

    struct relay_platform_socket_t
    {
        int type;
        relay_platform_socket_handle_t handle;
    };

    relay_platform_socket_t* relay_platform_socket_create(
        relay_address_t* address, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size);
    void relay_platform_socket_destroy(relay_platform_socket_t* socket);
}  // namespace relay
#endif