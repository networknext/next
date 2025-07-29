/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#pragma once

#include "CoreMinimal.h"
#include "SocketSubsystem.h"
#include "Containers/Ticker.h"

FName CreateNetworkNextSocketSubsystem();

void DestroyNetworkNextSocketSubsystem();

class FNetworkNextSocketSubsystem : public ISocketSubsystem, public FTSTickerObjectBase
{
private:

    TArray<class FNetworkNextSocket*> NetworkNextSockets;

public:

    static FNetworkNextSocketSubsystem* Create();

    static void Destroy();

    FNetworkNextSocketSubsystem() {}

    virtual bool Init(FString& Error) override;

    virtual void Shutdown() override;

    virtual FSocket* CreateSocket(const FName& SocketType, const FString& SocketDescription, const FName& ProtocolName) override;

    virtual void DestroySocket(FSocket* Socket) override;

    virtual FAddressInfoResult GetAddressInfo(const TCHAR* HostName, const TCHAR* ServiceName = NULL,
        EAddressInfoFlags QueryFlags = EAddressInfoFlags::Default,
        const FName ProtocolTypeName = NAME_None,
        ESocketType SocketType = ESocketType::SOCKTYPE_Unknown) override;

    virtual TSharedPtr<FInternetAddr> GetAddressFromString(const FString& InAddress) override;

    virtual bool RequiresChatDataBeSeparate() override {
        return false;
    }

    virtual bool RequiresEncryptedPackets() override {
        return true;
    }

    virtual bool GetHostName(FString& HostName) override;

    virtual TSharedRef<FInternetAddr> CreateInternetAddr() override;

    virtual bool HasNetworkDevice() override;

    virtual const TCHAR* GetSocketAPIName() const override;

    virtual ESocketErrors GetLastErrorCode() override;

    virtual ESocketErrors TranslateErrorCode(int32 Code) override;

    virtual bool IsSocketWaitSupported() const override { 
        return false; 
    }

    virtual bool Tick(float DeltaTime) override;
};
