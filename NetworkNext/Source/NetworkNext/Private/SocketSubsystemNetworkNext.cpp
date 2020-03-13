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

#include "SocketSubsystemNetworkNext.h"
#include "Misc/ConfigCacheIni.h"
#include "SocketNetworkNext.h"
#include "SocketSubsystemModule.h"
#include "SocketNetworkNextClient.h"
#include "SocketNetworkNextServer.h"

FSocketSubsystemNetworkNext* FSocketSubsystemNetworkNext::SocketSingleton = nullptr;

/**
 * Create the socket subsystem for the given platform service
 */
FName CreateNetworkNextSocketSubsystem()
{
	// Create and register our singleton factory with the main online subsystem for easy access
	FSocketSubsystemNetworkNext* SocketSubsystem = FSocketSubsystemNetworkNext::Create();
	FString Error;
	if (SocketSubsystem->Init(Error))
	{
		FSocketSubsystemModule& SSS = FModuleManager::LoadModuleChecked<FSocketSubsystemModule>("Sockets");
		SSS.RegisterSocketSubsystem(NETWORKNEXT_SUBSYSTEM, SocketSubsystem, true /* Make this subsystem the default */);
		return NETWORKNEXT_SUBSYSTEM;
	}
	else
	{
		FSocketSubsystemNetworkNext::Destroy();
		return NAME_None;
	}
}

/**
 * Tear down the socket subsystem for the given platform service
 */
void DestroyNetworkNextSocketSubsystem()
{
	FModuleManager& ModuleManager = FModuleManager::Get();

	if (ModuleManager.IsModuleLoaded("Sockets"))
	{
		FSocketSubsystemModule& SSS = FModuleManager::GetModuleChecked<FSocketSubsystemModule>("Sockets");
		SSS.UnregisterSocketSubsystem(NETWORKNEXT_SUBSYSTEM);
	}
	FSocketSubsystemNetworkNext::Destroy();
}

void FSocketSubsystemNetworkNext::InitializeNetworkNextIfRequired()
{
	FNetworkNextModule& NetworkNextModule = FModuleManager::LoadModuleChecked<FNetworkNextModule>("NetworkNext");
	return NetworkNextModule.InitializeNetworkNextIfRequired();
}

bool FSocketSubsystemNetworkNext::IsNetworkNextInitializedSuccessfully()
{
	FNetworkNextModule& NetworkNextModule = FModuleManager::LoadModuleChecked<FNetworkNextModule>("NetworkNext");
	return NetworkNextModule.IsNetworkNextSuccessfullyInitialized();
}

/** 
 * Singleton interface for this subsystem 
 * @return the only instance of this subsystem
 */
FSocketSubsystemNetworkNext* FSocketSubsystemNetworkNext::Create()
{
	if (SocketSingleton == nullptr)
	{
		SocketSingleton = new FSocketSubsystemNetworkNext();
	}

	return SocketSingleton;
}

/**
 * Performs socket clean up
 */
void FSocketSubsystemNetworkNext::Destroy()
{
	if (SocketSingleton != nullptr)
	{
		SocketSingleton->Shutdown();
		delete SocketSingleton;
		SocketSingleton = nullptr;
	}
}

bool FSocketSubsystemNetworkNext::Init(FString& Error)
{
	return true;
}

/**
 * Performs platform specific socket clean up
 */
void FSocketSubsystemNetworkNext::Shutdown()
{
	for (auto Socket : this->NetworkNextSockets)
	{
		Socket->Close();
	}

	this->NetworkNextSockets.Empty();
}

FSocket* FSocketSubsystemNetworkNext::CreateSocket(const FName& SocketType, const FString& SocketDescription
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_PROTOCOLTYPE)
	, ESocketProtocolFamily ProtocolType
#endif
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_FORCEUDP)
	, bool bForceUDP
#endif
)
{
	// This implementation is a fallback for any sockets created outside NetworkNextNetDriver.

	FSocket* NewSocket = nullptr;

	ISocketSubsystem* PlatformSocketSub = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	if (PlatformSocketSub)
	{
		NewSocket = PlatformSocketSub->CreateSocket(SocketType, SocketDescription
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_PROTOCOLTYPE)
			, ProtocolType
#endif
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_FORCEUDP)
			, bForceUDP
#endif
		);
	}

	if (!NewSocket)
	{
		UE_LOG(LogNetworkNext, Warning, TEXT("Failed to create socket %s [%s]"), *SocketType.ToString(), *SocketDescription);
	}

	return NewSocket;
}

FSocket* FSocketSubsystemNetworkNext::CreateSocketWithNetDriver(const FName& SocketType, const FString& SocketDescription, const UNetworkNextNetDriver* InNetDriver
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_PROTOCOLTYPE)
	, ESocketProtocolFamily ProtocolType
#endif
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_FORCEUDP)
	, bool bForceUDP
