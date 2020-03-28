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

#include "NetworkNextBuildConfig.h"
#include "CoreMinimal.h"
#include "UObject/WeakObjectPtr.h"
#include "SocketSubsystem.h"
#include "IPAddress.h"
#include "Containers/Ticker.h"

class Error;
class UNetworkNextNetDriver;

class FSocketSubsystemNetworkNext : public ISocketSubsystem, public FTickerObjectBase
{
private:

	void InitializeNetworkNextIfRequired();

	bool IsNetworkNextInitializedSuccessfully();

protected:

	/** Single instantiation of this subsystem */
	static FSocketSubsystemNetworkNext* SocketSingleton;

	/** Tracks Network Next sockets, so we can call update on them */
	TArray<class FSocketNetworkNext*> NetworkNextSockets;

	void AddSocket(class FSocketNetworkNext* InSocket)
	{
		NetworkNextSockets.Add(InSocket);
	}

	void RemoveSocket(class FSocketNetworkNext* InSocket)
	{
		NetworkNextSockets.RemoveSingleSwap(InSocket);
	}
	
public:

	/** 
	 * Singleton interface for this subsystem 
	 * @return the only instance of this subsystem
	 */
	static FSocketSubsystemNetworkNext* Create();

	static void Destroy();

	FSocketSubsystemNetworkNext()
	{
	}

	virtual bool Init(FString& Error) override;

	virtual void Shutdown() override;

	// Only implemented so that we implement the interface. The NetworkNextNetDriver actually uses the method that passes a reference to itself (seen below).
	virtual class FSocket* CreateSocket(const FName& SocketType, const FString& SocketDescription
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_PROTOCOLTYPE)
		, ESocketProtocolFamily ProtocolType
#endif
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_FORCEUDP)
		, bool bForceUDP
#endif
	) override;

	class FSocket* CreateSocketWithNetDriver(const FName& SocketType, const FString& SocketDescription, const UNetworkNextNetDriver* InNetDriver
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_PROTOCOLTYPE)
		, ESocketProtocolFamily ProtocolType
#endif
#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_FORCEUDP)
		, bool bForceUDP
#endif	
	);

	/**
	 * Cleans up a socket class
	 *
	 * @param Socket the socket object to destroy
	 */
	virtual void DestroySocket(class FSocket* Socket) override;

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
	virtual FAddressInfoResult GetAddressInfo(const TCHAR* HostName, const TCHAR* ServiceName = nullptr,
		EAddressInfoFlags QueryFlags = EAddressInfoFlags::Default,
		ESocketProtocolFamily ProtocolType = ESocketProtocolFamily::None,
		ESocketType SocketType = ESocketType::SOCKTYPE_Unknown) override;
#endif

	/**
	 * Does a DNS look up of a host name
	 *
	 * @param HostName the name of the host to look up
	 * @param OutAddr the address to copy the IP address to
	 */
	virtual ESocketErrors GetHostByName(const ANSICHAR* HostName, FInternetAddr& OutAddr) override;

	/**
	 * Some platforms require chat data (voice, text, etc.) to be placed into
	 * packets in a special way. This function tells the net connection
	 * whether this is required for this platform
	 */
	virtual bool RequiresChatDataBeSeparate() override
	{
		return false;
	}

	/**
	 * Some platforms require packets be encrypted. This function tells the
	 * net connection whether this is required for this platform
	 */
	virtual bool RequiresEncryptedPackets() override
	{
		return false;
	}

	/**
	 * Determines the name of the local machine
	 *
	 * @param HostName the string that receives the data
	 *
	 * @return true if successful, false otherwise
	 */
	virtual bool GetHostName(FString& HostName) override;

	/**
	 *	Create a proper FInternetAddr representation
	 * @param Address host address
	 * @param Port host port
	 */
	virtual TSharedRef<FInternetAddr> CreateInternetAddr(uint32 Address=0, uint32 Port=0) override;

	/**
	 * @return Whether the machine has a properly configured network device or not
	 */
	virtual bool HasNetworkDevice() override;

	/**
	 *	Get the name of the socket subsystem
	 * @return a string naming this subsystem
	 */
	virtual const TCHAR* GetSocketAPIName() const override;

	/**
	* Would usually return the last system error code - but we are doing our own error management in SDK so just return 'no error'
	* Note that we can't fall back to the default behaviour here because the last error defaults to SE_EINVAL on some platforms
	*/
	virtual ESocketErrors GetLastErrorCode() override;

	/**
	 * Translates the platform error code to a ESocketErrors enum
	 */
	virtual ESocketErrors TranslateErrorCode(int32 Code) override;

	/**
	 * Gets the list of addresses associated with the adapters on the local computer.
	 *
	 * @param OutAdresses - Will hold the address list.
	 *
	 * @return true on success, false otherwise.
	 */
	virtual bool GetLocalAdapterAddresses( TArray<TSharedPtr<FInternetAddr> >& OutAdresses ) override
	{
		bool bCanBindAll;

		OutAdresses.Add(GetLocalHostAddr(*GLog, bCanBindAll));

		return true;
	}

	/**
	 * Chance for the socket subsystem to get some time
	 *
	 * @param DeltaTime time since last tick
	 */
	virtual bool Tick(float DeltaTime) override;

#if defined(NETWORKNEXT_SOCKETSUBSYSTEM_INTERFACE_HAS_ISSOCKETWAITSUPPORTED)
	/**
	 * Waiting on a socket is not supported.
	 */
	virtual bool IsSocketWaitSupported() const override { return false; }
#endif
};

/**
 * Create the socket subsystem for the given platform service
 */
FName CreateNetworkNextSocketSubsystem();

/**
 * Tear down the socket subsystem for the given platform service
 */
void DestroyNetworkNextSocketSubsystem();
