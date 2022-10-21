/*
    Network Next SDK. Copyright © 2017 - 2022 Network Next, Inc.

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
#include "next.h"

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

    // Multiplay support
    {
        FString ConfigJsonFilePath;
        if (FParse::Value(FCommandLine::Get(), TEXT("-serverjson="), ConfigJsonFilePath))
        {
            FString FilePath = ConfigJsonFilePath;

            // Check if the JSON file exists and has the session id
            FString FileContents;
            if (!FFileHelper::LoadFileToString(FileContents, *FilePath))
            {
                UE_LOG(LogNetworkNext, Error, TEXT("Failed to load multiplay server.json file '%s'"), *FilePath);
            }
            else
            {
                TSharedPtr<FJsonObject> JSONObject;

                TSharedRef< TJsonReader<> > Reader = TJsonReaderFactory<>::Create(FileContents);

                if (!FJsonSerializer::Deserialize(Reader, JSONObject) || !JSONObject.IsValid())
                {
                    UE_LOG(LogNetworkNext, Error, TEXT("Invalid JSON in multiplay server.json file '%s'"), *FilePath);
                }
                else
                {
                    UE_LOG(LogNetworkNext, Display, TEXT("Loaded multiplay server.json file"));

                    FString ConnectionIp;
                    FString ConnectionPort;

                    if (JSONObject->HasTypedField<EJson::String>("ConnectionIP"))
                    {
                        ConnectionIp = JSONObject->GetStringField("ConnectionIP");
                    }

                    if (JSONObject->HasTypedField<EJson::String>("ConnectionPort"))
                    {
                        ConnectionPort = JSONObject->GetStringField("ConnectionPort");
                    }

                    ServerAddressWithPort = ConnectionIp + TEXT(":") + ConnectionPort;

                    UE_LOG(LogNetworkNext, Display, TEXT("Multiplay server address is '%s'"), *Addr.ToString(true));
                }
            }
        }

        FString VirtualTypeId;
        bool bIsCloud = false;
        if (FParse::Value(FCommandLine::Get(), TEXT("-virtualtypeid="), VirtualTypeId))
        {
            if (VirtualTypeId == TEXT("3"))
            {
                UE_LOG(LogNetworkNext, Display, TEXT("Multiplay server is running in the cloud"));
                DataCenter = TEXT("cloud");
                bIsCloud = true;
            }
        }

        FString MultiplayDatacenter;
        if (!bIsCloud && FParse::Value(FCommandLine::Get(), TEXT("-NetworkNext="), MultiplayDatacenter))
        {
            UE_LOG(LogNetworkNext, Display, TEXT("Multiplay datacenter is '%s'"), MultiplayDatacenter);
            DataCenter = MultiplayDatacenter;
        }
    }

    // create the network next server

    NetworkNextServer = next_server_create(this, TCHAR_TO_ANSI(*ServerAddressWithPort), TCHAR_TO_ANSI(*BindAddress), TCHAR_TO_ANSI(*DataCenter), &FNetworkNextSocketServer::OnPacketReceived, NULL);
    if (!NetworkNextServer)
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
        return false;

    next_server_send_packet(NetworkNextServer, &dest, Data, Count);
    
    BytesSent = Count;

    return true;
}

void FNetworkNextSocketServer::OnPacketReceived(next_server_t* server, void* context, const next_address_t* from, const uint8_t* packet_data, int packet_bytes)
{
    // IMPORTANT: This is called from main thread inside next_server_update

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

    if (Flags != ESocketReceiveFlags::None)
        return false;

    if (!UpdatedThisFrame)
    {
        // make sure we update the server prior to receiving any packets this frame
        next_server_update(NetworkNextServer);
        bUpdatedThisFrame = true;
    }

    PacketData NextPacket;
    if (!PacketQueue.Dequeue(NextPacket))
    {
        // we have finished receiving packets for this frame
        bUpdatedThisFrame = false;
        return false;
    }

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
