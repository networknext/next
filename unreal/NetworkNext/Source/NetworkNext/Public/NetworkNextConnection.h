/*
	Network Next. Copyright 2017 - 2025 Network Next, Inc.  
	Licensed under the Network Next Source Available License 1.0
*/

#pragma once

#include "CoreMinimal.h"
#include "IpConnection.h"
#include "NetworkNextConnection.generated.h"

UCLASS(transient, config = Engine)
class NETWORKNEXT_API UNetworkNextConnection : public UIpConnection
{						   
	GENERATED_BODY()
	
	virtual void InitRemoteConnection(class UNetDriver* InDriver, class FSocket* InSocket, const FURL& InURL, const class FInternetAddr& InRemoteAddr, EConnectionState InState, int32 InMaxPacket = 0, int32 InPacketOverhead = 0) override;

	virtual void InitLocalConnection(class UNetDriver* InDriver, class FSocket* InSocket, const FURL& InURL, EConnectionState InState, int32 InMaxPacket = 0, int32 InPacketOverhead = 0) override;
};
