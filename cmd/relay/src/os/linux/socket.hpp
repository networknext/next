#include "sysinfo.hpp"

#if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

#ifndef OS_LINUX_SOCKET
#define OS_LINUX_SOCKET

#include <array>
#include <cinttypes>
#include <cassert>

#include <sys/socket.h>
#include <sys/types.h>
#include <arpa/inet.h>
#include <fcntl.h>
#include <unistd.h>
#include <cstring>
#include <cstddef>

#include "net/address.hpp"
#include "net/net.hpp"
#include "util/logger.hpp"

#include "relay/relay_platform.hpp"

namespace os
{
	enum class SocketType : uint8_t
	{
		NonBlocking,
		Blocking
	};

	// SBS = Send Buffer Size
	// RBS = Receive Buffer Size
	// TODO figure out what the average SBS & RBS sizes are throughout the codebase
	template <SocketType Type, unsigned int SBS = 256u, unsigned int RBS = 256u>
	class Socket
	{
	public:
		Socket(float timeout = 0.0f);
		~Socket();

		bool create(net::Address& addr);

		bool send(const net::Address& to, const void* data, size_t size);

		size_t recv(net::Address& from, void* data, size_t max_size);

		void close();

	private:
		int mHandle = 0;
		const SocketType mType = Type;
		const float mTimeout;
		std::array<uint8_t, SBS> mSendBuffer;
		std::array<uint8_t, RBS> mReceiveBuffer;
	};

	template <SocketType Type, unsigned int SBS, unsigned int RBS>
	Socket<Type, SBS, RBS>::Socket(float timeout): mTimeout(timeout)
	{}

	template <SocketType Type, unsigned int SBS, unsigned int RBS>
	Socket<Type, SBS, RBS>::~Socket()
	{
		if (mHandle) {
			close();
		}
	}

