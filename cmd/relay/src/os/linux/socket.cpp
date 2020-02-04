#include "socket.hpp"

#if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

#include "relay/relay_platform_linux.hpp"
#include "util.hpp"

namespace legacy
{
	relay_platform_socket_t* relay_platform_socket_create(
	 legacy::relay_address_t* address, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size)
	{
		assert(address);
		assert(address->type != net::AddressType::None);

		relay_platform_socket_t* socket = (relay_platform_socket_t*)malloc(sizeof(relay_platform_socket_t));

		assert(socket);

		// create socket

		socket->type = socket_type;

		socket->handle =
		 ::socket((address->type == net::AddressType::IPv6) ? AF_INET6 : AF_INET, SOCK_DGRAM, IPPROTO_UDP);

		if (socket->handle < 0) {
			relay_printf("failed to create socket");
			return NULL;
		}

		// force IPv6 only if necessary

		if (address->type == net::AddressType::IPv6) {
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

		if (address->type == net::AddressType::IPv6) {
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
			                                 (((uint32_t)address->data.ipv4[2]) << 16) | (((uint32_t)address->data.ipv4[3]) << 24);
			socket_address.sin_port = net::relay_htons(address->port);

			if (bind(socket->handle, (sockaddr*)&socket_address, sizeof(socket_address)) < 0) {
				relay_printf("failed to bind socket (ipv4)");
				relay_platform_socket_destroy(socket);
				return NULL;
			}
		}

		// if bound to port 0 find the actual port we got

		if (address->port == 0) {
			if (address->type == net::AddressType::IPv6) {
				sockaddr_in6 sin;
				socklen_t len = sizeof(sin);
				if (getsockname(socket->handle, (sockaddr*)(&sin), &len) == -1) {
					relay_printf("failed to get socket port (ipv6)");
					relay_platform_socket_destroy(socket);
					return NULL;
				}
				address->port = relay::relay_platform_ntohs(sin.sin6_port);
			} else {
				sockaddr_in sin;
				socklen_t len = sizeof(sin);
				if (getsockname(socket->handle, (sockaddr*)(&sin), &len) == -1) {
					relay_printf("failed to get socket port (ipv4)");
					relay_platform_socket_destroy(socket);
					return NULL;
				}
				address->port = relay::relay_platform_ntohs(sin.sin_port);
			}
		}

		// set non-blocking io and receive timeout

		if (socket_type == os::SocketType::NonBlocking) {
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

	void relay_platform_socket_send_packet(
	 relay_platform_socket_t* socket, legacy::relay_address_t* to, const void* packet_data, int packet_bytes)
	{
		assert(socket);
		assert(to);
		assert(to->type == net::AddressType::IPv6 || to->type == net::AddressType::IPv4);
		assert(packet_data);
		assert(packet_bytes > 0);

		if (to->type == net::AddressType::IPv6) {
			sockaddr_in6 socket_address;
			memset(&socket_address, 0, sizeof(socket_address));
			socket_address.sin6_family = AF_INET6;
			for (int i = 0; i < 8; ++i) {
				((uint16_t*)&socket_address.sin6_addr)[i] = relay::relay_platform_htons(to->data.ipv6[i]);
			}
			socket_address.sin6_port = relay::relay_platform_htons(to->port);
			int result =
			 int(sendto(socket->handle, (char*)(packet_data), packet_bytes, 0, (sockaddr*)(&socket_address), sizeof(sockaddr_in6)));
			if (result < 0) {
				char address_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
				relay_address_to_string(to, address_string);
				relay_printf("sendto (%s) failed: %s", address_string, strerror(errno));
			}
		} else if (to->type == net::AddressType::IPv4) {
			sockaddr_in socket_address;
			memset(&socket_address, 0, sizeof(socket_address));
			socket_address.sin_family = AF_INET;
			socket_address.sin_addr.s_addr = (((uint32_t)to->data.ipv4[0])) | (((uint32_t)to->data.ipv4[1]) << 8) |
			                                 (((uint32_t)to->data.ipv4[2]) << 16) | (((uint32_t)to->data.ipv4[3]) << 24);
			socket_address.sin_port = relay::relay_platform_htons(to->port);
			int result =
			 int(sendto(socket->handle, (const char*)(packet_data), packet_bytes, 0, (sockaddr*)(&socket_address), sizeof(sockaddr_in)));
			if (result < 0) {
				char address_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
				relay_address_to_string(to, address_string);
				relay_printf("sendto (%s) failed: %s", address_string, strerror(errno));
			}
		} else {
			relay_printf("invalid address type. could not send packet");
		}
	}

	int relay_platform_socket_receive_packet(
	 relay_platform_socket_t* socket, legacy::relay_address_t* from, void* packet_data, int max_packet_size)
	{
		assert(socket);
		assert(from);
		assert(packet_data);
		assert(max_packet_size > 0);

		sockaddr_storage sockaddr_from;
		socklen_t from_length = sizeof(sockaddr_from);

		int result = int(recvfrom(socket->handle,
		 (char*)packet_data,
		 max_packet_size,
		 socket->type == RELAY_PLATFORM_SOCKET_NON_BLOCKING ? MSG_DONTWAIT : 0,
		 (sockaddr*)&sockaddr_from,
		 &from_length));

		if (result <= 0) {
			if (errno == EAGAIN || errno == EINTR) {
				return 0;
			}

			relay_printf("recvfrom failed with error %d", errno);

			return 0;
		}

		if (sockaddr_from.ss_family == AF_INET6) {
			sockaddr_in6* addr_ipv6 = (sockaddr_in6*)&sockaddr_from;
			from->type = static_cast<uint8_t>(net::AddressType::IPv6);  // RELAY_ADDRESS_IPV6;
			for (int i = 0; i < 8; ++i) {
				from->data.ipv6[i] = relay::relay_platform_ntohs(((uint16_t*)&addr_ipv6->sin6_addr)[i]);
			}
			from->port = relay::relay_platform_ntohs(addr_ipv6->sin6_port);
		} else if (sockaddr_from.ss_family == AF_INET) {
			sockaddr_in* addr_ipv4 = (sockaddr_in*)&sockaddr_from;
			from->type = static_cast<uint8_t>(net::AddressType::IPv4);  // RELAY_ADDRESS_IPV4;
			from->data.ipv4[0] = (uint8_t)((addr_ipv4->sin_addr.s_addr & 0x000000FF));
			from->data.ipv4[1] = (uint8_t)((addr_ipv4->sin_addr.s_addr & 0x0000FF00) >> 8);
			from->data.ipv4[2] = (uint8_t)((addr_ipv4->sin_addr.s_addr & 0x00FF0000) >> 16);
			from->data.ipv4[3] = (uint8_t)((addr_ipv4->sin_addr.s_addr & 0xFF000000) >> 24);
			from->port = relay::relay_platform_ntohs(addr_ipv4->sin_port);
		} else {
			assert(0);
			return 0;
		}

		assert(result >= 0);

		return result;
	}

}  // namespace legacy

#endif