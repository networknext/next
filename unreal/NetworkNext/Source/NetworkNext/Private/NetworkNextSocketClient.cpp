/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "NetworkNextSocketClient.h"
#include "NetworkNextModule.h"
#include "next_client.h"
#include "next_platform.h"
#include "next_crypto.h"

FNetworkNextSocketClient::FNetworkNextSocketClient(const FString& InSocketDescription, const FName& InSocketProtocol)
    : FNetworkNextSocket(ENetworkNextSocketType::TYPE_Client, InSocketDescription, InSocketProtocol)
{
    UE_LOG(LogNetworkNext, Display, TEXT("Client socket created"));
    NetworkNextClient = NULL;
    ServerPort = 0;
    bConnected = false;
    bUpdatedThisFrame = false;
}

FNetworkNextSocketClient::~FNetworkNextSocketClient()
{
    Close();
    UE_LOG(LogNetworkNext, Display, TEXT("Client socket destroyed"));
}

void FNetworkNextSocketClient::Update()
{
    // ...
}

bool FNetworkNextSocketClient::Close()
{
    if (NetworkNextClient)
    {
        next_client_destroy(NetworkNextClient);
        NetworkNextClient = NULL;
    }

    PacketData PassthroughPacket;
    while (PacketQueue.Dequeue(PassthroughPacket))
    {
        FMemory::Free(PassthroughPacket.packet_data);
    }

    UE_LOG(LogNetworkNext, Display, TEXT("Client socket closed"));

    ServerAddr = TEXT("");
    ServerPort = 0;

    return true;
}

bool FNetworkNextSocketClient::Bind(const FInternetAddr& Addr)
{
    // Always bind client sockets to 0.0.0.0:0. The Network Next SDK takes care of enabling dual stack where possible *and* selecting the best port.
    // On GDK platforms (XBoxOne, SeriesX), the preferred game port is used. This provides access to WMM and DSCP packet tagging on Xbox platforms.
    char BindAddress[256];
    TCString<char>::Sprintf(BindAddress, "0.0.0.0:0");

    UE_LOG(LogNetworkNext, Display, TEXT("Bind Client Socket (%s)"), ANSI_TO_TCHAR(BindAddress));

    if (NetworkNextClient)
    {
        Close();
    }

    NetworkNextClient = next_client_create( this, BindAddress, PacketReceived );
    if ( NetworkNextClient == NULL )
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Failed to create network next client"));
        return false;
    }

    UE_LOG(LogNetworkNext, Display, TEXT("Created network next client"));

    const int ClientPort = next_client_port(NetworkNextClient);

    UE_LOG(LogNetworkNext, Display, TEXT("Client port is %d"), ClientPort);

    return true;
}

void FNetworkNextSocketClient::NetworkNextConnect(const FString& Host, int32 Port)
{
    // connect to server

    if (!NetworkNextClient)
    {
        next_printf(NEXT_LOG_LEVEL_ERROR, "NULL client in network next connect");
        return;
    }

    char server_address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
    memset(server_address_buffer, 0, sizeof(server_address_buffer));
    snprintf(server_address_buffer, sizeof(server_address_buffer), "%s:%d", TCHAR_TO_ANSI(*Host), Port);

    const char* server_address = &server_address_buffer[0];

    next_client_open_session(NetworkNextClient, server_address);

    ServerAddr = Host;
    ServerPort = Port;
    bConnected = true;
}

bool FNetworkNextSocketClient::SendTo(const uint8* Data, int32 Count, int32& BytesSent, const FInternetAddr& Destination)
{
    if (!NetworkNextClient)
        return false;

    next_client_send_packet(NetworkNextClient, Data, Count);

    BytesSent = Count;

    return true;
}

bool FNetworkNextSocketClient::RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags)
{
    if (!NetworkNextClient)
        return false;

    if (!bConnected)
        return false;

    if (Flags != ESocketReceiveFlags::None)
        return false;

    if (!bUpdatedThisFrame)
    {
        // make sure we update the client prior to receiving any packets this frame
        next_client_update(NetworkNextClient);
        bUpdatedThisFrame = true;
    }

    PacketData Packet;
    if (!PacketQueue.Dequeue(Packet))
    {
        // we have finished packets for this frame
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

    // Packets *only* come from the server
    #undef SetPort
    bool bIsValid;
    Source.SetPort(ServerPort);
    Source.SetIp(*ServerAddr, bIsValid);
    check(bIsValid);

    return true;
}

void FNetworkNextSocketClient::GetAddress(FInternetAddr& OutAddr)
{
    // We *always* bind to "0.0.0.0". The Network Next SDK takes care of dual stack transparently.
    bool bIsValid;
    OutAddr.SetIp(TEXT("0.0.0.0"), bIsValid);
}

int32 FNetworkNextSocketClient::GetPortNo()
{
    // Return the actual port number the socket is bound to. This may be a system assigned port if the bind port was 0.
    return NetworkNextClient ? next_client_port(NetworkNextClient) : 0;
}

float FNetworkNextSocketClient::GetLatency()
{
    if (NetworkNextClient)
    {
        const next_client_stats_t* stats = next_client_stats(NetworkNextClient);
        if (stats->next)
        {
            return stats->next_rtt;
        }
        else
        {
            return stats->direct_rtt;
        }
    }
    else
    {
        return 0.0;
    }
}

float FNetworkNextSocketClient::GetJitter()
{
    if (NetworkNextClient)
    {
        const next_client_stats_t* stats = next_client_stats(NetworkNextClient);
        return (stats->jitter_client_to_server + stats->jitter_server_to_client) / 2.0f;
    }
    else
    {
        return 0.0;
    }
}

float FNetworkNextSocketClient::GetPacketLoss()
{
    if (NetworkNextClient)
    {
        const next_client_stats_t* stats = next_client_stats(NetworkNextClient);
        double total_packets_sent = (double) ( stats->packets_sent_client_to_server + stats->packets_sent_server_to_client );
        double total_packets_lost = (double) ( stats->packets_lost_client_to_server + stats->packets_lost_server_to_client );
        if (total_packets_sent > 0)
        {
            return float(total_packets_lost / total_packets_sent) * 100.0f;
        }
    }
    return 0.0;
}

void FNetworkNextSocketClient::PacketReceived(next_client_t* client, void* context, const next_address_t *from, const uint8_t* packet_data, int packet_bytes)
{
    // This callback is pumped on main thread from inside next_client_update

    FNetworkNextSocketClient* self = (FNetworkNextSocketClient*)context;

    uint8_t* packet_data_copy = (uint8_t*)FMemory::Malloc(packet_bytes);

    memcpy(packet_data_copy, packet_data, packet_bytes);

    self->PacketQueue.Enqueue({ packet_data_copy, packet_bytes });
}
