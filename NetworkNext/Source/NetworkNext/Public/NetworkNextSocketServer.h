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

#if !NEXT_PASSTHROUGH

#include "NetworkNextSocket.h"
#include "next.h"

class FNetworkNextSocketServer : public FNetworkNextSocket
{
private:

    next_server_t* NetworkNextServer;

    FString ServerAddress;

    struct PacketData {
        next_address_t from;
        const uint8_t* packet_data;
        int packet_bytes;
    };

    TQueue<PacketData> PacketQueue;

    uint32_t PacketQueueBytes;

public:

    FNetworkNextSocketServer(const FString& InSocketDescription, const FName& InSocketProtocol);

    virtual ~FNetworkNextSocketServer();

    virtual void UpdateNetworkNextSocket() override;

    /**
     * Closes the socket
     *
     * @param true if it closes without errors, false otherwise
     */
    virtual bool Close() override;

    /**
     * Binds a socket to a network byte ordered address
     *
     * @param Addr the address to bind to
     *
     * @return true if successful, false otherwise
     */
    virtual bool Bind(const FInternetAddr& Addr) override;

    /**
     * Sends a buffer to a network byte ordered address
     *
     * @param Data the buffer to send
     * @param Count the size of the data to send
     * @param BytesSent out param indicating how much was sent
     * @param Destination the network byte ordered address to send to
     */
    virtual bool SendTo(const uint8* Data, int32 Count, int32& BytesSent, const FInternetAddr& Destination) override;

    /**
     * Reads a packet from the socket.
     *
     * @param Data the buffer to read into
     * @param BufferSize the max size of the buffer
     * @param BytesRead out param indicating how many bytes were read from the socket
     * @param Source out param receiving the address of the sender of the data
     * @param Flags the receive flags (must be ESocketReceiveFlags::None)
     */
    virtual bool RecvFrom(uint8* Data, int32 BufferSize, int32& BytesRead, FInternetAddr& Source, ESocketReceiveFlags::Type Flags = ESocketReceiveFlags::None) override;

    /**
     * Gets the address the socket is bound to
     *
     * @param OutAddr address the socket is bound to
     */
    virtual void GetAddress(FInternetAddr& OutAddr) override;

    /**
     * Gets the port number this socket is bound to
     */
    virtual int32 GetPortNo() override;

    // Callback for packets received from the network next server

    static void OnPacketReceived(next_server_t* server, void* context, const next_address_t* from, const uint8_t* packet_data, int packet_bytes);
};

#endif // #if !NEXT_PASSTHROUGH