	template <SocketType Type, unsigned int SBS, unsigned int RBS>
	bool Socket<Type, SBS, RBS>::create(net::Address& addr)
	{
		assert(addr.Type != net::AddressType::None);

		// create socket
		{
			mHandle = ::socket((addr.Type == net::AddressType::IPv6) ? AF_INET6 : AF_INET, SOCK_DGRAM, IPPROTO_UDP);

			if (mHandle < 0) {
				Log("failed to create socket");
				return false;
			}
		}

		// force IPv6 only if necessary
		{
			if (addr.Type == net::AddressType::IPv6) {
				int yes = 1;
				if (setsockopt(mHandle, IPPROTO_IPV6, IPV6_V6ONLY, &yes, sizeof(yes)) != 0) {
					Log("failed to set socket ipv6 only");
					close();
					return false;
				}
			}
		}

		// increase socket send and receive buffer sizes
		{
			unsigned int sbs = SBS;
			if (setsockopt(mHandle, SOL_SOCKET, SO_SNDBUF, &sbs, sizeof(sbs)) != 0) {
				Log("failed to set socket send buffer size");
				close();
				return false;
			}

			unsigned int rbs = RBS;
			if (setsockopt(mHandle, SOL_SOCKET, SO_RCVBUF, &rbs, sizeof(int)) != 0) {
				Log("failed to set socket receive buffer size");
			}
		}

		// bind to port
		{
			if (addr.Type == net::AddressType::IPv6) {
				sockaddr_in6 socket_address;
				bzero(&socket_address, sizeof(socket_address));

				// TODO test if this works, it'll be faster than memset
				// std::fill(static_cast<uint8_t*>(&socket_address), static_cast<uint8_t*>(&socket_address) +
				// sizeof(socket_address));

				socket_address.sin6_family = AF_INET6;
				for (int i = 0; i < 8; i++) {
					reinterpret_cast<uint16_t*>(&socket_address.sin6_addr)[i] = net::relay_htons(addr.IPv6[i]);
				}

				socket_address.sin6_port = net::relay_htons(addr.Port);

				if (bind(mHandle, reinterpret_cast<sockaddr*>(&socket_address), sizeof(socket_address)) < 0) {
					Log("failed to bind socket (ipv6)");
					close();
					return false;
				}
			} else {
				sockaddr_in socket_address;
				bzero(&socket_address, sizeof(socket_address));
				socket_address.sin_family = AF_INET;
				socket_address.sin_addr.s_addr = (((uint32_t)addr.IPv4[0])) | (((uint32_t)addr.IPv4[1]) << 8) |
				                                 (((uint32_t)addr.IPv4[2]) << 16) | (((uint32_t)addr.IPv4[3]) << 24);
				socket_address.sin_port = net::relay_htons(addr.Port);

				if (bind(mHandle, reinterpret_cast<sockaddr*>(&socket_address), sizeof(socket_address)) < 0) {
					Log("failed to bind socket (ipv4)");
					close();
					return false;
				}
			}
		}

		// if bound to port 0, find the actual port we got
		{
			if (addr.Port == 0) {
				if (addr.Type == net::AddressType::IPv6) {
					sockaddr_in6 sin;
					if (getsockname(mHandle, reinterpret_cast<sockaddr*>(&sin), sizeof(sin)) < 0) {
						Log("failed to get socket port (ipv6)");
						close();
						return false;
					}
					addr.Port = relay::relay_platform_ntohs(sin.sin6_port);
				} else {
					sockaddr_in sin;
					if (getsockname(mHandle, reinterpret_cast<sockaddr*>(&sin), sizeof(sin)) < 0) {
						Log("failed to get socket port (ipv4)");
						close();
						return false;
					}
					addr.Port = relay::relay_platform_ntohs(sin.sin_port);
				}
			}
		}

		// set non-blocking io or receive timeout, or if neither then just blocking with no timeout
		{
			if (Type == SocketType::NonBlocking) {
				if (fcntl(mHandle, F_SETFL, O_NONBLOCK, 1) < 0) {
					Log("failed to set socket to non blocking");
					close();
					return false;
				}
			} else if (mTimeout > 0.0f) {
				timeval tv;
				tv.tv_sec = 0;
				tv.tv_usec = (int)(mTimeout * 1000000.0f);
				if (setsockopt(mHandle, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof(tv)) < 0) {
					Log("failed to set socket receive timeout");
					close();
					return false;
				}
			}
		}

		return true;
	}

	template <SocketType Type, unsigned int SBS, unsigned int RBS>
	inline bool Socket<Type, SBS, RBS>::send(const net::Address& to, const void* data, size_t size)
	{
		assert(to.Type == net::AddressType::IPv4 || to.Type == net::AddressType::IPv6);
		assert(data != nullptr);
		assert(size > 0);

		if (to.Type == net::AddressType::IPv6) {
			sockaddr_in6 socket_address;
			bzero(&socket_address, sizeof(socket_address));
			socket_address.sin6_family = AF_INET6;

			for (int i = 0; i < 8; i++) {
				reinterpret_cast<uint16_t*>(&socket_address.sin6_addr)[i] = relay::relay_platform_htons(to.IPv6[i]);
			}

			socket_address.sin6_port = relay::relay_platform_htons(to.Port);

			auto res = sendto(mHandle, data, size, 0, reinterpret_cast<sockaddr*>(&socket_address), sizeof(sockaddr_in6));
			if (res < 0) {
				std::string addr;
				to.toString(addr);
				Log("sendto (", addr, ") failed: ", strerror(errno));
				return false;
			}
		} else if (to.Type == net::AddressType::IPv4) {
			sockaddr_in socket_address;
			bzero(&socket_address, sizeof(socket_address));
			socket_address.sin_family = AF_INET;
			socket_address.sin_addr.s_addr =
			 (((uint32_t)to.IPv4[0])) | (((uint32_t)to.IPv4[1]) << 8) | (((uint32_t)to.IPv4[2]) << 16) | (((uint32_t)to.IPv4[3]) << 24);

			socket_address.sin_port = relay::relay_platform_htons(to.Port);

			auto res = sendto(mHandle, data, size, 0, reinterpret_cast<sockaddr*>(&socket_address), sizeof(sockaddr_in6));
			if (res < 0) {
				std::string addr;
				to.toString(addr);
				Log("sendto (", addr, ") failed: ", strerror(errno));
				return false;
			}
		} else {
			Log("invalid address type, could not send packet");
			return false;
		}

		return true;
	}

