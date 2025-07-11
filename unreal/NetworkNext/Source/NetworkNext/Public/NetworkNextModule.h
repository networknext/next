/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
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

    /** Call this from game mode OnPostLogin to enable acceleration for this player */
	static void NETWORKNEXT_API UpgradePlayer(AController* Controller, const FString& UserId);

private:

    /** The static handler for logs from the Network Next SDK */
    static void Log(int level, const char* format, ...);

    /** Static handlers for memory allocations from the Network Next SDK */
    static void* Malloc(void* context, size_t size);
    static void Free(void* context, void* src);

    /** Have we initialized the Network Next SDK? */
    bool m_initialized_sdk = false;
};
