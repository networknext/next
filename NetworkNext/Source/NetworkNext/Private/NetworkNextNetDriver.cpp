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

#include "NetworkNextNetDriver.h"
#include "NetworkNext.h"
#include "OnlineSubsystemNames.h"
#include "OnlineSubsystem.h"
#if defined(NETWORKNEXT_INCLUDE_ENGINE_BASE_TYPES_IN_NETDRIVER)
#include "Engine/Classes/Engine/EngineBaseTypes.h"
#endif
#include "SocketSubsystem.h"
#include "SocketSubsystemNetworkNext.h"

UNetworkNextNetDriver::UNetworkNextNetDriver()
{
	this->ClientSocket = nullptr;
	this->ServerSocket = nullptr;
}

bool UNetworkNextNetDriver::IsAvailable() const
{
	ISocketSubsystem* NetworkNextSockets = ISocketSubsystem::Get(NETWORKNEXT_SUBSYSTEM);
	if (NetworkNextSockets)
	{
		return true;
	}

	return false;
}

ISocketSubsystem* UNetworkNextNetDriver::GetSocketSubsystem()
{
	return ISocketSubsystem::Get(NETWORKNEXT_SUBSYSTEM);
}

bool UNetworkNextNetDriver::InitBase(bool bInitAsClient, FNetworkNotify* InNotify, const FURL& URL, bool bReuseAddressAndPort, FString& Error)
{
	UE_LOG(LogNetworkNext, Display, TEXT("UNetworkNextDriver::InitBase"));

	if (!UNetDriver::InitBase(bInitAsClient, InNotify, URL, bReuseAddressAndPort, Error))
	{
		UE_LOG(LogNetworkNext, Warning, TEXT("UIpNetDriver::InitBase failed"));
		return false;
	}
	
	ISocketSubsystem* SocketSubsystem = GetSocketSubsystem();
	if (SocketSubsystem == NULL)
	{
		UE_LOG(LogNetworkNext, Warning, TEXT("Unable to find socket subsystem"));
		Error = TEXT("Unable to find socket subsystem");
		return false;
	}

	if (Socket == NULL)
	{
		UE_LOG(LogNetworkNext, Warning, TEXT("Socket is NULL"));
		Socket = 0;
		Error = FString::Printf(TEXT("socket failed (%i)"), (int32)SocketSubsystem->GetLastErrorCode());
		return false;
	}

	LocalAddr = SocketSubsystem->GetLocalBindAddr(*GLog);

	if (bInitAsClient)
	{
		// force client to bind to an ephemeral port
		LocalAddr->SetPort(0);
	}
	else
	{
		// bind server to the specified port
		LocalAddr->SetPort(URL.Port);
	}

	int32 BoundPort = SocketSubsystem->BindNextPort(Socket, *LocalAddr, MaxPortCountToTry + 1, 1);

	UE_LOG(LogNet, Display, TEXT("%s bound to port %d"), *GetName(), BoundPort);

	return true;
}

bool UNetworkNextNetDriver::InitConnect(FNetworkNotify* InNotify, const FURL& ConnectURL, FString& Error)
{
	UE_LOG(LogNetworkNext, Display, TEXT("UNetworkNextDriver::InitConnect"));

	FSocketSubsystemNetworkNext* NetworkNextSockets = (FSocketSubsystemNetworkNext*)ISocketSubsystem::Get(NETWORKNEXT_SUBSYSTEM);
	if (NetworkNextSockets)
	{
		Socket = NetworkNextSockets->CreateSocketWithNetDriver(FName(TEXT("NetworkNextClientSocket")), TEXT("Unreal client (Network Next)"), this
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_PROTOCOLTYPE)
			, ESocketProtocolFamily::None
#endif
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_FORCEUDP)
			, false
#endif
		);
		if (Socket->GetDescription().StartsWith(TEXT("SOCKET_TYPE_NEXT_CLIENT_")))
		{
			ClientSocket = (FSocketNetworkNextClient*)Socket;
		}
	}

	return Super::InitConnect(InNotify, ConnectURL, Error);
}

bool UNetworkNextNetDriver::InitListen(FNetworkNotify* InNotify, FURL& ListenURL, bool bReuseAddressAndPort, FString& Error)
{
	UE_LOG(LogNetworkNext, Display, TEXT("UNetworkNextDriver::InitListen"));

	FSocketSubsystemNetworkNext* NetworkNextSockets = (FSocketSubsystemNetworkNext*)ISocketSubsystem::Get(NETWORKNEXT_SUBSYSTEM);
	if (NetworkNextSockets)
	{
		UE_LOG(LogNetworkNext, Display, TEXT("Create server socket"));

		Socket = NetworkNextSockets->CreateSocketWithNetDriver(FName(TEXT("NetworkNextServerSocket")), TEXT("Unreal server (Network Next)"), this
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_PROTOCOLTYPE)
			, ESocketProtocolFamily::None
#endif
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_FORCEUDP)
			, false
#endif
		);
		if (Socket->GetDescription().StartsWith(TEXT("SOCKET_TYPE_NEXT_SERVER_")))
		{
			ServerSocket = (FSocketNetworkNextServer*)Socket;
		}
	}
	else
	{
		UE_LOG(LogNetworkNext, Warning, TEXT("Could not find NetworkNextSockets module?"));
	}

	return Super::InitListen(InNotify, ListenURL, bReuseAddressAndPort, Error);
}

void UNetworkNextNetDriver::Shutdown()
{
	ClientSocket = nullptr;
	ServerSocket = nullptr;
	Super::Shutdown();
}

bool UNetworkNextNetDriver::IsNetResourceValid()
{
	return true;
}