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

#include "NetworkNextUtils.h"
#include "NetworkNextNetDriver.h"
#include "NetworkNextConnectionType.h"
#include "Runtime/Engine/Classes/GameFramework/PlayerController.h"
#include "Runtime/Engine/Classes/Engine/NetConnection.h"
#if defined(NETWORKNEXT_INCLUDE_WORLD_H_WITH_SHORT_PATH)
#include "Engine/World.h"
#else
#include "Engine/Classes/Engine/World.h"
#endif
#if defined(NETWORKNEXT_INCLUDE_IP_ADDRESS_HEADER_IN_UTILS)
#include "IPAddress.h"
#endif

void UNetworkNextUtils::SetConfig(const FNetworkNextConfig& Config)
{
	FNetworkNextModule& NetworkNextModule = FModuleManager::LoadModuleChecked<FNetworkNextModule>("NetworkNext");
	return NetworkNextModule.SetConfig(Config);
}

void UNetworkNextUtils::SetServerConfig(const FNetworkNextServerConfig& ServerConfig)
{
	FNetworkNextModule& NetworkNextModule = FModuleManager::LoadModuleChecked<FNetworkNextModule>("NetworkNext");
	return NetworkNextModule.SetServerConfig(ServerConfig);
}

FString UNetworkNextUtils::GetClientSessionId(UObject* WorldContextObject)
{
	return FString::Printf(TEXT("%016llx"), UNetworkNextUtils::GetClientSessionIdUint64(WorldContextObject));
}

uint64 UNetworkNextUtils::GetClientSessionIdUint64(UObject* WorldContextObject)
{
	if (WorldContextObject == nullptr)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Can not retrieve client session ID: No world context object."));
		return 0;
	}

	UWorld* World = WorldContextObject->GetWorld();

	if (World == nullptr)
	{
		return 0;
	}

	UNetworkNextNetDriver* NetDriver = Cast<UNetworkNextNetDriver>(World->GetNetDriver());

	if (NetDriver == nullptr)
	{
		return 0;
	}

	FSocketNetworkNextClient* ClientSocket = NetDriver->ClientSocket;

	if (ClientSocket == nullptr)
	{
		return 0;
	}

	return ClientSocket->GetSessionId();
}

FNetworkNextClientStats UNetworkNextUtils::GetClientStats(UObject* WorldContextObject)
{
	if (WorldContextObject == nullptr)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Can not retrieve client stats: No world context object."));
		return FNetworkNextClientStats::GetDisconnectedStats();
	}

	UWorld* World = WorldContextObject->GetWorld();

	if (World == nullptr)
	{
		return FNetworkNextClientStats::GetDisconnectedStats();
	}

	UNetworkNextNetDriver* NetDriver = Cast<UNetworkNextNetDriver>(World->GetNetDriver());

	if (NetDriver == nullptr)
	{
		return FNetworkNextClientStats::GetDisconnectedStats();
	}

	FSocketNetworkNextClient* ClientSocket = NetDriver->ClientSocket;

	if (ClientSocket == nullptr)
	{
		return FNetworkNextClientStats::GetDisconnectedStats();
	}

	return ClientSocket->GetClientStats();
}

void UNetworkNextUtils::UpgradePlayer(UObject* WorldContextObject, APlayerController* PlayerController, const FString& UserId, ENetworkNextPlatformType Platform, const FString& Tag)
{
	if (PlayerController == nullptr)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Can not upgrade player controller: No player controller."));
		return;
	}

	if (WorldContextObject == nullptr)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Can not upgrade player controller: No world context object."));
		return;
	}

	UWorld* World = WorldContextObject->GetWorld();

	if (World == nullptr)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Can not upgrade player controller: No current world."));
		return;
	}

	UNetworkNextNetDriver* NetDriver = Cast<UNetworkNextNetDriver>(World->GetNetDriver());

	if (NetDriver == nullptr)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Can not upgrade player controller: No networking driver."));
		return;
	}

	FSocketNetworkNextServer* ServerSocket = NetDriver->ServerSocket;

	if (ServerSocket == nullptr)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Can not upgrade player controller: No server socket."));
		return;
	}

	UNetConnection* Connection = PlayerController->GetNetConnection();

	if (Connection == nullptr)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Can not upgrade player controller: Player controller does not have a NetConnection."));
		return;
	}

#if defined(NETWORKNEXT_SUPPORTS_GETINTERNETADDR_ON_CONNECTION)
	TSharedPtr<FInternetAddr> RemoteAddr = Connection->GetInternetAddr();
#else
	TSharedRef<FInternetAddr> RemoteAddr = ISocketSubsystem::Get(NETWORKNEXT_SUBSYSTEM)->CreateInternetAddr(
		Connection->GetAddrAsInt(),
		Connection->GetAddrPort()
	);
#endif

	ServerSocket->UpgradeClient(RemoteAddr, UserId, Platform, Tag);
}