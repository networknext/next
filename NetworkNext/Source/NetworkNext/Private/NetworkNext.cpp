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

#include "NetworkNext.h"
#include "NetworkNextNetDriver.h"
#include "NetworkNextSocketSubsystem.h"
#include "Core.h"
#include "next.h"

DEFINE_LOG_CATEGORY(LogNetworkNext);

IMPLEMENT_MODULE(FNetworkNextModule, NetworkNext)

void FNetworkNextModule::StartupModule()
{
	UE_LOG(LogNetworkNext, Display, TEXT("Network Next Plugin loaded"));

	m_initialized_sdk = false;

	// Perform Network Next allocations through the UE4 allocator instead of default malloc/free
	next_allocator(&FNetworkNextModule::Malloc, &FNetworkNextModule::Free);

	// Setup logging to go to the "NetworkNext" log category
	next_log_function(&FNetworkNextModule::Log);

	// Set the log level. If you need to debug something, increase the log level to NEXT_LOG_LEVEL_DEBUG and send us a log and we're happy to help you out
	next_log_level(NEXT_LOG_LEVEL_INFO);

	// Setup default config values for the Network Next SDK
	next_config_t config;
	next_default_config(&config);

	// Get at the Network Next config from the NetworkNext.NetworkNextNetDriver object
	UNetworkNextNetDriver* NetDriver = NewObject<UNetworkNextNetDriver>();
	if (!NetDriver)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Network Next could not get config"));
		return;
	}

	// Copy across the Network Next hostname.
	// This is the custom backend setup for your studio. It is typically set to [yourcompany].spacecats.net
	if (NetDriver->NextHostname.Len() > 0)
	{
		FString Hostname = NetDriver->NextHostname;
		int Len = FMath::Min(Hostname.Len(), (int)sizeof(config.hostname));
		FMemory::Memzero(&config.hostname, sizeof(config.hostname));
		for (int i = 0; i < Len; i++)
		{
			config.hostname[i] = Hostname[i];
		}
	}

	// Copy across the customer public key from the net driver. This is only required for the client.
	if (NetDriver->CustomerPublicKey.Len() > 0)
	{
		FString PublicKey = NetDriver->CustomerPublicKey;
		int Len = FMath::Min(PublicKey.Len(), (int)sizeof(config.customer_public_key));
		FMemory::Memzero(&config.customer_public_key, sizeof(config.customer_public_key));
		for (int i = 0; i < Len; i++)
		{
			config.customer_public_key[i] = PublicKey[i];
		}
	}

	// Copy across the customer private key from the net driver. This is only required for the server.
	// IMPORTANT: Do not set the private key on the client. You must not let your players know your customer private key!
	if (NetDriver->CustomerPrivateKey.Len() > 0)
	{
		FString PrivateKey = NetDriver->CustomerPrivateKey;
		int Len = FMath::Min(PrivateKey.Len(), (int)sizeof(config.customer_private_key));
		FMemory::Memzero(&config.customer_private_key, sizeof(config.customer_private_key));
		for (int i = 0; i < Len; i++)
		{
			config.customer_private_key[i] = PrivateKey[i];
		}
	}

	// Copy across the disable flag.
	// If you set this on the client, that client will not be monitored or accelerated, but can still connect to servers with Network Next enabled.
	// If you set this on the server, the server will not monitor or accelerate any clients, even if those clients have Network Next enabled.
	// TLDR: Enabled and disabled clients and servers are compatible!
	config.disable_network_next = NetDriver->DisableNetworkNext;

	// Print out the network next hostname for ease of debugging
	FString hostname = FString(config.hostname);
	UE_LOG(LogNetworkNext, Display, TEXT("Hostname is %s"), *hostname);

	// Initialize the Network Next SDK
	if (next_init(NULL, &config) != NEXT_OK)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Network Next SDK failed to initalize!"));
		return;
	}

	UE_LOG(LogNetworkNext, Display, TEXT("Network Next SDK initialized"));

	m_initialized_sdk = true;

	CreateNetworkNextSocketSubsystem();
}

void FNetworkNextModule::ShutdownModule()
{
	UE_LOG(LogNetworkNext, Display, TEXT("Network Next SDK shutting down"));

	next_term();

	m_initialized_sdk = false;

	DestroyNetworkNextSocketSubsystem();

	UE_LOG(LogNetworkNext, Display, TEXT("Network Next Plugin unloaded"));
}

void FNetworkNextModule::Log(int level, const char* format, ...)
{
	va_list args;
	va_start(args, format);
	char buffer[1024];
	vsnprintf(buffer, sizeof(buffer), format, args);
	va_end(args);

	FString Message = FString(buffer);

	switch (level)
	{
	case NEXT_LOG_LEVEL_ERROR:
		UE_LOG(LogNetworkNext, Error, TEXT("%s"), *Message);
		break;
	case NEXT_LOG_LEVEL_WARN:
		UE_LOG(LogNetworkNext, Warning, TEXT("%s"), *Message);
		break;
	case NEXT_LOG_LEVEL_INFO:
		UE_LOG(LogNetworkNext, Display, TEXT("%s"), *Message);
		break;
	case NEXT_LOG_LEVEL_DEBUG:
	default:
		UE_LOG(LogNetworkNext, VeryVerbose, TEXT("%s"), *Message);
		break;
	}
}

void* FNetworkNextModule::Malloc(void* context, size_t size)
{
	return FMemory::Malloc(size);
}

void FNetworkNextModule::Free(void* context, void* src)
{
	return FMemory::Free(src);
}
