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

#include "NetworkNextSocketServer.h"

#if !NEXT_PASSTHROUGH

FNetworkNextSocketServer::FNetworkNextSocketServer(const FString& InSocketDescription, const FName& InSocketProtocol)
    : FNetworkNextSocket(ENetworkNextSocketType::TYPE_Server, InSocketDescription, InSocketProtocol)
{
    UE_LOG(LogNetworkNext, Display, TEXT("Server socket created"));
    NetworkNextServer = NULL;
}

FNetworkNextSocketServer::~FNetworkNextSocketServer()
{
    Close();
    UE_LOG(LogNetworkNext, Display, TEXT("Server socket destroyed"));
}

void FNetworkNextSocketServer::UpdateNetworkNextSocket()
{
    if (NetworkNextServer)
    {
        next_server_update(NetworkNextServer);
    }
}

bool FNetworkNextSocketServer::Close()
{
    if (NetworkNextServer)
    {
        next_server_destroy(NetworkNextServer);
        NetworkNextServer = NULL;
        ServerAddress = "";
        PacketQueue.Empty();
        UE_LOG(LogNetworkNext, Display, TEXT("Server socket closed"));
    }
    return true;
}

static bool ExtractServerAddressOnly(const FString & ServerAddressWithPort, FString & ServerAddressOnly)
{
    int32 LastColon;
    if (!ServerAddressWithPort.FindLastChar(TEXT(":")[0], LastColon))
    {
        return false;
    }        

    if (ServerAddressWithPort[0] == TCHAR('['))
    {
        // ipv6 in network next form, eg. [::1]:20000
        ServerAddressOnly = *(ServerAddressWithPort.Mid(1, LastColon - 2)); /* for the brackets */
    }
    else
    {
        // ipv4 in network next form, eg. 127.0.0.1:20000
        ServerAddressOnly = *(ServerAddressWithPort.Mid(0, LastColon));
    }

    return true;
}

bool FNetworkNextSocketServer::Bind(const FInternetAddr& Addr)
{
    Close();

    UE_LOG(LogNetworkNext, Display, TEXT("Bind Server Socket (%s)"), *Addr.ToString(true));

    FString BindAddress = Addr.ToString(true);
    FString ServerAddressWithPort = FString::Printf(TEXT("127.0.0.1:%d"), Addr.GetPort());
    FString DataCenter = "local";

    FString ServerAddressOnly;
    if (!ExtractServerAddressOnly(ServerAddressWithPort, ServerAddressOnly))
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Could not extract server address only"));
        return false;
    }

    NetworkNextServer = next_server_create(this, TCHAR_TO_ANSI(*ServerAddressWithPort), TCHAR_TO_ANSI(*BindAddress), TCHAR_TO_ANSI(*DataCenter), &FNetworkNextSocketServer::OnPacketReceived, NULL);
    if (!NetworkNextServer)
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Failed to create network next server"));
        return false;
    }

    UE_LOG(LogNetworkNext, Display, TEXT("Created network next server"));

    ServerAddress = ServerAddressOnly;

    UE_LOG(LogNetworkNext, Display, TEXT("Server address is %s"), *ServerAddress);

    UE_LOG(LogNetworkNext, Display, TEXT("Server port is %d"), GetPortNo());

    return true;
}

bool FNetworkNextSocketServer::SendTo(const uint8* Data, int32 Count, int32& BytesSent, const FInternetAddr& Destination)
{
    if (!NetworkNextServer)
        return false;
    
    next_address_t dest;
    if (next_address_parse(&dest, TCHAR_TO_ANSI(*(Destination.ToString(true)))) != NEXT_OK)
        return false;

    next_server_send_packet(NetworkNextServer, &dest, Data, Count);
    
    BytesSent = Count;

    return true;
}

void FNetworkNextSocketServer::OnPacketReceived(next_server_t* server, void* context, const next_address_t* from, const uint8_t* packet_data, int packet_bytes)
{
    FNetworkNextSocketServer* self = (FNetworkNextSocketServer*)context;

    uint8_t* packet_data_copy = (uint8_t*)malloc(packet_bytes);

    memcpy(packet_data_copy, packet_data, packet_bytes);

    self->PacketQueue.Enqueue({
        *from,
        packet_data_copy,
        packet_bytes,
        });
}

bool FNetworkNextSocketServer::RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags)
{
    if (!NetworkNextServer)
        return false;

    if (PacketQueueBytes == 0)
        return false;

    if (Flags != ESocketReceiveFlags::None)
        return false;

    PacketData NextPacket;
    if (!PacketQueue.Dequeue(NextPacket))
        return false;

    int CopySize = BufferSize;
    if (NextPacket.packet_bytes < CopySize)
        CopySize = NextPacket.packet_bytes;

    // Copy data from packet to buffer.
    memcpy(Data, NextPacket.packet_data, CopySize);
    BytesRead = CopySize;
    free((void*)NextPacket.packet_data);

    // Convert Network Next address to string.
    char next_address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_address_to_string(&NextPacket.from, next_address_buffer);
    uint8_t address_type = NextPacket.from.type;
    uint16_t address_port = NextPacket.from.port;

    // Now manually parse the from address string, since UE4 does not support parsing the address and port combined.
    FString NextAddressAsUE4String = FString(ANSI_TO_TCHAR(next_address_buffer));
    int32 LastColon;
    if (!NextAddressAsUE4String.FindLastChar(TEXT(":")[0], LastColon))
        return false;
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
    return bIsValid;
}

void FNetworkNextSocketServer::GetAddress(FInternetAddr& OutAddr)
{
    if (NetworkNextServer)
    {
        // Return the address the server socket is listening on
        bool IsValid = false;
        OutAddr.SetIp(*ServerAddress, IsValid);
    }
    else
    {
        // Not bound yet. We don't have any address!
        bool IsValid = false;
        OutAddr.SetIp(TEXT("0.0.0.0"), IsValid);
    }
}

int32 FNetworkNextSocketServer::GetPortNo()
{
    // Return the port number that the server socket is listening on
    return NetworkNextServer ? next_server_port(NetworkNextServer) : 0;
}

#endif // #if !NEXT_PASSTHROUGH