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

#include "NetworkNextSocketPassthrough.h"

FNetworkNextSocketPassthrough::FNetworkNextSocketPassthrough(const FString& InSocketDescription, const FName& InSocketProtocol)
    : FNetworkNextSocket(ENetworkNextSocketType::TYPE_Passthrough, InSocketDescription, InSocketProtocol)
{
    UE_LOG(LogNetworkNext, Display, TEXT("Passthrough socket created"));
    bBound = false;
    PlatformSocket = NULL;
    ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
    if (!PlatformSubsystem)
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Could not find platform socket subsystem"));
        return;
    }
    PlatformSocket = PlatformSubsystem->CreateSocket(NAME_DGram, InSocketDescription, InSocketProtocol);
    if (!PlatformSocket)
    {
        UE_LOG(LogNetworkNext, Error, TEXT("Could not create internal platform socket"));
        return;
    }
}

FNetworkNextSocketPassthrough::~FNetworkNextSocketPassthrough()
{
    Close();
    ISocketSubsystem* PlatformSubsystem = ISocketSubsystem::Get(PLATFORM_SOCKETSUBSYSTEM);
    if (PlatformSubsystem)
    {
        PlatformSubsystem->DestroySocket(PlatformSocket);
        PlatformSocket = NULL;
    }
    UE_LOG(LogNetworkNext, Display, TEXT("Passthrough socket destroyed"));
}

bool FNetworkNextSocketPassthrough::Close()
{
    return PlatformSocket ? PlatformSocket->Close() : true;
}

bool FNetworkNextSocketPassthrough::Bind(const FInternetAddr& Addr)
{
    bBound = PlatformSocket ? PlatformSocket->Bind(Addr) : false;
    return bBound;
}

bool FNetworkNextSocketPassthrough::SendTo(const uint8* Data, int32 Count, int32& BytesSent, const FInternetAddr& Destination)
{
    return PlatformSocket ? PlatformSocket->SendTo(Data, Count, BytesSent, Destination) : false;
}

bool FNetworkNextSocketPassthrough::RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags)
{
    return PlatformSocket ? PlatformSocket->RecvFrom(Data, BufferSize, BytesRead, Source, Flags) : false;
}

void FNetworkNextSocketPassthrough::GetAddress(FInternetAddr& OutAddr)
{
    if (PlatformSocket && bBound)
    {
        PlatformSocket->GetAddress(OutAddr);
    }
    else
    {
        // Not bound. We don't know the address yet.
        bool bIsValid;
        OutAddr.SetIp(TEXT("0.0.0.0"), bIsValid);
    }
}

int32 FNetworkNextSocketPassthrough::GetPortNo()
{
    return PlatformSocket ? PlatformSocket->GetPortNo() : 0;
}
