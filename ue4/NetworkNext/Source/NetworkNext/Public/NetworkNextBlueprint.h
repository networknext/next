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

#pragma once

#include "CoreMinimal.h"
#include "Kismet/BlueprintFunctionLibrary.h"
#include "NetworkNextBlueprint.generated.h"

UCLASS()
class NETWORKNEXT_API UNetworkNextBlueprint : public UBlueprintFunctionLibrary
{
    GENERATED_BODY()

public:

    // server functions

    /**
     * Upgrades a player session on Network Next
     */
    UFUNCTION(BlueprintCallable, BlueprintAuthorityOnly, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Upgrade Player Session"))
    static void UpgradePlayer(UObject* WorldContextObject, APlayerController* PlayerController, const FString& UserId);

    /**
     * Trigger a server event (or server events) for a player
     */
    static void ServerEvent(UObject* WorldContextObject, APlayerController* PlayerController, uint64 ServerEvents);

    // client functions

    /**
     * Report server
     */
    UFUNCTION(BlueprintCallable, BlueprintAuthorityOnly, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Report Server"))
    static void ReportServer(UObject* WorldContextObject);

    /**
     * Get latency
     */
    UFUNCTION(BlueprintCallable, BlueprintAuthorityOnly, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Get Latency"))
    static float GetLatency(UObject* WorldContextObject);

    /**
     * Get jitter
     */
    UFUNCTION(BlueprintCallable, BlueprintAuthorityOnly, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Get Jitter"))
    static float GetJitter(UObject* WorldContextObject);

    /**
     * Get packet loss
     */
    UFUNCTION(BlueprintCallable, BlueprintAuthorityOnly, Category = "Network Next", meta = (WorldContext = "WorldContextObject", DisplayName = "Get Packet Loss"))
    static float GetPacketLoss(UObject* WorldContextObject);
};
