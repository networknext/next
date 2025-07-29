/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "NetworkNextConnection.h"
#include "NetworkNextModule.h"

void UNetworkNextConnection::InitRemoteConnection(class UNetDriver* InDriver, class FSocket* InSocket, const FURL& InURL, const class FInternetAddr& InRemoteAddr, EConnectionState InState, int32 InMaxPacket, int32 InPacketOverhead)
{
    UE_LOG(LogNetworkNext, Display, TEXT("NetworkNextConnection::InitRemoteConnection"));

    DisableAddressResolution();

    Super::InitRemoteConnection(InDriver, InSocket, InURL, InRemoteAddr, InState, InMaxPacket, InPacketOverhead);
}

void UNetworkNextConnection::InitLocalConnection(UNetDriver* InDriver, class FSocket* InSocket, const FURL& InURL, EConnectionState InState, int32 InMaxPacket, int32 InPacketOverhead)
{
    UE_LOG(LogNetworkNext, Display, TEXT("NetworkNextConnection::InitLocalConnection"));

    DisableAddressResolution();

    Super::InitLocalConnection(InDriver, InSocket, InURL, InState, InMaxPacket, InPacketOverhead);
}
