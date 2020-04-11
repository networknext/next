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
#include "Core.h"
#include "Misc/Base64.h"
#include "Kismet/KismetMathLibrary.h"
#include "Modules/ModuleManager.h"
#include "Interfaces/IPluginManager.h"
#if defined(NETWORKNEXT_INCLUDE_NEXT_H_WITH_SHORT_PATH)
#include "next.h"
#else
#include "NetworkNextLibrary/next/include/next.h"
#endif
#include "SocketSubsystemNetworkNext.h"
#include "NetworkNextNetDriver.h"
#include "UObject/UObjectGlobals.h"
#if WITH_EDITOR
#include "Editor.h"
#endif
#include <cstdarg>

DEFINE_LOG_CATEGORY(LogNetworkNext);

void FNetworkNextModule::StartupModule()
{
#if defined(NETWORKNEXT_ENABLE_DELAY_LOAD)
	FString BaseDir = IPluginManager::Get().FindPlugin("NetworkNext")->GetBaseDir();

#if defined(NETWORKNEXT_ENABLE_DELAY_LOAD_WIN64)
	FString LibraryPath = FPaths::Combine(*BaseDir, TEXT("Source/ThirdParty/NetworkNextLibrary/next/lib/Windows-x86_64/Release/next.dll"));
#elif defined(NETWORKNEXT_ENABLE_DELAY_LOAD_WIN32)
	FString LibraryPath = FPaths::Combine(*BaseDir, TEXT("Source/ThirdParty/NetworkNextLibrary/next/lib/Windows-x86/Release/next.dll"));
// <PS4
#elif defined(NETWORKNEXT_ENABLE_DELAY_LOAD_PS4)
	FString LibraryPath = "next.prx";
// PS4>
#else
#error Unsupported delay load path in NetworkNext.cpp
#endif

	this->NetworkNextHandle = FPlatformProcess::GetDllHandle(*LibraryPath);

	if (!this->NetworkNextHandle)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Delayed load of network next library failed."));
	}

#endif
  
	next_allocator(&FNetworkNextModule::Malloc, &FNetworkNextModule::Free);

	next_log_function(&FNetworkNextModule::NextLogFunction);
	next_log_level(NEXT_LOG_LEVEL_DEBUG);

	this->NetworkNextIsInitialized = false;
	this->NetworkNextIsSuccessfullyInitialized = false;
	this->NetworkNextConfig = nullptr;

	CreateNetworkNextSocketSubsystem();

#if WITH_EDITOR
	FEditorDelegates::EndPIE.AddRaw(this, &FNetworkNextModule::ShutdownNetworkNextAfterPIE);
#endif
}

