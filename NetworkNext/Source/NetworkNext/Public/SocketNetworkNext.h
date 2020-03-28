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

#pragma once

#include "NetworkNextBuildConfig.h"
#include "CoreMinimal.h"
#include "OnlineSubsystemNames.h"
#include "Sockets.h"
#include "NetworkNext.h"

enum class ESocketNetworkNextType : uint8
{
	TYPE_None,
	TYPE_Client,
	TYPE_Server
};

class FSocketNetworkNext :
	public FSocket
{
private:

	ESocketNetworkNextType NetworkNextType;

public:
#if defined(NETWORKNEXT_HAS_ESOCKETPROTOCOLFAMILY)
	FSocketNetworkNext(const FString& InSocketDescription, ESocketProtocolFamily InSocketProtocol, ESocketNetworkNextType InNetworkNextType);
#else
	FSocketNetworkNext(const FString& InSocketDescription, ESocketNetworkNextType InNetworkNextType);
#endif
	
#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_SHUTDOWN)
	virtual bool Shutdown(ESocketShutdownMode Mode) override;
#endif

	virtual void UpdateNetworkNextSocket() = 0;

	// These are all the TCP methods that we don't support.

	virtual bool Connect(const FInternetAddr& Addr) override;

	virtual bool Listen(int32 MaxBacklog) override;

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_WAITFORPENDINGCONNECTION)
	virtual bool WaitForPendingConnection(bool& bHasPendingConnection, const FTimespan& WaitTime) override;
#endif

	virtual class FSocket* Accept(const FString& SocketDescription) override;

	virtual class FSocket* Accept(FInternetAddr& OutAddr, const FString& SocketDescription) override;

	virtual bool Send(const uint8* Data, int32 Count, int32& BytesSent) override;

	virtual bool Recv(uint8* Data, int32 BufferSize, int32& BytesRead, ESocketReceiveFlags::Type Flags = ESocketReceiveFlags::None) override;

	virtual bool Wait(ESocketWaitConditions::Type Condition, FTimespan WaitTime) override;

	virtual ESocketConnectionState GetConnectionState() override;

	virtual bool GetPeerAddress(FInternetAddr& OutAddr) override;

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_HASPENDINGCONNECTION)
	virtual bool HasPendingConnection(bool& bHasPendingConnection) override;
#endif

	/**
	 * Sets this socket into non-blocking mode
	 *
	 * @param bIsNonBlocking whether to enable blocking or not
	 *
	 * @return true if successful, false otherwise
	 */
	virtual bool SetNonBlocking(bool bIsNonBlocking = true) override;

	/**
	 * Sets a socket into broadcast mode (UDP only)
	 *
	 * @param bAllowBroadcast whether to enable broadcast or not
	 *
	 * @return true if successful, false otherwise
	 */
	virtual bool SetBroadcast(bool bAllowBroadcast = true) override;

	virtual bool JoinMulticastGroup(const FInternetAddr& GroupAddress) override;

	virtual bool LeaveMulticastGroup(const FInternetAddr& GroupAddress) override;

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_JOINMULTICASTGROUP_INTERFACEADDR_OVERLOAD)
	virtual bool JoinMulticastGroup(const FInternetAddr& GroupAddress, const FInternetAddr& InterfaceAddress) override;
#endif

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_LEAVEMULTICASTGROUP_INTERFACEADDR_OVERLOAD)
	virtual bool LeaveMulticastGroup(const FInternetAddr& GroupAddress, const FInternetAddr& InterfaceAddress) override;
#endif

	virtual bool SetMulticastLoopback(bool bLoopback) override;

	virtual bool SetMulticastTtl(uint8 TimeToLive) override;

#if defined(NETWORKNEXT_SOCKET_INTERFACE_HAS_SETMULTICASTINTERFACE)
	virtual bool SetMulticastInterface(const FInternetAddr& InterfaceAddress) override;
#endif

	/**
	 * Sets whether a socket can be bound to an address in use
	 *
	 * @param bAllowReuse whether to allow reuse or not
	 *
	 * @return true if the call succeeded, false otherwise
	 */
	virtual bool SetReuseAddr(bool bAllowReuse = true) override;

	/**
	 * Sets whether and how long a socket will linger after closing
	 *
	 * @param bShouldLinger whether to have the socket remain open for a time period after closing or not
	 * @param Timeout the amount of time to linger before closing
	 *
	 * @return true if the call succeeded, false otherwise
	 */
	virtual bool SetLinger(bool bShouldLinger = true, int32 Timeout = 0) override;

	/**
	 * Enables error queue support for the socket
	 *
	 * @param bUseErrorQueue whether to enable error queueing or not
	 *
	 * @return true if the call succeeded, false otherwise
	 */
	virtual bool SetRecvErr(bool bUseErrorQueue = true) override;

	/**
	 * Sets the size of the send buffer to use
	 *
	 * @param Size the size to change it to
	 * @param NewSize the out value returning the size that was set (in case OS can't set that)
	 *
	 * @return true if the call succeeded, false otherwise
	 */
	virtual bool SetSendBufferSize(int32 Size,int32& NewSize) override;

	/**
	 * Sets the size of the receive buffer to use
	 *
	 * @param Size the size to change it to
	 * @param NewSize the out value returning the size that was set (in case OS can't set that)
	 *
	 * @return true if the call succeeded, false otherwise
	 */
	virtual bool SetReceiveBufferSize(int32 Size,int32& NewSize) override;
};
