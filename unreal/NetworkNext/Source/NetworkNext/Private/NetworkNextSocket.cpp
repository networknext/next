/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "NetworkNextSocket.h"
#include "NetworkNextModule.h"

FNetworkNextSocket::FNetworkNextSocket(ENetworkNextSocketType InSocketType, const FString& InSocketDescription, FName InSocketProtocol)
    : FSocket(SOCKTYPE_Datagram, InSocketDescription, InSocketProtocol)
{
    this->NetworkNextSocketType = InSocketType;
}

bool FNetworkNextSocket::Shutdown(ESocketShutdownMode Mode)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::Shutdown is not supported"))
    return false;
}

bool FNetworkNextSocket::Connect(const FInternetAddr& Addr)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::Connect is not supported"))
    return false;
}

bool FNetworkNextSocket::Listen(int32 MaxBacklog)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::Listen is not supported"))
    return false;
}

bool FNetworkNextSocket::WaitForPendingConnection(bool& bHasPendingConnection, const FTimespan& WaitTime)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::WaitForPendingConnection is not supported"))
    return false;
}

bool FNetworkNextSocket::HasPendingData(uint32& PendingDataSize)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::HasPendingData is not supported"))
    return false;
}


FSocket* FNetworkNextSocket::Accept(const FString& InSocketDescription)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::Accept is not supported"))
    return NULL;
}

FSocket* FNetworkNextSocket::Accept(FInternetAddr& OutAddr, const FString& InSocketDescription)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::Accept is not supported"))
    return NULL;
}

bool FNetworkNextSocket::Send(const uint8* Data, int32 Count, int32& BytesSent)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::Send is not supported"))
    BytesSent = 0;
    return false;
}

bool FNetworkNextSocket::Recv(uint8* Data, int32 BufferSize, int32& BytesRead, ESocketReceiveFlags::Type Flags)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::Recv is not supported"))
    BytesRead = 0;
    return false;
}

bool FNetworkNextSocket::Wait(ESocketWaitConditions::Type Condition, FTimespan WaitTime)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::Wait is not supported"))
    return false;
}

ESocketConnectionState FNetworkNextSocket::GetConnectionState()
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::GetConnectionState is not supported"))
    return SCS_NotConnected;
}

bool FNetworkNextSocket::GetPeerAddress(FInternetAddr& OutAddr)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::GetPeerAddress is not supported"))
    return false;
}

bool FNetworkNextSocket::SetNonBlocking(bool bIsNonBlocking)
{
    // *** IGNORED ***
    return true;
}

bool FNetworkNextSocket::SetBroadcast(bool bAllowBroadcast)
{
    // *** IGNORED ***
    return true;
}

bool FNetworkNextSocket::JoinMulticastGroup(const FInternetAddr& GroupAddress)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::JoinMulticastGroup is not supported"))
    return false;
}

bool FNetworkNextSocket::LeaveMulticastGroup(const FInternetAddr& GroupAddress)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::LeaveMulticastGroup is not supported"))
    return false;
}

bool FNetworkNextSocket::JoinMulticastGroup(const FInternetAddr& GroupAddress, const FInternetAddr& InterfaceAddress)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::JoinMulticastGroup is not supported"))
    return false;
}

bool FNetworkNextSocket::LeaveMulticastGroup(const FInternetAddr& GroupAddress, const FInternetAddr& InterfaceAddress)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::LeaveMulticastGroup is not supported"))
    return false;
}

bool FNetworkNextSocket::SetMulticastLoopback(bool bLoopback)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::SetMulticastLoopback is not supported"))
    return false;
}

bool FNetworkNextSocket::SetMulticastTtl(uint8 TimeToLive)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::SetMulticastTtl is not supported"))
    return false;
}

bool FNetworkNextSocket::SetMulticastInterface(const FInternetAddr& InterfaceAddress)
{
    // *** NOT SUPPORTED ***
    UE_LOG(LogNetworkNext, Warning, TEXT("FNetworkNextSocket::SetMulticastInterface is not supported"))
    return false;
}

bool FNetworkNextSocket::SetReuseAddr(bool bAllowReuse)
{
    // *** IGNORED ***
    return true;
}

bool FNetworkNextSocket::SetLinger(bool bShouldLinger, int32 Timeout)
{
    // *** IGNORED ***
    return true;
}

bool FNetworkNextSocket::SetRecvErr(bool bUseErrorQueue)
{
    // *** IGNORED ***
    return true;
}

bool FNetworkNextSocket::SetSendBufferSize(int32 Size, int32& NewSize)
{
    // *** IGNORED ***
    return true;
}

bool FNetworkNextSocket::SetReceiveBufferSize(int32 Size, int32& NewSize)
{
    // *** IGNORED ***
    return true;
}

bool FNetworkNextSocket::SetNoDelay(bool flag)
{
    // *** IGNORED ***
    return false;
}
