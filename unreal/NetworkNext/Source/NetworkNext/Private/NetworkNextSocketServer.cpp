/*
	Network Next. Copyright 2017 - 2025 Network Next, Inc.  
	Licensed under the Network Next Source Available License 1.0
*/

#include "NetworkNextSocketServer.h"
#include "NetworkNextModule.h"
#include "next_server.h"
#include "next_address.h"
#include "next_platform.h"
#include "next_base64.h"

#ifdef SetPort
#undef SetPort
#endif // #ifdef SetPort

FNetworkNextSocketServer::FNetworkNextSocketServer(const FString& InSocketDescription, const FName& InSocketProtocol)
	: FNetworkNextSocket(ENetworkNextSocketType::TYPE_Server, InSocketDescription, InSocketProtocol)
{
	UE_LOG(LogNetworkNext, Display, TEXT("Server socket created"));
	NetworkNextServer = NULL;
	bUpdatedThisFrame = false;
}

FNetworkNextSocketServer::~FNetworkNextSocketServer()
{
	Close();
	UE_LOG(LogNetworkNext, Display, TEXT("Server socket destroyed"));
}

void FNetworkNextSocketServer::Update()
{
	// ...
}

bool FNetworkNextSocketServer::Close()
{
	if (NetworkNextServer)
	{
		next_server_destroy(NetworkNextServer);
		NetworkNextServer = NULL;
	}

	PacketData PassthroughPacket;
	while (PacketQueue.Dequeue(PassthroughPacket))
	{
		FMemory::Free(PassthroughPacket.packet_data);
	}

	UE_LOG(LogNetworkNext, Display, TEXT("Server socket closed"));

	return true;
}

bool FNetworkNextSocketServer::Bind(const FInternetAddr& Addr)
{
	Close();

	FString ServerAddress = Addr.ToString(true);

	FString ServerAddressOverride;
	if (FParse::Value(FCommandLine::Get(), TEXT("-serverAddress="), ServerAddressOverride))
	{
		ServerAddress = ServerAddressOverride;
	}

	UE_LOG(LogNetworkNext, Display, TEXT("Server address is '%s'"), *ServerAddress);

	FString datacenter = "local";
	FString datacenterOverride;
	if (FParse::Value(FCommandLine::Get(), TEXT("-datacenter="), datacenterOverride))
	{
		datacenter = datacenterOverride;
	}

	UE_LOG(LogNetworkNext, Display, TEXT("Datacenter is '%s'"), *datacenter);

	int port = Addr.GetPort();
	UE_LOG(LogNetworkNext, Display, TEXT("Server is binding to port %d"), port);
	char bind_address[64];
	snprintf(bind_address, sizeof(bind_address), "0.0.0.0:%d", Addr.GetPort());

	NetworkNextServer = next_server_create(this, TCHAR_TO_ANSI(*ServerAddress), bind_address, TCHAR_TO_ANSI(*datacenter), PacketReceived);
	if ( NetworkNextServer == NULL )
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Failed to create network next server"));
		return false;
	}

	UE_LOG(LogNetworkNext, Display, TEXT("Created network next server"));

	return true;
}

bool FNetworkNextSocketServer::SendTo(const uint8* Data, int32 Count, int32& BytesSent, const FInternetAddr& Destination)
{
	if (!NetworkNextServer)
		return false;
	
	next_address_t dest;
	if (next_address_parse(&dest, TCHAR_TO_ANSI(*(Destination.ToString(true)))) != NEXT_OK)
	{
		UE_LOG(LogNetworkNext, Warning, TEXT("Invalid address passed to FNetworkNextSocketServer::SendTo (%s)"), *Destination.ToString(true));
		return false;
	}

	next_server_send_packet(NetworkNextServer, &dest, Data, Count);
	
	BytesSent = Count;

	return true;
}

bool FNetworkNextSocketServer::RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags)
{
	if (!NetworkNextServer)
		return false;

	if (Flags != ESocketReceiveFlags::None)
		return false;

	if (!bUpdatedThisFrame)
	{
		// make sure we update the server prior to receiving any packets this frame
		next_server_update(NetworkNextServer);
		bUpdatedThisFrame = true;
	}

	PacketData Packet;
	if (!PacketQueue.Dequeue(Packet))
	{
		// we have finished receiving packets for this frame
		bUpdatedThisFrame = false;
		return false;
	}

	// drop packet if it is too large to copy to the recieve buffer
	if (Packet.packet_bytes > BufferSize)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Packet is too large to receive. Packet is %d bytes, but buffer is only %d bytes."), Packet.packet_bytes, BufferSize);
		FMemory::Free(Packet.packet_data);
		return false;
	}
	
	// Copy data from packet to buffer.
	memcpy(Data, Packet.packet_data, Packet.packet_bytes);
	BytesRead = Packet.packet_bytes;
	FMemory::Free(Packet.packet_data);

	// Convert next address to string.
	char next_address_string[NEXT_MAX_ADDRESS_STRING_LENGTH];
	next_address_to_string(&Packet.from, next_address_string);
	FString UnrealAddressString = FString(ANSI_TO_TCHAR(next_address_string));
	bool bIsValid;
	Source.SetIp(*UnrealAddressString, bIsValid);
	Source.SetPort(Packet.from.port);
	return bIsValid;
}

void FNetworkNextSocketServer::GetAddress(FInternetAddr& OutAddr)
{
	// We *always* bind to 0.0.0.0
	bool IsValid = false;
	OutAddr.SetIp(TEXT("0.0.0.0"), IsValid);
}

int32 FNetworkNextSocketServer::GetPortNo()
{
	// Return the port number that the server socket is listening on
	return NetworkNextServer ? next_server_port(NetworkNextServer) : 0;
}

void FNetworkNextSocketServer::UpgradeClient(TSharedPtr<const FInternetAddr> RemoteAddr, const FString& UserId)
{
	if (!NetworkNextServer)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("UpgradeClient called before the server socket was bound."));
		return;
	}

	if (!RemoteAddr.IsValid())
	{
		UE_LOG(LogNetworkNext, Error, TEXT("UpgradeClient called on an invalid RemoteAddr."));
		return;
	}

	FString ClientAddress = RemoteAddr.Get()->ToString(true);

	next_address_t from;
	if (next_address_parse(&from, TCHAR_TO_ANSI(*ClientAddress)) != NEXT_OK)
	{
		UE_LOG(LogNetworkNext, Warning, TEXT("UpgradeClient called with unparsable IP address: %s"), *ClientAddress);
		return;
	}

	next_server_upgrade_session(
		NetworkNextServer,
		&from,
		TCHAR_TO_ANSI(*UserId)
	);
}

void FNetworkNextSocketServer::PacketReceived(next_server_t* server, void* context, const next_address_t* client_address, const uint8_t* packet_data, int packet_bytes)
{
	// IMPORTANT: This is called from main thread inside next_server_update

	FNetworkNextSocketServer* self = (FNetworkNextSocketServer*)context;

	uint8_t* packet_data_copy = (uint8_t*)FMemory::Malloc(packet_bytes);

	memcpy(packet_data_copy, packet_data, packet_bytes);

	self->PacketQueue.Enqueue({ *client_address, packet_data_copy, packet_bytes });
}

bool FNetworkNextSocketServer::IsReady()
{
	if (NetworkNextServer == NULL)
		return false;

	return next_server_ready(NetworkNextServer);
}
