/*
	Network Next. Copyright 2017 - 2025 Network Next, Inc.  
	Licensed under the Network Next Source Available License 1.0
*/

#pragma once

#include "NetworkNextSocket.h"

struct next_client_t;

#ifndef NEXT_ADDRESS_ALREADY_DEFINED
struct next_address_t
{
	union { uint8_t ipv4[4]; uint16_t ipv6[8]; } data;
	uint16_t port;
	uint8_t type;
};
#define NEXT_ADDRESS_ALREADY_DEFINED
#endif // #ifndef NEXT_ADDRESS_ALREADY_DEFINED

class FNetworkNextSocketClient : public FNetworkNextSocket
{
private:

	next_client_t* NetworkNextClient;

	FString ServerAddr;
	int32 ServerPort;

	bool bConnected;
	bool bUpdatedThisFrame;

	struct PacketData {
		uint8_t* packet_data;
		int packet_bytes;
	};

	TQueue<PacketData> PacketQueue;

public:

	FNetworkNextSocketClient(const FString& InSocketDescription, const FName& InSocketProtocol);

	virtual ~FNetworkNextSocketClient();

	virtual void Update() override;

	/**
	 * Closes the socket
	 *
	 * @param true if it closes without errors, false otherwise
	 */
	virtual bool Close() override;

	/**
	 * Binds the socket
	 *
	 * @param Addr the address to bind to
	 *
	 * @return true if successful, false otherwise
	 */
	virtual bool Bind(const FInternetAddr& Addr) override;

	/**
	 * Connect to server via IP address.
	 */

	void NetworkNextConnect(const FString& Host, int32 Port);

	/**
	 * Sends a packet to the server.
	 *
	 * @param Data the buffer to send
	 * @param Count the size of the data to send
	 * @param BytesSent out param indicating how much was sent
	 * @param Destination this parameter is ignored. the packet is always sent to the server
	 */
	virtual bool SendTo(const uint8* Data, int32 Count, int32& BytesSent, const FInternetAddr& Destination) override;

	/**
	 * Reads a packet from the socket.
	 *
	 * @param Data the buffer to read into
	 * @param BufferSize the max size of the buffer
	 * @param BytesRead out param indicating how many bytes were read from the socket
	 * @param Source out param receiving the address of the sender of the data
	 * @param Flags the receive flags (must be ESocketReceiveFlags::None)
	 */
	virtual bool RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags = ESocketReceiveFlags::None) override;

	/**
	 * Gets the address the socket is bound to. It is always "0.0.0.0"
	 *
	 * @param OutAddr address the socket is bound to
	 */
	virtual void GetAddress(FInternetAddr& OutAddr) override;

	/**
	 * Gets the port number the socket is bound to.
	 */
	virtual int32 GetPortNo() override;

	/**
	 * Gets the network next round trip latency in milliseconds.
	 * 
	 * This latency is a measurement of true network delay, and excludes any quantization effects
	 * caused by the interplay between client and server framerates.
	 * 
	 * If the client is accelerated, this latency value includes the effect of acceleration.
	 */

	float GetLatency();

	/**
	 * Gets the network next jitter in milliseconds.
	 */

	float GetJitter();

	/**
	 * Gets the network next packet loss %
	 */

	float GetPacketLoss();

protected:

	// Callback for receiving packets sent from the server

	static void PacketReceived(next_client_t *client, void* context, const next_address_t* from, const uint8_t* packet_data, int packet_bytes);
};
