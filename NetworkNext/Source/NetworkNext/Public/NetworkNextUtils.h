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
#include "Kismet/BlueprintFunctionLibrary.h"
#include "NetworkNextClientStats.h"
#include "NetworkNextConfig.h"
#include "NetworkNextServerConfig.h"
#include "NetworkNextUtils.generated.h"

UCLASS()
class NETWORKNEXT_API UNetworkNextUtils : public UBlueprintFunctionLibrary
{
	GENERATED_BODY()

public:

	/**
	 * Sets the configuration for Network Next. This function only has an effect if it's called before the first network socket is opened.
	 * That is, you should call this function during the OnPlay of your game instance, before any networked world is opened.
	 */
	UFUNCTION(BlueprintCallable, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Set Network Next Configuration"))
	static void SetConfig(const FNetworkNextConfig& Config);

	/**
	 * Sets the server configuration for Network Next. This will be used when server socket is bound.
	 */
	UFUNCTION(BlueprintCallable, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Set Network Next Server Configuration"))
	static void SetServerConfig(const FNetworkNextServerConfig& ServerConfig);

	/**
	 * Returns the current client session ID.
	 */
	UFUNCTION(BlueprintPure, BlueprintCosmetic, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Get Client Session ID"))
	static FString GetClientSessionId(UObject* WorldContextObject);
	
	/**
	 * Returns the current client statistics.
	 */
	UFUNCTION(BlueprintCallable, BlueprintCosmetic, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Get Client Session Statistics"))
	static FNetworkNextClientStats GetClientStats(UObject* WorldContextObject);
	
	/**
	 * Upgrades a player session to Network Next.
	 */
	UFUNCTION(BlueprintCallable, BlueprintAuthorityOnly, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Upgrade Player Session"))
	static void UpgradePlayer(UObject* WorldContextObject, APlayerController* PlayerController, const FString& UserId);
	
	/**
	 * Returns the current client session ID.
	 */
	static uint64 GetClientSessionIdUint64(UObject* WorldContextObject);
};
