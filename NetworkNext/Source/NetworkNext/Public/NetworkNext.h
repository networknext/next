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

#include "Modules/ModuleManager.h"

DECLARE_LOG_CATEGORY_EXTERN(LogNetworkNext, Display, Display);

#ifndef NETWORKNEXT_SUBSYSTEM
#define NETWORKNEXT_SUBSYSTEM FName(TEXT("NETWORKNEXT"))
#endif

class FNetworkNextModule : public IModuleInterface
{
public:

	/** Called right after the module DLL has been loaded and the module object has been created */
	virtual void StartupModule() override;
	
	/** Called before the module is unloaded, right before the module object is destroyed */
	virtual void ShutdownModule() override;

	/** This plugin does not support dynamic reloading */
	virtual bool SupportsDynamicReloading() override {
		return false;
	};

private:

	/** The static handler for logs coming from the Network Next SDK */
	static void Log(int level, const char* format, ...);

	/** Static handlers for memory allocations from the Network Next SDK */
	static void* Malloc(void* context, size_t size);
	static void Free(void* context, void* src);

	/** Have we initialized the Network Next SDK? */
	bool m_initialized_sdk = false;
};
