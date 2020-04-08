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

#include "SocketNetworkNextServer.h"
#include "NetworkNextNetDriver.h"
#include "AssertionMacros.h"
#include "NetworkNext.h"
#include "IPAddress.h"
#include "NetworkNextServerConfig.h"

#if defined(NETWORKNEXT_HAS_ESOCKETPROTOCOLFAMILY)
FSocketNetworkNextServer::FSocketNetworkNextServer(const FString& InSocketDescription, ESocketProtocolFamily InSocketProtocol, const UNetworkNextNetDriver* InNetDriver) :
	FSocketNetworkNext(InSocketDescription, InSocketProtocol, ESocketNetworkNextType::TYPE_Server)
#else
FSocketNetworkNextServer::FSocketNetworkNextServer(const FString& InSocketDescription, const UNetworkNextNetDriver* InNetDriver) :
	FSocketNetworkNext(InSocketDescription, ESocketNetworkNextType::TYPE_Server)
#endif
{
	check(InNetDriver);

	this->NetDriver = InNetDriver;
	this->NetworkNextServer = nullptr;
	this->PacketQueueSize = 0;
	this->bBound = false;
	this->ServerPort = 0;

	UE_LOG(LogNetworkNext, Display, TEXT("FSocketNetworkNextServer Created"));
}

FSocketNetworkNextServer::~FSocketNetworkNextServer()
{
	this->Close();

	UE_LOG(LogNetworkNext, Display, TEXT("FSocketNetworkNextServer Destroyed"));
}

void FSocketNetworkNextServer::UpdateNetworkNextSocket()
{
	if (this->bBound)
	{
		next_server_update(this->NetworkNextServer);
	}
}

void FSocketNetworkNextServer::UpgradeClient(const TSharedPtr<FInternetAddr>& RemoteAddr, const FString& UserId, ENetworkNextPlatformType Platform, const FString& Tag)
{
	UE_LOG(LogNetworkNext, Display, TEXT("Upgrade Client"));

	if (!this->bBound)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("UpgradeClient called on SocketNetworkNextServer that was not yet bound."));
		return;
	}

	if (!RemoteAddr.IsValid())
	{
		UE_LOG(LogNetworkNext, Error, TEXT("UpgradeClient called on SocketNetworkNextServer with an invalid RemoteAddr."));
		return;
	}

	FString ClientAddress = RemoteAddr.Get()->ToString(true);

	uint32_t next_platform;
	switch (Platform)
	{
	case ENetworkNextPlatformType::PlatformType_Windows:
		next_platform = NEXT_PLATFORM_WINDOWS;
		break;
	case ENetworkNextPlatformType::PlatformType_Mac:
		next_platform = NEXT_PLATFORM_MAC;
		break;
	case ENetworkNextPlatformType::PlatformType_Linux:
		next_platform = NEXT_PLATFORM_LINUX;
		break;
	case ENetworkNextPlatformType::PlatformType_Switch:
		next_platform = NEXT_PLATFORM_SWITCH;
		break;
	case ENetworkNextPlatformType::PlatformType_PS4:
		next_platform = NEXT_PLATFORM_PS4;
		break;
	case ENetworkNextPlatformType::PlatformType_XboxOne:
		next_platform = NEXT_PLATFORM_XBOX_ONE;
		break;
	case ENetworkNextPlatformType::PlatformType_iOS:
		next_platform = NEXT_PLATFORM_IOS;
		break;
	default:
	case ENetworkNextPlatformType::PlatformType_Unknown:
		next_platform = 0 /* Unknown */;
		break;
	}

	next_address_t from;
	if (next_address_parse(&from, TCHAR_TO_ANSI(*ClientAddress)) != NEXT_OK)
	{
		UE_LOG(LogNetworkNext, Warning, TEXT("UpgradeClient called with unparsable IP address: %s"), *ClientAddress);
		return;
	}

	next_server_upgrade_session(
		this->NetworkNextServer,
		&from,
		TCHAR_TO_ANSI(*UserId),
		next_platform,
		TCHAR_TO_ANSI(*Tag)
	);
}

void FSocketNetworkNextServer::OnPacketReceived(next_server_t* server, void* context, const next_address_t* from, const uint8_t* packet_data, int packet_bytes)
{
	FSocketNetworkNextServer* self = (FSocketNetworkNextServer*)context;

	next_address_t* next_address_copy = (next_address_t*)malloc(sizeof(next_address_t));
	memcpy(next_address_copy, from, sizeof(next_address_t));

	uint8_t* packet_data_copy = (uint8_t*)malloc(packet_bytes);
	memcpy(packet_data_copy, packet_data, packet_bytes);

	self->PacketQueue.Enqueue({
		next_address_copy,
		packet_data_copy,
		packet_bytes,
		});
	self->PacketQueueSize += packet_bytes;
}

bool FSocketNetworkNextServer::Close()
{
	UE_LOG(LogNetworkNext, Display, TEXT("Close Server Socket"));

	if (this->bBound)
	{
		next_server_destroy(this->NetworkNextServer);
		this->NetworkNextServer = nullptr;
		this->bBound = false;
		this->ServerPort = 0;
		this->ServerAddress = "";
	}

	return true;
}

