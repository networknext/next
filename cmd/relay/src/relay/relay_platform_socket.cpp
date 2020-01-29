#include "relay_platform_socket.hpp"

#include <cassert>
#include <cstdlib>
#include <cstring>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <fcntl.h>

#include "config.hpp"
#include "net/net.hpp"

#define RELAY_SOCKET_NON_BLOCKING 0
#define RELAY_SOCKET_BLOCKING 1

/*
 * Once windows platforms are developed this will have to change. Likely all this stuff will move into relay_platform_*.cpp
 */

namespace relay
{
    relay_platform_socket_t* relay_platform_socket_create(
        relay_address_t* address, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size)
    {
        assert(address);
        assert(address->type != RELAY_ADDRESS_NONE);

        relay_platform_socket_t* socket = (relay_platform_socket_t*)malloc(sizeof(relay_platform_socket_t));

        assert(socket);

        // create socket

        socket->type = socket_type;

        socket->handle = ::socket((address->type == RELAY_ADDRESS_IPV6) ? AF_INET6 : AF_INET, SOCK_DGRAM, IPPROTO_UDP);

        if (socket->handle < 0) {
            relay_printf("failed to create socket");
            return NULL;
        }

        // force IPv6 only if necessary

        if (address->type == RELAY_ADDRESS_IPV6) {
            int yes = 1;
            if (setsockopt(socket->handle, IPPROTO_IPV6, IPV6_V6ONLY, (char*)(&yes), sizeof(yes)) != 0) {
                relay_printf("failed to set socket ipv6 only");
                relay_platform_socket_destroy(socket);
                return NULL;
            }
        }

        // increase socket send and receive buffer sizes

        if (setsockopt(socket->handle, SOL_SOCKET, SO_SNDBUF, (char*)(&send_buffer_size), sizeof(int)) != 0) {
            relay_printf("failed to set socket send buffer size");
            return NULL;
        }

        if (setsockopt(socket->handle, SOL_SOCKET, SO_RCVBUF, (char*)(&receive_buffer_size), sizeof(int)) != 0) {
            relay_printf("failed to set socket receive buffer size");
            relay_platform_socket_destroy(socket);
            return NULL;
        }

        // bind to port

        if (address->type == RELAY_ADDRESS_IPV6) {
            sockaddr_in6 socket_address;
            memset(&socket_address, 0, sizeof(sockaddr_in6));
            socket_address.sin6_family = AF_INET6;
            for (int i = 0; i < 8; ++i) {
                ((uint16_t*)&socket_address.sin6_addr)[i] = net::relay_htons(address->data.ipv6[i]);
            }
            socket_address.sin6_port = net::relay_htons(address->port);

            if (bind(socket->handle, (sockaddr*)&socket_address, sizeof(socket_address)) < 0) {
                relay_printf("failed to bind socket (ipv6)");
                relay_platform_socket_destroy(socket);
                return NULL;
            }
        } else {
            sockaddr_in socket_address;
            memset(&socket_address, 0, sizeof(socket_address));
            socket_address.sin_family = AF_INET;
            socket_address.sin_addr.s_addr = (((uint32_t)address->data.ipv4[0])) | (((uint32_t)address->data.ipv4[1]) << 8) |
                                             (((uint32_t)address->data.ipv4[2]) << 16) |
                                             (((uint32_t)address->data.ipv4[3]) << 24);
            socket_address.sin_port = relay_htons(address->port);

            if (bind(socket->handle, (sockaddr*)&socket_address, sizeof(socket_address)) < 0) {
                relay_printf("failed to bind socket (ipv4)");
                relay_platform_socket_destroy(socket);
                return NULL;
            }
        }

        // if bound to port 0 find the actual port we got

        if (address->port == 0) {
            if (address->type == RELAY_ADDRESS_IPV6) {
                sockaddr_in6 sin;
                socklen_t len = sizeof(sin);
                if (getsockname(socket->handle, (sockaddr*)(&sin), &len) == -1) {
                    relay_printf("failed to get socket port (ipv6)");
                    relay_platform_socket_destroy(socket);
                    return NULL;
                }
                address->port = relay_platform_ntohs(sin.sin6_port);
            } else {
                sockaddr_in sin;
                socklen_t len = sizeof(sin);
                if (getsockname(socket->handle, (sockaddr*)(&sin), &len) == -1) {
                    relay_printf("failed to get socket port (ipv4)");
                    relay_platform_socket_destroy(socket);
                    return NULL;
                }
                address->port = relay_platform_ntohs(sin.sin_port);
            }
        }

        // set non-blocking io and receive timeout

        if (socket_type == RELAY_SOCKET_NON_BLOCKING) {
            if (fcntl(socket->handle, F_SETFL, O_NONBLOCK, 1) == -1) {
                relay_printf("failed to set socket to non-blocking");
                relay_platform_socket_destroy(socket);
                return NULL;
            }
        } else if (timeout_seconds > 0.0f) {
            // set receive timeout
            struct timeval tv;
            tv.tv_sec = 0;
            tv.tv_usec = (int)(timeout_seconds * 1000000.0f);
            if (setsockopt(socket->handle, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof(tv)) < 0) {
                relay_printf("failed to set socket receive timeout");
                relay_platform_socket_destroy(socket);
                return NULL;
            }
        } else {
            // socket is blocking with no timeout
        }

        return socket;
    }

    void relay_platform_socket_destroy(relay_platform_socket_t* socket)
    {
        assert(socket);
        if (socket->handle != 0) {
            close(socket->handle);
        }
        free(socket);
    }
}  // namespace relay