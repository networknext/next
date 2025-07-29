/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "NetworkNextModule.h"
#include "NetworkNextNetDriver.h"
#include "NetworkNextSocketSubsystem.h"
#include "NetworkNextSocketServer.h"
#include "Core.h"
#include "next.h"

DEFINE_LOG_CATEGORY(LogNetworkNext);

IMPLEMENT_MODULE(FNetworkNextModule, NetworkNext)

// gfiedler: Autodetect calls to this function are *always* made from the same background thread and never run in ||, hence it is safe for us to use statics here

static volatile int unreal_http_request_complete;

void HandleHttpRequestComplete(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful)
{
    (void) Request;
    (void) Response;
    (void) bWasSuccessful;
    unreal_http_request_complete = 1;
}

static bool unreal_http_request_function( const char * url, const char * header, char * output, size_t output_size )
{
    FString URL(url);
    FString Header(header);

    UE_LOG(LogNetworkNext, Display, TEXT("unreal_http_request: URL='%s', Header='%s'"), *URL, *Header);

    FHttpModule & HttpModule = FHttpModule::Get();

    FHttpRequestPtr Request = HttpModule.CreateRequest();

    Request->SetVerb(TEXT("GET"));

    if (header[0] != '\0')
    {
        Request->SetHeader(TEXT("Metadata-Flavor"), TEXT("Google"));
    }

    Request->SetURL(URL);

    unreal_http_request_complete = 0;

    Request->OnProcessRequestComplete().BindStatic(&HandleHttpRequestComplete);

    if (!Request->ProcessRequest())
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Failed to start HTTP Request"));
        return false;
    }

    while (!unreal_http_request_complete)
    {
        FPlatformProcess::Sleep(0.1);
    }

    FHttpResponsePtr Response = Request->GetResponse();

    if (Response && Response->GetResponseCode() == 200)
    {
        FString ResponseData = Response->GetContentAsString();
        next_copy_string( output, TCHAR_TO_ANSI(*ResponseData), output_size );
        return true;
    }
    else
    {
        return false;
    }
}

void FNetworkNextModule::StartupModule()
{
    UE_LOG(LogNetworkNext, Display, TEXT("Network Next Plugin loaded"));

    m_initialized_sdk = false;

    next_config_t config;
    next_default_config( &config );

    const char * buyer_public_key = "XXX";

    next_copy_string(config.buyer_public_key, buyer_public_key, sizeof(config.buyer_public_key));

#if WITH_SERVER_CODE
    // IMPORTANT: You must never expose your private key to your players!
    const char * buyer_private_key = "XXXXXX";
    next_copy_string(config.buyer_private_key, buyer_private_key, sizeof(config.buyer_private_key));
#endif // #if WITH_SERVER

    next_log_level(NEXT_LOG_LEVEL_DEBUG);

    // IMPORTANT: Use Unreal HTTP module instead of "curl" for autodetecting server datacenter
    next_autodetect_http_request_function(unreal_http_request_function);

    if (next_init(NULL, &config) != NEXT_OK)
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Network Next SDK failed to initalize!"));
        return;
    }

    UE_LOG(LogNetworkNext, Display, TEXT("Network Next SDK initialized"));

    m_initialized_sdk = true;

    // Perform network next allocations through the unreal allocator instead of default malloc/free
    next_allocator(&FNetworkNextModule::Malloc, &FNetworkNextModule::Free);

    // Setup logging to go to the "NetworkNext" log category
    next_log_function(&FNetworkNextModule::Log);

    CreateNetworkNextSocketSubsystem();
}

void FNetworkNextModule::ShutdownModule()
{
    UE_LOG(LogNetworkNext, Display, TEXT("Network Next Plugin shutting down"));

    DestroyNetworkNextSocketSubsystem();

    next_term();

    m_initialized_sdk = false;

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
    default:
        UE_LOG(LogNetworkNext, Display, TEXT("%s"), *Message);
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

void FNetworkNextModule::UpgradePlayer(AController* Controller, const FString& UserId)
{
    UE_LOG(LogNetworkNext, Display, TEXT("FNetworkNextModule::UpgradePlayer"));

    if (Controller == nullptr)
    {
        UE_LOG(LogNetworkNext, Display, TEXT("Controller is NULL"));
        return;
    }

    UWorld* World = Controller->GetWorld();

    if (World == nullptr)
    {
        UE_LOG(LogNetworkNext, Display, TEXT("World is NULL"));
        return;
    }

    UNetworkNextNetDriver* NetDriver = Cast<UNetworkNextNetDriver>(World->GetNetDriver());

    if (NetDriver == nullptr)
    {
        UE_LOG(LogNetworkNext, Display, TEXT("NetDriver is NULL"));
        return;
    }

    FNetworkNextSocketServer* ServerSocket = NetDriver->GetServerSocket();

    if (ServerSocket == nullptr)
    {
        UE_LOG(LogNetworkNext, Display, TEXT("ServerSocket is NULL"));
        return;
    }

    UNetConnection* Connection = Controller->GetNetConnection();

    if (Connection == nullptr)
    {
        UE_LOG(LogNetworkNext, Display, TEXT("Connection is NULL"));
        return;
    }

    ServerSocket->UpgradeClient(Connection->GetRemoteAddr(), UserId);

    UE_LOG(LogNetworkNext, Display, TEXT("Upgraded Player!"));
}
