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
#include "SocketSubsystem.h"
#include "IpNetDriver.h"
#include "NetworkNextSocketSubsystem.h"
#include "NetworkNextSocketServer.h"
#include "NetworkNextSocketClient.h"

bool UNetworkNextNetDriver::IsAvailable() const
{
	ISocketSubsystem* NetworkNextSockets = ISocketSubsystem::Get(NETWORKNEXT_SUBSYSTEM);
	if (NetworkNextSockets)
	{
		UE_LOG(LogNetworkNext, Display, TEXT("Network Next net driver is available"));
		return true;
	}

	UE_LOG(LogNetworkNext, Display, TEXT("Network Next net driver is NOT available"));

	return false;
}

ISocketSubsystem* UNetworkNextNetDriver::GetSocketSubsystem()
{
	return ISocketSubsystem::Get(NETWORKNEXT_SUBSYSTEM);
}

bool UNetworkNextNetDriver::InitBase(bool bInitAsClient, FNetworkNotify* InNotify, const FURL& URL, bool bReuseAddressAndPort, FString& Error)
{
	UE_LOG(LogNetworkNext, Display, TEXT("UNetworkNextNetDriver::InitBase"));

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

	if (GetSocket() == NULL)
	{
		UE_LOG(LogNetworkNext, Warning, TEXT("GetSocket() is NULL"));
		Error = FString::Printf(TEXT("socket failed (%i)"), (int32)SocketSubsystem->GetLastErrorCode());
		return false;
	}

	LocalAddr = SocketSubsystem->GetLocalBindAddr(*GLog);

	if (bInitAsClient)
	{
		// bind client to an ephemeral port
		LocalAddr->SetPort(0);
	}
	else
	{
		// bind server to the specified port
		LocalAddr->SetPort(URL.Port);
	}

	int32 BoundPort = SocketSubsystem->BindNextPort(GetSocket(), *LocalAddr, MaxPortCountToTry + 1, 1);

	UE_LOG(LogNet, Display, TEXT("%s bound to port %d"), *GetName(), BoundPort);

	return true;
}

bool UNetworkNextNetDriver::InitConnect(FNetworkNotify* InNotify, const FURL& ConnectURL, FString& Error)
{
	UE_LOG(LogNetworkNext, Display, TEXT("UNetworkNextNetDriver::InitConnect"));

    FNetworkNextSocketSubsystem* NetworkNextSockets = (FNetworkNextSocketSubsystem*)ISocketSubsystem::Get(NETWORKNEXT_SUBSYSTEM);
    if (!NetworkNextSockets)
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Could not find network next socket subsystem"));
        return false;
    }
    
    FSocket* NewSocket = NetworkNextSockets->CreateSocket(FName(TEXT("NetworkNextSocketClient")), TEXT("Unreal client (Network Next)"), FName(TEXT("NetworkNext")));
    if (!NewSocket)
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Could not create network next client socket"));
        return false;
    }

    SetSocketAndLocalAddress(NewSocket);

    ClientSocket = (FNetworkNextSocketClient*)NewSocket;
    ServerSocket = NULL;

	return Super::InitConnect(InNotify, ConnectURL, Error);
}

bool UNetworkNextNetDriver::InitListen(FNetworkNotify* InNotify, FURL& ListenURL, bool bReuseAddressAndPort, FString& Error)
{
	UE_LOG(LogNetworkNext, Display, TEXT("UNetworkNextNetDriver::InitListen"));

	FNetworkNextSocketSubsystem* NetworkNextSockets = (FNetworkNextSocketSubsystem*)ISocketSubsystem::Get(NETWORKNEXT_SUBSYSTEM);
    if (!NetworkNextSockets)
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Could not find network next socket subsystem"));
        return false;
    }

    FSocket* NewSocket = NetworkNextSockets->CreateSocket(FName(TEXT("NetworkNextSocketServer")), TEXT("Unreal server (Network Next)"), FName(TEXT("NetworkNext")));
    if (!NewSocket)
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Could not create network next server socket"));
        return false;
    }

    SetSocketAndLocalAddress(NewSocket);

    ServerSocket = (FNetworkNextSocketServer*)NewSocket;
    ClientSocket = NULL;

    return Super::InitListen(InNotify, ListenURL, bReuseAddressAndPort, Error);
}

void UNetworkNextNetDriver::Shutdown()
{
	UE_LOG(LogNetworkNext, Display, TEXT("FNetworkNextNetDriver::Shutdown"));

	ClientSocket = NULL;
	ServerSocket = NULL;

	Super::Shutdown();
}

bool UNetworkNextNetDriver::IsNetResourceValid()
{
	return true;
}
