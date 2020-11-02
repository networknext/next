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

#include "NetworkNextSocketSubsystem.h"
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
	UE_LOG(LogNetworkNext, Display, TEXT("Network Next socket subsystem initialized"));

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
#if !NEXT_PASSTHROUGH

    if (SocketType == FName("NetworkNextSocketServer"))
    {
        // server socket
        UE_LOG(LogNetworkNext, Display, TEXT("Create network next server socket"));
        FString ModifiedSocketDescription = SocketDescription;
        ModifiedSocketDescription.InsertAt(0, TEXT("SOCKET_TYPE_NEXT_SERVER_"));
        FNetworkNextSocketServer* Socket = new FNetworkNextSocketServer(ModifiedSocketDescription, ProtocolName);
        NetworkNextSockets.Add(Socket);
        return Socket;
    }
    else if (SocketType == FName("NetworkNextSocketClient"))
    {
        // client socket
        UE_LOG(LogNetworkNext, Display, TEXT("Create network next client socket"));
        FString ModifiedSocketDescription = SocketDescription;
        ModifiedSocketDescription.InsertAt(0, TEXT("SOCKET_TYPE_NEXT_CLIENT_"));
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

#else // #if !NEXT_PASSTHROUGH

    if (SocketType == FName("NetworkNextSocketServer") || SocketType == FName("NetworkNextSocketClient"))
    {
        // passthrough socket
        UE_LOG(LogNetworkNext, Display, TEXT("Create passthrough socket"));
        FString ModifiedSocketDescription = SocketDescription;
        ModifiedSocketDescription.InsertAt(0, TEXT("SOCKET_TYPE_NEXT_PASSTHROUGH_"));
        FNetworkNextSocketPassthrough* Socket = new FNetworkNextSocketPassthrough(ModifiedSocketDescription, ProtocolName);
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

#endif // #if !NEXT_PASSTHROUGH
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
        UE_LOG(LogNetworkNext, Display, TEXT("Destroyed platform socket"));
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
    // todo: make sure this system ticks *before* the system that is processing packets for this frame
    for (auto Socket : NetworkNextSockets)
    {
        Socket->UpdateNetworkNextSocket();
    }
    return true;
}
