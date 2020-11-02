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
#include "SocketSubsystem.h"
#include "Containers/Ticker.h"

FName CreateNetworkNextSocketSubsystem();

void DestroyNetworkNextSocketSubsystem();

class FNetworkNextSocketSubsystem : public ISocketSubsystem, public FTickerObjectBase
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
