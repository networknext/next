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

#include "NetworkNextSocketClient.h"

#if !NEXT_PASSTHROUGH

FNetworkNextSocketClient::FNetworkNextSocketClient(const FString& InSocketDescription, const FName& InSocketProtocol)
    : FNetworkNextSocket(ENetworkNextSocketType::TYPE_Client, InSocketDescription, InSocketProtocol)
{
    UE_LOG(LogNetworkNext, Display, TEXT("Client socket created"));
    NetworkNextClient = NULL;
    ServerPort = NULL;
    bConnected = false;
}

FNetworkNextSocketClient::~FNetworkNextSocketClient()
{
    Close();
    UE_LOG(LogNetworkNext, Display, TEXT("Client socket destroyed"));
}

void FNetworkNextSocketClient::UpdateNetworkNextSocket()
{
    if (NetworkNextClient)
    {
        next_client_update(NetworkNextClient);
    }
}

bool FNetworkNextSocketClient::Close()
{
    if (NetworkNextClient)
    {
        next_client_destroy(NetworkNextClient);
        NetworkNextClient = NULL;
        ServerAddrAndPort = TEXT("");
        ServerAddr = TEXT("");
        ServerPort = 0;
        PacketQueue.Empty();
        UE_LOG(LogNetworkNext, Display, TEXT("Client socket closed"));
    }
    return true;
}

bool FNetworkNextSocketClient::Bind(const FInternetAddr& Addr)
{
    // We must ignore the local bind address and bind to 0.0.0.0 instead. XBoxOne binds to "::0" otherwise and breaks.
    int BindPort = Addr.GetPort();
    char BindAddress[256];
    sprintf_s(BindAddress, "0.0.0.0:%d", BindPort);

    UE_LOG(LogNetworkNext, Display, TEXT("Bind Client Socket (%s)"), ANSI_TO_TCHAR(BindAddress));

    if (NetworkNextClient)
    {
        Close();
    }

    NetworkNextClient = next_client_create(this, BindAddress, &FNetworkNextSocketClient::OnPacketReceived, NULL);
    if (NetworkNextClient == NULL)
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Failed to create network next client"));
        return false;
    }

    UE_LOG(LogNetworkNext, Display, TEXT("Created network next client"));

    const int ClientPort = next_client_port(NetworkNextClient);

    UE_LOG(LogNetworkNext, Display, TEXT("Client port is %d"), ClientPort);

    return true;
}

static FString IntToIPv4String(uint32 ipAsInt)
{
    return FString::Printf(TEXT("%d.%d.%d.%d"), (ipAsInt >> 24), ((ipAsInt >> 16) & 0xff), ((ipAsInt >> 8) & 0xff), (ipAsInt & 0xff));
}

bool FNetworkNextSocketClient::SendTo(const uint8* Data, int32 Count, int32& BytesSent, const FInternetAddr& Destination)
{
#if PLATFORM_HAS_BSD_IPV6_SOCKETS

    /*
    *	When PLATFORM_HAS_BSD_IPV6_SOCKETS is set, Destination is of type FInternetAddrBSDIPv6.
    *	FInternetAddrBSDIPv6 converts IPv4 addresses to IPv6, so we convert them back to IPv4 here.
    */

    uint32 ipAddressAsInt = 0;
    Destination.GetIp(ipAddressAsInt);

    FString serverAddr = IntToIPv4String(ipAddressAsInt);
    int32 serverPort = Destination.GetPort();
    FString serverAddrAndPort = FString::Printf(TEXT("%s:%d"), *serverAddr, serverPort);
    
#else

    FString serverAddr = Destination.ToString(false);
    int32 serverPort = Destination.GetPort();
    FString serverAddrAndPort = Destination.ToString(true);

#endif

    if (!NetworkNextClient)
        return false;

    // The first send indicates the server we want to connect to
    if (!bConnected)
    {
        ServerAddrAndPort = serverAddrAndPort;
        ServerAddr = serverAddr;
        ServerPort = serverPort;
        bConnected = true;

        next_client_open_session(NetworkNextClient, TCHAR_TO_ANSI(*ServerAddrAndPort));
    }

    if (!serverAddrAndPort.Equals(ServerAddrAndPort))
    {
        // The client socket can *only* send to the server address
        UE_LOG(LogNetworkNext, Error, TEXT("Attempted to use client socket to send data to %s, but it can only send data to the server %s"), *serverAddrAndPort, *ServerAddrAndPort);
        return false;
    }

    next_client_send_packet(NetworkNextClient, Data, Count);

    BytesSent = Count;

    return true;
}

void FNetworkNextSocketClient::OnPacketReceived(next_client_t* client, void* context, const uint8_t* packet_data, int packet_bytes)
{
    FNetworkNextSocketClient* self = (FNetworkNextSocketClient*)context;

    uint8_t* packet_data_copy = (uint8_t*)malloc(packet_bytes);
    
    memcpy(packet_data_copy, packet_data, packet_bytes);

    self->PacketQueue.Enqueue({
        packet_data_copy,
        packet_bytes,
        });
}

bool FNetworkNextSocketClient::RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags)
{
    if (!NetworkNextClient)
        return false;

    if (!bConnected)
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

    // Packects *only* come from the server
    bool bIsValid;
    Source.SetPort(ServerPort);
    Source.SetIp(*ServerAddr, bIsValid);
    check(bIsValid);
    return true;
}

void FNetworkNextSocketClient::GetAddress(FInternetAddr& OutAddr)
{
    // We *always* bind to "0.0.0.0"
    bool bIsValid;
    OutAddr.SetIp(TEXT("0.0.0.0"), bIsValid);
}

int32 FNetworkNextSocketClient::GetPortNo()
{
    // Return the actual port number the socket is bound to. This may be a system assigned port if the bind port was 0.
    return NetworkNextClient ? next_client_port(NetworkNextClient) : 0;
}

#endif // #if !NEXT_PASSTHROUGH
