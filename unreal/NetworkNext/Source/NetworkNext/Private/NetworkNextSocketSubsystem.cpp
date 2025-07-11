/*
	Network Next. Copyright 2017 - 2025 Network Next, Inc.  
	Licensed under the Network Next Source Available License 1.0
*/

#include "NetworkNextSocketSubsystem.h"
#include "NetworkNextSocket.h"
#include "NetworkNextSocketClient.h"
#include "NetworkNextSocketServer.h"
#include "NetworkNextModule.h"
#include "SocketSubsystemModule.h"

FName CreateNetworkNextSocketSubsystem()
{
	FNetworkNextSocketSubsystem* PlatformSubsystem = FNetworkNextSocketSubsystem::Create();
	FString Error;
	if (PlatformSubsystem->Init(Error))
	{
		FSocketSubsystemModule & sockets = FModuleManager::LoadModuleChecked<FSocketSubsystemModule>("Sockets");
		sockets.RegisterSocketSubsystem(NETWORKNEXT_SUBSYSTEM, PlatformSubsystem, true);
		return NETWORKNEXT_SUBSYSTEM;
	}
	else
	{
		FNetworkNextSocketSubsystem::Destroy();
		return NAME_None;
	}
}

void DestroyNetworkNextSocketSubsystem()
{
	FModuleManager& ModuleManager = FModuleManager::Get();
	if (ModuleManager.IsModuleLoaded("Sockets"))
	{
		FSocketSubsystemModule& socketSubsystem = FModuleManager::GetModuleChecked<FSocketSubsystemModule>("Sockets");
		socketSubsystem.UnregisterSocketSubsystem(NETWORKNEXT_SUBSYSTEM);
	}
	FNetworkNextSocketSubsystem::Destroy();
}

static FNetworkNextSocketSubsystem* SocketSingleton;

FNetworkNextSocketSubsystem* FNetworkNextSocketSubsystem::Create()
{
	if (!SocketSingleton)
	{
		SocketSingleton = new FNetworkNextSocketSubsystem();
	}

	return SocketSingleton;
}

void FNetworkNextSocketSubsystem::Destroy()
{
	if (SocketSingleton)
	{
		SocketSingleton->Shutdown();
		delete SocketSingleton;
		SocketSingleton = NULL;
	}
}

bool FNetworkNextSocketSubsystem::Init(FString& Error)
{
	UE_LOG(LogNetworkNext, Display, TEXT("Network Next socket subsystem initializing"));

	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);

	UE_LOG(LogNetworkNext, Display, TEXT("Platform default socket subsystem is %s"), PlatformSubsystem->GetSocketAPIName());

	return true;
}

void FNetworkNextSocketSubsystem::Shutdown()
{
	UE_LOG(LogNetworkNext, Display, TEXT("Network Next socket subsystem shutdown"));

	for (auto Socket : this->NetworkNextSockets)
	{
		Socket->Close();
	}

	NetworkNextSockets.Empty();
}

FSocket* FNetworkNextSocketSubsystem::CreateSocket(const FName& SocketType, const FString& SocketDescription, const FName& ProtocolName)
{
	if (SocketType == FName("NetworkNextSocketServer"))
	{
		// server socket
		UE_LOG(LogNetworkNext, Display, TEXT("Create network next server socket"));
		FString ModifiedSocketDescription = SocketDescription;
		ModifiedSocketDescription.InsertAt(0, TEXT("SOCKET_TYPE_NETWORK_NEXT_SERVER_"));
		FNetworkNextSocketServer* Socket = new FNetworkNextSocketServer(ModifiedSocketDescription, ProtocolName);
		NetworkNextSockets.Add(Socket);
		return Socket;
	}
	else if (SocketType == FName("NetworkNextSocketClient"))
	{
		// client socket
		UE_LOG(LogNetworkNext, Display, TEXT("Create network next client socket"));
		FString ModifiedSocketDescription = SocketDescription;
		ModifiedSocketDescription.InsertAt(0, TEXT("SOCKET_TYPE_NETWORK_NEXT_CLIENT_"));
		FNetworkNextSocketClient* Socket = new FNetworkNextSocketClient(ModifiedSocketDescription, ProtocolName);
		NetworkNextSockets.Add(Socket);
		return Socket;
	}
	else
	{
		// platform socket
		UE_LOG(LogNetworkNext, Display, TEXT("Create platform socket"));
		ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
		return PlatformSubsystem->CreateSocket(SocketType, SocketDescription, ProtocolName);
	}
}

void FNetworkNextSocketSubsystem::DestroySocket(FSocket* Socket)
{
	if (NetworkNextSockets.RemoveSingleSwap((FNetworkNextSocket*)Socket))
	{
		// network next socket
		UE_LOG(LogNetworkNext, Display, TEXT("Destroy network next socket"));
		delete Socket;
	}
	else
	{
		// platform socket
		UE_LOG(LogNetworkNext, Display, TEXT("Destroy platform socket"));
		ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
		PlatformSubsystem->DestroySocket(Socket);
	}
}

FAddressInfoResult FNetworkNextSocketSubsystem::GetAddressInfo(const TCHAR* HostName, const TCHAR* ServiceName, EAddressInfoFlags QueryFlags, const FName ProtocolTypeName, ESocketType SocketType)
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->GetAddressInfo(HostName, ServiceName, QueryFlags, ProtocolTypeName, SocketType);
}

TSharedPtr<FInternetAddr> FNetworkNextSocketSubsystem::GetAddressFromString(const FString& InAddress)
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->GetAddressFromString(InAddress);
}

bool FNetworkNextSocketSubsystem::GetHostName(FString& HostName)
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->GetHostName(HostName);
}

TSharedRef<FInternetAddr> FNetworkNextSocketSubsystem::CreateInternetAddr()
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->CreateInternetAddr();
}

bool FNetworkNextSocketSubsystem::HasNetworkDevice()
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->HasNetworkDevice();
}

const TCHAR* FNetworkNextSocketSubsystem::GetSocketAPIName() const
{
	return TEXT("Network Next");
}

ESocketErrors FNetworkNextSocketSubsystem::GetLastErrorCode()
{
	return ESocketErrors::SE_NO_ERROR;
}

ESocketErrors FNetworkNextSocketSubsystem::TranslateErrorCode(int32 Code)
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->TranslateErrorCode(Code);
}

bool FNetworkNextSocketSubsystem::Tick(float DeltaTime)
{
	for (auto Socket : NetworkNextSockets)
	{
		Socket->Update();
	}
	return true;
}