void FNetworkNextModule::InitializeNetworkNextIfRequired()
{
	if (!this->NetworkNextIsInitialized)
	{
		this->NetworkNextIsInitialized = true;
		this->NetworkNextIsSuccessfullyInitialized = false;

		next_config_t config;
		next_default_config(&config);

		// temporary. force UE4 plugin to point to pubg specific hostname. make configurable later.
		strcpy_s(config.hostname, "pubg.networknext.com");

		bool publicKeySet = false;
		bool privateKeySet = false;

		// Use values from engine / game config. We just make a UNetworkNextNetDriver
		// here to get at the config.
		UNetworkNextNetDriver* NetDriver = NewObject<UNetworkNextNetDriver>();
		if (NetDriver->CustomerPublicKeyBase64.Len() > 0)
		{
			FString PublicKey = NetDriver->CustomerPublicKeyBase64;
			int Len = FMath::Min(PublicKey.Len(), 256);
			FMemory::Memzero(&config.customer_public_key, sizeof(config.customer_public_key));
			for (int i = 0; i < Len; i++)
			{
				config.customer_public_key[i] = PublicKey[i];
			}

			publicKeySet = true;
		}
		if (NetDriver->CustomerPrivateKeyBase64.Len() > 0)
		{
			FString PrivateKey = NetDriver->CustomerPrivateKeyBase64;
			int Len = FMath::Min(PrivateKey.Len(), 256);
			FMemory::Memzero(&config.customer_private_key, sizeof(config.customer_private_key));
			for (int i = 0; i < Len; i++)
			{
				config.customer_private_key[i] = PrivateKey[i];
			}

			privateKeySet = true;
		}

		// Use values from programmatic config.
		if (this->NetworkNextConfig != nullptr)
		{
			if (this->NetworkNextConfig->PublicKeyBase64.Len() > 0)
			{
				int Len = FMath::Min(this->NetworkNextConfig->PublicKeyBase64.Len(), 256);
				FMemory::Memzero(&config.customer_public_key, sizeof(config.customer_public_key));
				for (int i = 0; i < Len; i++)
				{
					config.customer_public_key[i] = this->NetworkNextConfig->PublicKeyBase64[i];
				}

				publicKeySet = true;
			}
			if (this->NetworkNextConfig->PrivateKeyBase64.Len() > 0)
			{
				int Len = FMath::Min(this->NetworkNextConfig->PrivateKeyBase64.Len(), 256);
				FMemory::Memzero(&config.customer_private_key, sizeof(config.customer_private_key));
				for (int i = 0; i < Len; i++)
				{
					config.customer_private_key[i] = this->NetworkNextConfig->PrivateKeyBase64[i];
				}

				privateKeySet = true;
			}

			config.socket_send_buffer_size = this->NetworkNextConfig->SocketSendBufferSize;
			config.socket_receive_buffer_size = this->NetworkNextConfig->SocketReceiveBufferSize;
			config.disable_network_next = this->NetworkNextConfig->DisableNetworkNext;
			config.disable_tagging = this->NetworkNextConfig->DisableTagging;
		}

		// If failed to get keys at this point - fall back to environment variables
		if (!publicKeySet)
		{
			FString publicKey = GetEnvironmentVariable(FString(TEXT("NEXT_CUSTOMER_PUBLIC_KEY")));

			int Len = FMath::Min(publicKey.Len(), 256);
			FMemory::Memzero(&config.customer_public_key, sizeof(config.customer_public_key));
			for (int i = 0; i < Len; i++)
			{
				config.customer_public_key[i] = publicKey[i];
			}

			publicKeySet = true;
		}
		
		if (!privateKeySet)
		{
			FString privateKey = GetEnvironmentVariable(FString(TEXT("NEXT_CUSTOMER_PRIVATE_KEY")));

			int Len = FMath::Min(privateKey.Len(), 256);
			FMemory::Memzero(&config.customer_private_key, sizeof(config.customer_private_key));
			for (int i = 0; i < Len; i++)
			{
				config.customer_private_key[i] = privateKey[i];
			}

			privateKeySet = true;
		}

		FString hostname = FString(config.hostname);
		UE_LOG(LogNetworkNext, Display, TEXT("Hostname is: %s"), *hostname);
			
		if (next_init(nullptr, &config) != NEXT_OK)
		{
			UE_LOG(LogNetworkNext, Error, TEXT("Network Next could not be initialized"));
			return;
		}

		this->NetworkNextIsSuccessfullyInitialized = true;

		UE_LOG(LogNetworkNext, Display, TEXT("Network Next initialized"));
	}
}

FString FNetworkNextModule::GetEnvironmentVariable(const FString& EnvName)
{
	int32 ResultLength = 512;
	TCHAR* Result = new TCHAR[ResultLength];
	FPlatformMisc::GetEnvironmentVariable(*EnvName, Result, ResultLength);
	return FString(ResultLength, Result);
}

bool FNetworkNextModule::IsNetworkNextSuccessfullyInitialized()
{
	return this->NetworkNextIsInitialized && this->NetworkNextIsSuccessfullyInitialized;
}

void FNetworkNextModule::SetConfig(const FNetworkNextConfig& Config)
{
	if (this->NetworkNextIsInitialized)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Network Next is already initialized; you can not call SetConfig now."));
		return;
	}

	this->NetworkNextConfig = new FNetworkNextConfig(Config);
}

void FNetworkNextModule::SetServerConfig(const FNetworkNextServerConfig& ServerConfig)
{
	this->NetworkNextServerConfig = new FNetworkNextServerConfig(ServerConfig);
}

const FNetworkNextServerConfig* FNetworkNextModule::GetServerConfig() const
{
	return this->NetworkNextServerConfig;
}

void FNetworkNextModule::ShutdownNetworkNextAfterPIE(bool bIsSimulating)
{
	if (this->NetworkNextIsInitialized)
	{
		next_term();
	}

	this->NetworkNextIsInitialized = false;
	this->NetworkNextIsSuccessfullyInitialized = false;
}

void FNetworkNextModule::ShutdownModule()
{
	DestroyNetworkNextSocketSubsystem();

	if (this->NetworkNextIsInitialized)
	{
		next_term();
	}

#if defined(NETWORKNEXT_ENABLE_DELAY_LOAD_PATH)
	FPlatformProcess::FreeDllHandle(this->NetworkNextHandle);
	this->NetworkNextHandle = nullptr;
#endif
}

void FNetworkNextModule::NextLogFunction(int level, const char* format, ...)
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

IMPLEMENT_MODULE(FNetworkNextModule, NetworkNext)