	template <SocketType Type, unsigned int SBS, unsigned int RBS>
	inline size_t Socket<Type, SBS, RBS>::recv(net::Address& from, void* data, size_t max_size)
	{
		assert(data != nullptr);
		assert(max_size > 0);

		sockaddr_storage sockaddr_from;

		auto res = recvfrom(mHandle,
		 data,
		 max_size,
		 Type == SocketType::NonBlocking ? MSG_DONTWAIT : 0,
		 reinterpret_cast<sockaddr*>(&sockaddr_from),
		 sizeof(sockaddr_from));

		if (res <= 0) {
			// if not a timeout, log the error
			if (errno != EAGAIN && errno != EINTR) {
				Log("recvfrom failed with error: ", strerror(errno));
			}

			return 0;
		}

		if (sockaddr_from.ss_family == AF_INET6) {
			sockaddr_in6* addr_ipv6 = reinterpret_cast<sockaddr_in6*>(&sockaddr_from);
			from.Type = net::AddressType::IPv6;
			for (int i = 0; i < 8; i++) {
				from.IPv6[i] = relay::relay_platform_ntohs(reinterpret_cast<uint16_t*>(&addr_ipv6->sin6_addr)[i]);
			}
			from.Port = relay::relay_platform_ntohs(addr_ipv6->sin6_port);
		} else if (sockaddr_from.ss_family == AF_INET) {
			sockaddr_in* addr_ipv4 = reinterpret_cast<sockaddr_in*>(&sockaddr_from);
			from.Type = net::AddressType::IPv4;
			from.IPv4[0] = static_cast<uint8_t>((addr_ipv4->sin_addr.s_addr & 0x000000FF));
			from.IPv4[1] = static_cast<uint8_t>((addr_ipv4->sin_addr.s_addr & 0x0000FF00) >> 8);
			from.IPv4[2] = static_cast<uint8_t>((addr_ipv4->sin_addr.s_addr & 0x00FF0000) >> 16);
			from.IPv4[3] = static_cast<uint8_t>((addr_ipv4->sin_addr.s_addr & 0xFF000000) >> 24);
			from.Port = relay::relay_platform_ntohs(addr_ipv4->sin_port);
		} else {
			Log("received packet with invalid ss family: ", sockaddr_from.ss_family);
			return 0;
		}

		assert(res >= 0);

		return res;
	}

	template <SocketType Type, unsigned int SBS, unsigned int RBS>
	inline void Socket<Type, SBS, RBS>::close()
	{
		::close(mHandle);
	}

    inline bool operator==(SocketType st, int i) {
        return static_cast<int>(st) == i;
    }

    inline bool operator==(int i, SocketType st) {
        return static_cast<int>(st) == i;
    }
}  // namespace os

namespace legacy
{
	typedef int relay_platform_socket_handle_t;

	struct relay_platform_socket_t
	{
		int type;
		relay_platform_socket_handle_t handle;
	};

	relay_platform_socket_t* relay_platform_socket_create(
	 legacy::relay_address_t* address, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size);

	void relay_platform_socket_destroy(relay_platform_socket_t* socket);

	void relay_platform_socket_send_packet(
	 relay_platform_socket_t* socket, legacy::relay_address_t* to, const void* packet_data, int packet_bytes);

	int relay_platform_socket_receive_packet(
	 relay_platform_socket_t* socket, legacy::relay_address_t* from, void* packet_data, int max_packet_size);
}  // namespace legacy
#endif

#endif