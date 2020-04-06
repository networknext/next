/*
    Network Next SDK 3.4.0

    Copyright © 2017 - 2020 Network Next, Inc.

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
#include "NetworkNextPackage.h"
#include "NetworkNextConnectionType.h"
#if defined(NETWORKNEXT_INCLUDE_NEXT_H_WITH_SHORT_PATH)
#include "next.h"
#else
#include "NetworkNextLibrary/next/include/next.h"
#endif
#include "NetworkNextConfig.generated.h"

USTRUCT(BlueprintType)
struct FNetworkNextConfig
{
	GENERATED_BODY()

	/** The public key as a base-64 encoded string. If this is left empty, it is read from the CustomerPublicKeyBase64 configuration value. If that is left empty, the public key is read from the NEXT_CUSTOMER_PUBLIC_KEY environment variable. */
	UPROPERTY(EditAnywhere, BlueprintReadWrite, Meta = (DisplayName = "Public Key (Base64)"))
	FString PublicKeyBase64;

	/** The private key as a base-64 encoded string. If this is left empty, it is read from the CustomerPrivateKeyBase64 configuration value. If that is left empty, the private key is read from the NEXT_CUSTOMER_PRIVATE_KEY environment variable. This value must never be shipped to players / game clients. */
	UPROPERTY(EditAnywhere, BlueprintReadWrite, Meta = (DisplayName = "Private Key (Base64)"))
	FString PrivateKeyBase64;
	
	/** The socket send buffer size. If you don't set this, the default is used. */
	UPROPERTY(EditAnywhere, BlueprintReadWrite)
	int32 SocketSendBufferSize;
	
	/** The socket receive buffer size. If you don't set this, the default is used. */
	UPROPERTY(EditAnywhere, BlueprintReadWrite)
	int32 SocketReceiveBufferSize;

	/** The socket receive buffer size. If you don't set this, the default is used. */
	UPROPERTY(EditAnywhere, BlueprintReadWrite)
	int32 SocketReceiveBufferSize;

	/** Disable network next */
	UPROPERTY(EditAnywhere, BlueprintReadWrite)
	bool DisableNetworkNext;

	/** Disable tagging */
	UPROPERTY(EditAnywhere, BlueprintReadWrite)
	bool DisableTagging;

	FNetworkNextConfig()
	{
		next_config_t default_config;
		next_default_config(&default_config);
		this->PublicKeyBase64 = "";
		this->PrivateKeyBase64 = "";
		this->SocketSendBufferSize = default_config.socket_send_buffer_size;
		this->SocketReceiveBufferSize = default_config.socket_receive_buffer_size;
		this->DisableNetworkNext = default_config.disable_network_next;
		this->DisableTagging = default_config.disable_tagging;
	}
};