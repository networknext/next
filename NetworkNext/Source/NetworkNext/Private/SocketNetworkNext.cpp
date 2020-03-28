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

#include "SocketNetworkNext.h"

#if defined(NETWORKNEXT_HAS_ESOCKETPROTOCOLFAMILY)
FSocketNetworkNext::FSocketNetworkNext(const FString& InSocketDescription, ESocketProtocolFamily InSocketProtocol, ESocketNetworkNextType InNetworkNextType) :
	FSocket(SOCKTYPE_Datagram, InSocketDescription, InSocketProtocol)
#else
FSocketNetworkNext::FSocketNetworkNext(const FString& InSocketDescription, ESocketNetworkNextType InNetworkNextType) :
	FSocket(SOCKTYPE_Datagram, InSocketDescription)
#endif
{
	this->NetworkNextType = InNetworkNextType;
}

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_SHUTDOWN)
bool FSocketNetworkNext::Shutdown(ESocketShutdownMode Mode)
{
	/** Not supported */
	return false;
}
#endif

bool FSocketNetworkNext::Connect(const FInternetAddr& Addr)
{
	/** Not supported - connectionless (UDP) only */
	return false;
}

bool FSocketNetworkNext::Listen(int32 MaxBacklog)
{
	/** Not supported - connectionless (UDP) only */
	return false;
}

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_WAITFORPENDINGCONNECTION)
bool FSocketNetworkNext::WaitForPendingConnection(bool& bHasPendingConnection, const FTimespan& WaitTime)
{
	/** Not supported - connectionless (UDP) only */
	return false;
}
#endif

FSocket* FSocketNetworkNext::Accept(const FString& InSocketDescription)
{
	/** Not supported - connectionless (UDP) only */
	return nullptr;
}

FSocket* FSocketNetworkNext::Accept(FInternetAddr& OutAddr, const FString& InSocketDescription)
{
	/** Not supported - connectionless (UDP) only */
	return nullptr;
}

bool FSocketNetworkNext::Send(const uint8* Data, int32 Count, int32& BytesSent)
{
	/** Not supported - connectionless (UDP) only */
	BytesSent = 0;
	return false;
}

bool FSocketNetworkNext::Recv(uint8* Data, int32 BufferSize, int32& BytesRead, ESocketReceiveFlags::Type Flags)
{
	/** Not supported - connectionless (UDP) only */
	BytesRead = 0;
	return false;
}

bool FSocketNetworkNext::Wait(ESocketWaitConditions::Type Condition, FTimespan WaitTime)
{
	// not supported
	return false;
}

/**
 * Determines the connection state of the socket
 */
ESocketConnectionState FSocketNetworkNext::GetConnectionState()
{
	/** Not supported - connectionless (UDP) only */
	return SCS_NotConnected;
}

bool FSocketNetworkNext::GetPeerAddress(FInternetAddr& OutAddr)
{
	// don't support this	
	return false;
}

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_HASPENDINGCONNECTION)
bool FSocketNetworkNext::HasPendingConnection(bool& bHasPendingConnection)
{
	bHasPendingConnection = false;
	return false;
}
#endif

/**
 * Sets this socket into non-blocking mode
 *
 * @param bIsNonBlocking whether to enable blocking or not
 *
 * @return true if successful, false otherwise
 */
bool FSocketNetworkNext::SetNonBlocking(bool bIsNonBlocking) 
{
	/** Ignored, not supported */
	return true;
}

/**
 * Sets a socket into broadcast mode (UDP only)
 *
 * @param bAllowBroadcast whether to enable broadcast or not
 *
 * @return true if successful, false otherwise
 */
bool FSocketNetworkNext::SetBroadcast(bool bAllowBroadcast) 
{
	/** Ignored, not supported */
	return true;
}

bool FSocketNetworkNext::JoinMulticastGroup(const FInternetAddr& GroupAddress)
{
	return false;
}


bool FSocketNetworkNext::LeaveMulticastGroup(const FInternetAddr& GroupAddress)
{
	return false;
}

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_JOINMULTICASTGROUP_INTERFACEADDR_OVERLOAD)
bool FSocketNetworkNext::JoinMulticastGroup(const FInternetAddr& GroupAddress, const FInternetAddr& InterfaceAddress)
{
	return false;
}
#endif

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_LEAVEMULTICASTGROUP_INTERFACEADDR_OVERLOAD)
bool FSocketNetworkNext::LeaveMulticastGroup(const FInternetAddr& GroupAddress, const FInternetAddr& InterfaceAddress)
{
	return false;
}
#endif

bool FSocketNetworkNext::SetMulticastLoopback(bool bLoopback)
{
	return false;
}


bool FSocketNetworkNext::SetMulticastTtl(uint8 TimeToLive)
{
	return false;
}

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_SETMULTICASTINTERFACE)
bool FSocketNetworkNext::SetMulticastInterface(const FInternetAddr& InterfaceAddress)
{
	return false;
}
#endif

/**
 * Sets whether a socket can be bound to an address in use
 *
 * @param bAllowReuse whether to allow reuse or not
 *
 * @return true if the call succeeded, false otherwise
 */
bool FSocketNetworkNext::SetReuseAddr(bool bAllowReuse) 
{
	/** Ignored, not supported */
	return true;
}

/**
 * Sets whether and how long a socket will linger after closing
 *
 * @param bShouldLinger whether to have the socket remain open for a time period after closing or not
 * @param Timeout the amount of time to linger before closing
 *
 * @return true if the call succeeded, false otherwise
 */
bool FSocketNetworkNext::SetLinger(bool bShouldLinger, int32 Timeout) 
{
	/** Ignored, not supported */
	return true;
}

/**
 * Enables error queue support for the socket
 *
 * @param bUseErrorQueue whether to enable error queueing or not
 *
 * @return true if the call succeeded, false otherwise
 */
bool FSocketNetworkNext::SetRecvErr(bool bUseErrorQueue) 
{
	/** Ignored, not supported */
	return true;
}

/**
 * Sets the size of the send buffer to use
 *
 * @param Size the size to change it to
 * @param NewSize the out value returning the size that was set (in case OS can't set that)
 *
 * @return true if the call succeeded, false otherwise
 */
bool FSocketNetworkNext::SetSendBufferSize(int32 Size,int32& NewSize) 
{
	/** Ignored, not supported */
	return true;
}

/**
 * Sets the size of the receive buffer to use
 *
 * @param Size the size to change it to
 * @param NewSize the out value returning the size that was set (in case OS can't set that)
 *
 * @return true if the call succeeded, false otherwise
 */
bool FSocketNetworkNext::SetReceiveBufferSize(int32 Size,int32& NewSize) 
{
	/** Ignored, not supported */
	return true;
}