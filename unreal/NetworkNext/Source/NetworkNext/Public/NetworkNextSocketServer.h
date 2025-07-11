/*
	Network Next. Copyright 2017 - 2025 Network Next, Inc.  
	Licensed under the Network Next Source Available License 1.0
*/

#pragma once

#include "NetworkNextSocket.h"

struct next_server_t;

#ifndef NEXT_ADDRESS_ALREADY_DEFINED
struct next_address_t
{
	union { uint8_t ipv4[4]; uint16_t ipv6[8]; } data;
	uint16_t port;
	uint8_t type;
};
#define NEXT_ADDRESS_ALREADY_DEFINED
#endif // #ifndef NEXT_ADDRESS_ALREADY_DEFINED

class FNetworkNextSocketServer : public FNetworkNextSocket
{
private:

	next_server_t* NetworkNextServer;

	bool bUpdatedThisFrame;

	struct PacketData {
		next_address_t from;
		uint8_t* packet_data;
		int packet_bytes;
	};

	TQueue<PacketData> PacketQueue;

public:

	FNetworkNextSocketServer(const FString& InSocketDescription, const FName& InSocketProtocol);

	virtual ~FNetworkNextSocketServer();

	virtual void Update() override;

	/**
	 * Closes the socket
	 *
	 * @param true if it closes without errors, false otherwise
	 */
	virtual bool Close() override;

	/**
	 * Binds a socket to a network byte ordered address
	 *
	 * @param Addr the address to bind to
	 *
	 * @return true if successful, false otherwise
	 */
	virtual bool Bind(const FInternetAddr& Addr) override;

	/**
	 * Sends a buffer to a network byte ordered address
	 *
	 * @param Data the buffer to send
	 * @param Count the size of the data to send
	 * @param BytesSent out param indicating how much was sent
	 * @param Destination the network byte ordered address to send to
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
	 * Gets the address the socket is bound to
	 *
	 * @param OutAddr address the socket is bound to
	 */
	virtual void GetAddress(FInternetAddr& OutAddr) override;

	/**
	 * Gets the port number this socket is bound to
	 */
	virtual int32 GetPortNo() override;

	/**
	 * Call this to enabling monitoring and (potentially) acceleration for a client address connected to this server.
	 * Typically, this is called via blueprint on the GameMode object after a player joins, via the Network Next blueprint function "UpgradePlayer".
	 */
	void UpgradeClient(TSharedPtr<const FInternetAddr> RemoteAddr, const FString& UserId);

	/**
	 * Returns true when the server is ready to receive client connections.
	 */
	bool IsReady();

protected:

	// Callback when packets are received from clients

	static void PacketReceived(next_server_t* server, void* context, const next_address_t* client_address, const uint8_t* packet_data, int packet_bytes);
};
