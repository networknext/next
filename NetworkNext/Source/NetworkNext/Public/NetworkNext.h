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
#include "NetworkNextConfig.h"
#include "NetworkNextServerConfig.h"
#include "Modules/ModuleManager.h"

DECLARE_LOG_CATEGORY_EXTERN(LogNetworkNext, Display, Display);

#ifndef NETWORKNEXT_SUBSYSTEM
#define NETWORKNEXT_SUBSYSTEM FName(TEXT("NETWORKNEXT"))
#endif

class FNetworkNextModule : public IModuleInterface
{
public:

	/** IModuleInterface implementation */
	virtual void StartupModule() override;
	virtual void ShutdownModule() override;

	void InitializeNetworkNextIfRequired();
	bool IsNetworkNextSuccessfullyInitialized();
	void SetConfig(const FNetworkNextConfig& Config);
	void SetServerConfig(const FNetworkNextServerConfig& ServerConfig);
	const FNetworkNextServerConfig* GetServerConfig() const;

private:
	
	void ShutdownNetworkNextAfterPIE(bool bIsSimulating);

	/** The static handler for log messages coming from the Network Next library */
	static void NextLogFunction(int level, const char* format, ...);

	/** Static handlers for memory allocations */
	static void* Malloc(void* context, size_t size);
	static void Free(void* context, void* src);

	bool NetworkNextIsInitialized;
	bool NetworkNextIsSuccessfullyInitialized;

	UPROPERTY()
	FNetworkNextConfig* NetworkNextConfig;

	UPROPERTY()
	FNetworkNextServerConfig* NetworkNextServerConfig;

#if defined(NETWORKNEXT_ENABLE_DELAY_LOAD)
	void* NetworkNextHandle;
#endif

	FString GetEnvironmentVariable(const FString& EnvName);
};
