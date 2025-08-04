/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#pragma once

#include "CoreMinimal.h"
#include "UObject/ObjectMacros.h"
#include "IpNetDriver.h"
#include "NetworkNextNetDriver.generated.h"
                                                    
class FNetworkNotify;

UCLASS(transient, config = Engine)
class UNetworkNextNetDriver : public UIpNetDriver
{
public:

    GENERATED_BODY()

    virtual class ISocketSubsystem* GetSocketSubsystem() override;
    virtual bool IsAvailable() const override;
    virtual bool InitBase(bool bInitAsClient, FNetworkNotify* InNotify, const FURL& URL, bool bReuseAddressAndPort, FString& Error) override;
    virtual bool InitConnect(FNetworkNotify* InNotify, const FURL& ConnectURL, FString& Error) override;
    virtual bool InitListen(FNetworkNotify* InNotify, FURL& ListenURL, bool bReuseAddressAndPort, FString& Error) override;
    virtual void Shutdown() override;
    virtual bool IsNetResourceValid() override;

    bool IsClient() const;
    bool IsServer() const;

    class FNetworkNextSocketClient* GetClientSocket();
    class FNetworkNextSocketServer* GetServerSocket();

private:

    bool bIsClient;
    bool bIsServer;

    friend class UNetworkNextBlueprint;
};
