/*
    Network Next SDK. Copyright © 2017 - 2020 Network Next, Inc.

	Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following
	conditions are met:

	1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

	2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions
	   and the following disclaimer in the documentation and/or other materials provided with the distribution.

	3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote
	   products derived from this software without specific prior written permission.

	THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
	INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
	IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
	CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
	OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
	NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

#pragma once

#include "NetworkNextBuildConfig.h"
#include "CoreMinimal.h"
#include "OnlineSubsystemNames.h"
#include "SocketSubsystem.h"
#include "Sockets.h"
#include "UObject/ObjectMacros.h"
#include "NetworkNext.h"
#include "SocketNetworkNext.h"
#if defined(NETWORKNEXT_INCLUDE_NEXT_H_WITH_SHORT_PATH)
#include "next.h"
#else
#include "NetworkNextLibrary/next/include/next.h"
#endif
#include "Queue.h"
#include "NetworkNextClientStats.h"

class UNetworkNextNetDriver;

class FSocketNetworkNextClient :
	public FSocketNetworkNext
{
private:

	next_client_t* NetworkNextClient;

	static void OnPacketReceived(next_client_t* client, void* context, const uint8_t* packet_data, int packet_bytes);

	struct PacketData {
		const uint8_t* packet_data;
		int packet_bytes;
	};

	TQueue<PacketData> PacketQueue;

	uint32_t PacketQueueSize;

	FString ServerAddrAndPort;
	FString ServerAddr;
	int32 ServerPort;

	bool bConnected;
	bool bBound;
	int ClientPort;

	FNetworkNextClientStats ClientStats;

public:
#if defined(NETWORKNEXT_HAS_ESOCKETPROTOCOLFAMILY)
	FSocketNetworkNextClient(const FString& InSocketDescription, ESocketProtocolFamily InSocketProtocol, const UNetworkNextNetDriver* InNetDriver);
#else
	FSocketNetworkNextClient(const FString& InSocketDescription, const UNetworkNextNetDriver* InNetDriver);
#endif

	virtual ~FSocketNetworkNextClient();

	virtual void UpdateNetworkNextSocket() override;

	uint64_t GetSessionId();

	const FNetworkNextClientStats& GetClientStats();

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
	* Queries the socket to determine if there is pending data on the queue
	*
	* @param PendingDataSize out parameter indicating how much data is on the pipe for a single recv call
	*
	* @return true if the socket has data, false otherwise
	*/
	virtual bool HasPendingData(uint32& PendingDataSize) override;

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
	 * Reads a chunk of data from the socket. Gathers the source address too
	 *
	 * @param Data the buffer to read into
	 * @param BufferSize the max size of the buffer
	 * @param BytesRead out param indicating how many bytes were read from the socket
	 * @param Source out param receiving the address of the sender of the data
	 * @param Flags the receive flags (must be ESocketReceiveFlags::None)
	 */
	virtual bool RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags = ESocketReceiveFlags::None) override;

	/**
	 * Reads the address the socket is bound to and returns it
	 * 
	 * @param OutAddr address the socket is bound to
	 */
	virtual void GetAddress(FInternetAddr& OutAddr) override;

	/**
	 * Reads the port this socket is bound to.
	 */ 
	virtual int32 GetPortNo() override;

	/**
	 * Converts an ip address stored as a uint32 to a dot separated ipv4 string
	 * For example: 574878368 would become "34.67.242.160"
	 *
	 * @param ipAsInt address the socket is bound to
	 */
	FString IntToIPv4String(uint32 ipAsInt);
};