bool FSocketNetworkNextServer::Bind(const FInternetAddr& Addr)
{
	if (this->bBound)
	{
		next_server_destroy(this->NetworkNextServer);
		this->NetworkNextServer = nullptr;
		this->bBound = false;
		this->ServerPort = 0;
		this->ServerAddress = "";
	}

	UE_LOG(LogNetworkNext, Display, TEXT("Bind Server Socket (%s)"), *Addr.ToString(true));

	FNetworkNextModule& NetworkNextModule = FModuleManager::LoadModuleChecked<FNetworkNextModule>("NetworkNext");

	const FNetworkNextServerConfig* ServerConfig = NetworkNextModule.GetServerConfig();

	// Setup defaults
	FString BindAddr = Addr.ToString(true);
	FString ServerAddr = FString::Printf(TEXT("127.0.0.1:%d"), Addr.GetPort());
	FString DataCenter = "local";

	// Use server config values if set
	if (ServerConfig != nullptr)
	{
		if (ServerConfig->ServerAddress.Len() > 0)
		{
			ServerAddr = ServerConfig->ServerAddress;
		}

		if (ServerConfig->Datacenter.Len() > 0)
		{
			DataCenter = ServerConfig->Datacenter;
		}
	}

	this->NetworkNextServer = next_server_create(
		this,
		TCHAR_TO_ANSI(*ServerAddr),
		TCHAR_TO_ANSI(*BindAddr),
		TCHAR_TO_ANSI(*DataCenter),
		&FSocketNetworkNextServer::OnPacketReceived
	);

	if (!this->NetworkNextServer)
	{
		// We could not create the server, do not set bBound to true and
		// return false to indicate that we couldn't start listening.
		UE_LOG(LogNetworkNext, Error, TEXT("Failed to create network next server"));
		return false;
	}

	this->bBound = true;
	this->ServerPort = next_server_port(NetworkNextServer);

	int32 LastColon;
	if (!ServerAddr.FindLastChar(TEXT(":")[0], LastColon))
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Failed split address and port"));
		return false;
	}

	if (ServerAddr[0] == TCHAR('['))
	{
		// ipv6 in network next form, eg. [::1]:20000
		this->ServerAddress = *(ServerAddr.Mid(1, LastColon - 2)); /* for the brackets */
	}
	else
	{
		// ipv4 in network next form, eg. 127.0.0.1:20000
		this->ServerAddress = *(ServerAddr.Mid(0, LastColon));
	}

	UE_LOG(LogNetworkNext, Display, TEXT("ServerAddress is %s"), *this->ServerAddress);
	UE_LOG(LogNetworkNext, Display, TEXT("ServerPort is %d"), this->ServerPort);

	return true;
}

bool FSocketNetworkNextServer::HasPendingData(uint32& PendingDataSize)
{
	if (!this->bBound)
	{
		PendingDataSize = 0;
		return false;
	}

	PendingDataSize = this->PacketQueueSize;
	return this->PacketQueueSize > 0;
}

bool FSocketNetworkNextServer::SendTo(const uint8* Data, int32 Count, int32& BytesSent, const FInternetAddr& Destination)
{
	if (!this->bBound)
	{
		return false;
	}

	next_address_t dest;
	if (next_address_parse(&dest, TCHAR_TO_ANSI(*(Destination.ToString(true)))) != NEXT_OK)
	{
		return false;
	}

	next_server_send_packet(
		this->NetworkNextServer,
		&dest,
		Data,
		Count
	);
	BytesSent = Count;

	return true;
}

bool FSocketNetworkNextServer::RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags)
{
	if (!this->bBound)
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

	// Convert Network Next address to string.
	char next_address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
	next_address_to_string(NextPacket.from, next_address_buffer);
	uint8_t address_type = NextPacket.from->type;
	uint16_t address_port = NextPacket.from->port;
	free((void*)NextPacket.from);

	// Now try to parse the string, since UE4 does not support passing the address and port combined.
	FString NextAddressAsUE4String = FString(ANSI_TO_TCHAR(next_address_buffer));
	int32 LastColon;
	if (!NextAddressAsUE4String.FindLastChar(TEXT(":")[0], LastColon))
	{
		return false;
	}

	bool bIsValid = false;
	switch (address_type)
	{
	case NEXT_ADDRESS_NONE:
		return false;
	case NEXT_ADDRESS_IPV4:
		Source.SetIp(*(NextAddressAsUE4String.Mid(0, LastColon)), bIsValid);
		Source.SetPort(address_port);
		break;
	case NEXT_ADDRESS_IPV6:
		Source.SetIp(*(NextAddressAsUE4String.Mid(1, LastColon - 2) /* for the brackets */), bIsValid);
		Source.SetPort(address_port);
		break;
	}

	if (!bIsValid)
	{
		return false;
	}

	return true;
}

/**
 * Reads the address the socket is bound to and returns it
 *
 * @param OutAddr address the socket is bound to
 */
void FSocketNetworkNextServer::GetAddress(FInternetAddr& OutAddr)
{
	bool IsValid = false;
	OutAddr.SetIp(*ServerAddress, IsValid);
	OutAddr.SetPort(ServerPort);
}

/**
 * Reads the port this socket is bound to.
 */
int32 FSocketNetworkNextServer::GetPortNo()
{
	return ServerPort;
}
