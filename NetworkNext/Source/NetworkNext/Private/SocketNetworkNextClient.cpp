/*
    Network Next SDK 3.1.3

    Copyright © 2017 - 2019 Network Next, Inc.

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

#include "SocketNetworkNextClient.h"
#include "NetworkNextNetDriver.h"
#include "NetworkNextClientStats.h"
#include "AssertionMacros.h"
#include "NetworkNext.h"

#if defined(NETWORKNEXT_HAS_ESOCKETPROTOCOLFAMILY)
FSocketNetworkNextClient::FSocketNetworkNextClient(const FString& InSocketDescription, ESocketProtocolFamily InSocketProtocol, const UNetworkNextNetDriver* InNetDriver) :
	FSocketNetworkNext(InSocketDescription, InSocketProtocol, ESocketNetworkNextType::TYPE_Client)
#else
FSocketNetworkNextClient::FSocketNetworkNextClient(const FString& InSocketDescription, const UNetworkNextNetDriver* InNetDriver) :
	FSocketNetworkNext(InSocketDescription, ESocketNetworkNextType::TYPE_Client)
#endif
{
	check(InNetDriver);

	this->NetworkNextClient = next_client_create(
		TCHAR_TO_ANSI(*(InNetDriver->CustomerPublicKeyBase64)),
		this,
		&FSocketNetworkNextClient::OnPacketReceived
	);

	if (this->NetworkNextClient == nullptr)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Unable to create Network Next client."));
	}

	this->PacketQueueSize = 0;
	this->bConnected = false;
	this->ClientStats = FNetworkNextClientStats::GetDisconnectedStats();
}

FSocketNetworkNextClient::~FSocketNetworkNextClient()
{
	if (this->NetworkNextClient != nullptr)
	{
		next_client_destroy(this->NetworkNextClient);
	}

	this->NetworkNextClient = nullptr;
}

uint64_t FSocketNetworkNextClient::GetSessionId()
{
	if (!this->bConnected || this->NetworkNextClient == nullptr)
	{
		return 0;
	}

	return next_client_session_id(this->NetworkNextClient);
}

const FNetworkNextClientStats& FSocketNetworkNextClient::GetClientStats()
{
	return this->ClientStats;
}

void FSocketNetworkNextClient::UpdateNetworkNextSocket()
{
	if (this->bConnected && this->NetworkNextClient != nullptr)
	{
		next_client_update(this->NetworkNextClient);

		const next_client_stats_t* client_stats = next_client_stats(this->NetworkNextClient);

		switch (client_stats->connection_type)
		{
		case NEXT_CONNECTION_TYPE_WIRED:
			this->ClientStats.ConnectionType = ENetworkNextConnectionType::ConnectionType_Wired;
			break;
		case NEXT_CONNECTION_TYPE_WIFI:
			this->ClientStats.ConnectionType = ENetworkNextConnectionType::ConnectionType_Wifi;
			break;
		case NEXT_CONNECTION_TYPE_CELLULAR:
			this->ClientStats.ConnectionType = ENetworkNextConnectionType::ConnectionType_Cellular;
			break;
		default:
			this->ClientStats.ConnectionType = ENetworkNextConnectionType::ConnectionType_Unknown;
			break;
		}

		this->ClientStats.OnNetworkNext = client_stats->next;
		this->ClientStats.DirectRtt = client_stats->direct_rtt;
		this->ClientStats.DirectJitter = client_stats->direct_jitter;
		this->ClientStats.DirectPacketLoss = client_stats->direct_packet_loss;
		this->ClientStats.NetworkNextRtt = client_stats->next_rtt;
		this->ClientStats.NetworkNextJitter = client_stats->next_jitter;
		this->ClientStats.NetworkNextPacketLoss = client_stats->next_packet_loss;
		this->ClientStats.KbpsUp = client_stats->kbps_up;
		this->ClientStats.KbpsDown = client_stats->kbps_down;
	}
}

void FSocketNetworkNextClient::OnPacketReceived(next_client_t* client, void* context, const uint8_t* packet_data, int packet_bytes)
{
	FSocketNetworkNextClient* self = (FSocketNetworkNextClient*)context;

	uint8_t* packet_data_copy = (uint8_t*)malloc(packet_bytes);
	memcpy(packet_data_copy, packet_data, packet_bytes);

	self->PacketQueue.Enqueue({
		packet_data_copy,
		packet_bytes,
	});
	self->PacketQueueSize += packet_bytes;
}

bool FSocketNetworkNextClient::Close()
{
	if (this->NetworkNextClient != nullptr)
	{
		next_client_close_session(this->NetworkNextClient);
	}

	this->bConnected = false;

	return true;
}

bool FSocketNetworkNextClient::Bind(const FInternetAddr& Addr)
{
	// Ignored on client
	return true;
}

bool FSocketNetworkNextClient::HasPendingData(uint32& PendingDataSize)
{
	if (!this->bConnected || this->NetworkNextClient == nullptr)
	{
		PendingDataSize = 0;
		return false;
	}

	PendingDataSize = this->PacketQueueSize;
	return this->PacketQueueSize > 0;
}

bool FSocketNetworkNextClient::SendTo(const uint8* Data, int32 Count, int32& BytesSent, const FInternetAddr& Destination)
{
	if (this->NetworkNextClient == nullptr)
	{
		// We could not create the Network Next client, so we can't do anything.
		BytesSent = 0;
		return false;
	}

	// The first send indicates the server that we want to connect to.
	if (!bConnected)
	{
		this->ServerAddrAndPort = Destination.ToString(true);
		this->ServerAddr = Destination.ToString(false);
		this->ServerPort = Destination.GetPort();

		next_client_open_session(
			this->NetworkNextClient,
			TCHAR_TO_ANSI(*this->ServerAddrAndPort)
		);

		this->bConnected = true;
	}

	FString Target = Destination.ToString(true);

	if (!Target.Equals(this->ServerAddrAndPort))
	{
		// Sockets in Network Next can only ever send to the same destination.
		UE_LOG(
			LogNetworkNext, 
			Error, 
			TEXT("Attempted to use socket to send data to %s, but it's already been used to send data to %s."), 
			*Target, 
			*this->ServerAddrAndPort
		);
		return false;
	}

	next_client_send_packet(
		this->NetworkNextClient,
		Data,
		Count
	);
	BytesSent = Count;

	return true;
}

bool FSocketNetworkNextClient::RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags)
{
	if (!this->bConnected || this->NetworkNextClient == nullptr)
	{
		return false;
	}

	if (this->PacketQueueSize == 0)
	{
		return false;
	}

	if (Flags != ESocketReceiveFlags::None)
	{
		return false;
	}

	PacketData NextPacket;
	if (!this->PacketQueue.Dequeue(NextPacket))
	{
		return false;
	}

	int CopySize = BufferSize;
	if (NextPacket.packet_bytes < CopySize)
	{
		CopySize = NextPacket.packet_bytes;
	}

	// Copy data from packet to buffer.
	memcpy(Data, NextPacket.packet_data, CopySize);
	BytesRead = CopySize;
	free((void*)NextPacket.packet_data);

	// We just assign the server address to the source, since it can only come from
	// the server in a Network Next client.
	bool bIsValid;
	Source.SetPort(this->ServerPort);
	Source.SetIp(*this->ServerAddr, bIsValid);

	// ServerAddr originally came from UE4, so it should always be valid.
	check(bIsValid);

	return true;
}

/**
 * Reads the address the socket is bound to and returns it
 * 
 * @param OutAddr address the socket is bound to
 */
void FSocketNetworkNextClient::GetAddress(FInternetAddr& OutAddr)
{
	// Dummy address for this interface.
	bool bIsValid;
	OutAddr.SetIp(TEXT("0.0.0.0"), bIsValid);
}

/**
 * Reads the port this socket is bound to.
 */ 
int32 FSocketNetworkNextClient::GetPortNo()
{
	// Dummy port for this interface.
	return 50000;
}
