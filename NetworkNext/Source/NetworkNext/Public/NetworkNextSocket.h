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

#include "CoreMinimal.h"
#include "OnlineSubsystemNames.h"
#include "Sockets.h"
#include "NetworkNext.h"

/*
	There are three types of Next sockets in the Network Next UE4 plugin:

		1. Network Next Client
		2. Network Next Server
		3. Passthrough

	Each socket implements *only* UDP datagram support. All TCP style socket support is stubbed out.

	The client socket is used by client when it connects to a server.

	The server socket is used by a listen server.

	The passthrough socket is used for platforms that don't have the Network Next SDK.

	This makes it possible to use the Network Next UE4 plugin, even if some of your platforms aren't supported (yet).

	All other socket use is directed towards the regular platform socket interface and socket types. We do not get in the way of those!
*/

enum class ENetworkNextSocketType : uint8
{
	TYPE_None,
	TYPE_Client,
	TYPE_Server,
    TYPE_Passthrough
};

class FNetworkNextSocket : public FSocket
{
private:

	ENetworkNextSocketType NetworkNextType;

public:

	FNetworkNextSocket(ENetworkNextSocketType InNetworkNextType, const FString& InSocketDescription, FName InProtocol);

    virtual void UpdateNetworkNextSocket() = 0;

	// IMPORTANT: All methods below are stubbed out. Please don't use them :)

	virtual bool Shutdown(ESocketShutdownMode Mode) override;

	virtual bool Connect(const FInternetAddr& Addr) override;

	virtual bool Listen(int32 MaxBacklog) override;

	virtual bool WaitForPendingConnection(bool& bHasPendingConnection, const FTimespan& WaitTime) override;

	virtual bool HasPendingData(uint32& PendingDataSize) override;

	virtual class FSocket* Accept(const FString& SocketDescription) override;

	virtual class FSocket* Accept(FInternetAddr& OutAddr, const FString& SocketDescription) override;

	virtual bool Send(const uint8* Data, int32 Count, int32& BytesSent) override;

	virtual bool Recv(uint8* Data, int32 BufferSize, int32& BytesRead, ESocketReceiveFlags::Type Flags = ESocketReceiveFlags::None) override;

	virtual bool Wait(ESocketWaitConditions::Type Condition, FTimespan WaitTime) override;

	virtual ESocketConnectionState GetConnectionState() override;

	virtual bool GetPeerAddress(FInternetAddr& OutAddr) override;

	virtual bool SetNonBlocking(bool bIsNonBlocking = true) override;

	virtual bool SetBroadcast(bool bAllowBroadcast = true) override;

	virtual bool JoinMulticastGroup(const FInternetAddr& GroupAddress) override;

	virtual bool LeaveMulticastGroup(const FInternetAddr& GroupAddress) override;

	virtual bool JoinMulticastGroup(const FInternetAddr& GroupAddress, const FInternetAddr& InterfaceAddress) override;

	virtual bool LeaveMulticastGroup(const FInternetAddr& GroupAddress, const FInternetAddr& InterfaceAddress) override;

	virtual bool SetMulticastLoopback(bool bLoopback) override;

	virtual bool SetMulticastTtl(uint8 TimeToLive) override;

	virtual bool SetMulticastInterface(const FInternetAddr& InterfaceAddress) override;

	virtual bool SetReuseAddr(bool bAllowReuse = true) override;

	virtual bool SetLinger(bool bShouldLinger = true, int32 Timeout = 0) override;

	virtual bool SetRecvErr(bool bUseErrorQueue = true) override;

	virtual bool SetSendBufferSize(int32 Size,int32& NewSize) override;

	virtual bool SetReceiveBufferSize(int32 Size,int32& NewSize) override;

    virtual bool SetNoDelay(bool flag) override;
};
