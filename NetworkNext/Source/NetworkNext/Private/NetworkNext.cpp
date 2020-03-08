/*
    Network Next SDK 3.1.3

    Copyright © 2017 - 2019 Network Next, Inc.

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
#include "Modules/ModuleManager.h"
#include "Interfaces/IPluginManager.h"
#if defined(NETWORKNEXT_INCLUDE_NEXT_H_WITH_SHORT_PATH)
#include "next.h"
#else
#include "NetworkNextLibrary/next/include/next.h"
#endif
#include "SocketSubsystemNetworkNext.h"
#include <cstdarg>

DEFINE_LOG_CATEGORY(LogNetworkNext);

void FNetworkNextModule::StartupModule()
{
	next_log_function(&FNetworkNextModule::NextLogFunction);
	next_log_level(NEXT_LOG_LEVEL_DEBUG);

	if (next_init() != NEXT_OK)
	{
		UE_LOG(LogNetworkNext, Error, TEXT("Network Next could not be initialized."));
		return;
	}

	CreateNetworkNextSocketSubsystem();
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

void FNetworkNextModule::ShutdownModule()
{
	DestroyNetworkNextSocketSubsystem();

	next_term();
}

IMPLEMENT_MODULE(FNetworkNextModule, NetworkNext)