#endif
)
{
	// NOTE: The socket descriptions here are prefixed with a string that indicates
	// the underlying FSocket C++ type. The NetworkNextNetDriver looks at the description of
	// the resulting socket to determine whether it's safe to cast to back to SocketNetworkNextClient or
	// SocketNetworkNextServer. It has to check in this manner because FSocket does not inherit from
	// UObject, and C++ RTTI is not enabled in UE4.

	if (SocketType == FName("NetworkNextClientSocket"))
	{
		this->InitializeNetworkNextIfRequired();
		if (this->IsNetworkNextInitializedSuccessfully())
		{
			FString ModifiedSocketDescription = SocketDescription;
			ModifiedSocketDescription.InsertAt(0, TEXT("SOCKET_TYPE_NEXT_CLIENT_"));
#if defined(NETWORKNEXT_HAS_ESOCKETPROTOCOLFAMILY)
			FSocketNetworkNextClient* Socket = new FSocketNetworkNextClient(ModifiedSocketDescription, ProtocolType, InNetDriver);
#else
			FSocketNetworkNextClient* Socket = new FSocketNetworkNextClient(ModifiedSocketDescription, InNetDriver);
#endif
			AddSocket(Socket);
			return Socket;
		}
	}
	else if (SocketType == FName("NetworkNextServerSocket"))
	{
		this->InitializeNetworkNextIfRequired();
		if (this->IsNetworkNextInitializedSuccessfully())
		{
			FString ModifiedSocketDescription = SocketDescription;
			ModifiedSocketDescription.InsertAt(0, TEXT("SOCKET_TYPE_NEXT_SERVER_"));
#if defined(NETWORKNEXT_HAS_ESOCKETPROTOCOLFAMILY)
			FSocketNetworkNextServer* Socket = new FSocketNetworkNextServer(ModifiedSocketDescription, ProtocolType, InNetDriver);
#else
			FSocketNetworkNextServer* Socket = new FSocketNetworkNextServer(ModifiedSocketDescription, InNetDriver);
#endif
			AddSocket(Socket);
			return Socket;
		}
	}

	FString ModifiedSocketDescription = SocketDescription;
	ModifiedSocketDescription.InsertAt(0, TEXT("SOCKET_TYPE_NATIVE_"));
	return this->CreateSocket(SocketType, ModifiedSocketDescription
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_PROTOCOLTYPE)
		, ProtocolType
#endif
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_FORCEUDP)
		, bForceUDP
#endif
	);
}

/**
 * Cleans up a socket class
 *
 * @param Socket the socket object to destroy
 */
void FSocketSubsystemNetworkNext::DestroySocket(FSocket* Socket)
{
	// Possible non Network Next socket here PLATFORM_SOCKETSUBSYSTEM, but its just a pointer compare
	RemoveSocket((FSocketNetworkNext*)Socket);
	delete Socket;
}

#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_GETADDRESSINFO)
/**
 * Gets the address information of the given hostname and outputs it into an array of resolvable addresses.
 * It is up to the caller to determine which one is valid for their environment.
 *
 * @param HostName string version of the queryable hostname or ip address
 * @param ServiceName string version of a service name ("http") or a port number ("80")
 * @param QueryFlags What flags are used in making the getaddrinfo call. Several flags can be used at once by ORing the values together.
 *                   Platforms are required to translate this value into a the correct flag representation.
 * @param ProtocolType this is used to limit results from the call. Specifying None will search all valid protocols.
 *					   Callers will find they rarely have to specify this flag.
 * @param SocketType What socket type should the results be formatted for. This typically does not change any formatting results and can
 *                   be safely left to the default value.
 *
 * @return the array of results from GetAddrInfo
 */
FAddressInfoResult FSocketSubsystemNetworkNext::GetAddressInfo(const TCHAR* HostName, const TCHAR* ServiceName,
	EAddressInfoFlags QueryFlags, ESocketProtocolFamily ProtocolType,	ESocketType SocketType)
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->GetAddressInfo(HostName, ServiceName, QueryFlags, ProtocolType, SocketType);
}
#endif

/**
 * Does a DNS look up of a host name
 *
 * @param HostName the name of the host to look up
 * @param OutAddr the address to copy the IP address to
 */
ESocketErrors FSocketSubsystemNetworkNext::GetHostByName(const ANSICHAR* HostName, FInternetAddr& OutAddr)
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->GetHostByName(HostName, OutAddr);
}

/**
 * Determines the name of the local machine
 *
 * @param HostName the string that receives the data
 *
 * @return true if successful, false otherwise
 */
bool FSocketSubsystemNetworkNext::GetHostName(FString& HostName)
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->GetHostName(HostName);
}

/**
 *	Create a proper FInternetAddr representation
 * @param Address host address
 * @param Port host port
 */
TSharedRef<FInternetAddr> FSocketSubsystemNetworkNext::CreateInternetAddr(uint32 Address, uint32 Port)
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->CreateInternetAddr(Address, Port);
}

/**
 * @return Whether the machine has a properly configured network device or not
 */
bool FSocketSubsystemNetworkNext::HasNetworkDevice() 
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->HasNetworkDevice();
}

/**
 *	Get the name of the socket subsystem
 * @return a string naming this subsystem
 */
const TCHAR* FSocketSubsystemNetworkNext::GetSocketAPIName() const 
{
	return TEXT("NetworkNextSockets");
}

/**
 * Would usually return the last system error code - but we are doing our own error management in SDK so just return 'no error'
 * Note that we can't fall back to the default behaviour here because the last error defaults to SE_EINVAL on some platforms
 */
ESocketErrors FSocketSubsystemNetworkNext::GetLastErrorCode()
{
	return ESocketErrors::SE_NO_ERROR;
}

/**
 * Translates the platform error code to a ESocketErrors enum
 */
ESocketErrors FSocketSubsystemNetworkNext::TranslateErrorCode(int32 Code)
{
	ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
	return PlatformSubsystem->TranslateErrorCode(Code);
}

/**
 * Chance for the socket subsystem to get some time
 *
 * @param DeltaTime time since last tick
 */
bool FSocketSubsystemNetworkNext::Tick(float DeltaTime)
{	
    QUICK_SCOPE_CYCLE_COUNTER(STAT_SocketSubsystemNetworkNext_Tick);

	for (auto Socket : this->NetworkNextSockets)
	{
		Socket->UpdateNetworkNextSocket();
	}

	return true;
}