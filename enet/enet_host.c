/** 
 @file host.c
 @brief ENet host management functions
*/
#define ENET_BUILDING_LIB 1
#include <string.h>
#include "enet.h"

#if ENET_NETWORK_NEXT

ENetAddress enet_address_from_next( const struct next_address_t * in )
{
    next_assert( in );
    next_assert( in->type == NEXT_ADDRESS_IPV4 );   // enet only supports ipv4
    ENetAddress out;
    out.host =  ((enet_uint32)in->data.ipv4[0])         | 
               (((enet_uint32)in->data.ipv4[1]) << 8  ) | 
               (((enet_uint32)in->data.ipv4[2]) << 16 ) |
               (((enet_uint32)in->data.ipv4[3]) << 24 );
    out.port = in->port;
    return out;
}

void enet_address_to_next( const ENetAddress * in, struct next_address_t * out )
{
    next_assert( in );
    next_assert( out );
    out->type = NEXT_ADDRESS_IPV4;
    out->port = in->port;
    out->data.ipv4[0] = in->host & 0xFF;
    out->data.ipv4[1] = ( in->host >> 8 )  & 0xFF;
    out->data.ipv4[2] = ( in->host >> 16 ) & 0xFF;
    out->data.ipv4[3] = ( in->host >> 24 ) & 0xFF;
}

void enet_network_next_client_packet_received( struct next_client_t * client, void * context, const struct next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; 
    (void) context; 
    (void) packet_data; 
    (void) packet_bytes;

    next_printf( NEXT_LOG_LEVEL_DEBUG, "network next client received packet from server (%d bytes)", packet_bytes );

    ENetHost * host = (ENetHost*) context;

    struct ENetInternalPacket * packet = enet_malloc( sizeof( struct ENetInternalPacket ) );
    
    next_assert( packet );
    
    if ( !packet )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "could not allocate packet for client packet receive" );
        return;
    }

    packet->from = enet_address_from_next( from );
    next_assert( packet_bytes > 0 );
    next_assert( packet_bytes <= NEXT_MTU );
    memcpy( packet->data, packet_data, packet_bytes );
    packet->size = packet_bytes;

    enet_list_insert( enet_list_end( &host->receivePacketQueue), packet );
}

void enet_network_next_server_packet_received( struct next_server_t * server, void * context, const struct next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) server;
    (void) context;
    (void) from;
    (void) packet_data;

    next_printf( NEXT_LOG_LEVEL_DEBUG, "network next server received packet from client (%d bytes)", packet_bytes );

    ENetHost * host = (ENetHost*) context;

    struct ENetInternalPacket * packet = enet_malloc( sizeof( struct ENetInternalPacket ) );
    
    next_assert( packet );
    
    if ( !packet )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "could not allocate packet for server packet receive" );
        return;
    }

    packet->from = enet_address_from_next( from );
    next_assert( packet_bytes > 0 );
    next_assert( packet_bytes <= NEXT_MTU );
    memcpy( packet->data, packet_data, packet_bytes );
    packet->size = packet_bytes;

    enet_list_insert( enet_list_end( &host->receivePacketQueue ), packet );
}

#endif // #if ENET_NETWORK_NEXT

/** @defgroup host ENet host functions
    @{
*/

/** Creates a host for communicating to peers.  

    @param address   the address at which other peers may connect to this host.  If NULL, then no peers may connect to the host.
    @param peerCount the maximum number of peers that should be allocated for the host.
    @param channelLimit the maximum number of channels allowed; if 0, then this is equivalent to ENET_PROTOCOL_MAXIMUM_CHANNEL_COUNT
    @param incomingBandwidth downstream bandwidth of the host in bytes/second; if 0, ENet will assume unlimited bandwidth.
    @param outgoingBandwidth upstream bandwidth of the host in bytes/second; if 0, ENet will assume unlimited bandwidth.

    @returns the host on success and NULL on failure

    @remarks ENet will strategically drop packets on specific sides of a connection between hosts
    to ensure the host's bandwidth is not overwhelmed.  The bandwidth parameters also determine
    the window size of a connection which limits the amount of reliable packets that may be in transit
    at any given time.
*/
ENetHost *
#if ENET_NETWORK_NEXT
enet_host_create (const ENetHostConfig * config, size_t peerCount, size_t channelLimit, enet_uint32 incomingBandwidth, enet_uint32 outgoingBandwidth)
#else // #if ENET_NETWORK_NEXT
enet_host_create (const ENetAddress * address, size_t peerCount, size_t channelLimit, enet_uint32 incomingBandwidth, enet_uint32 outgoingBandwidth)
#endif // #if ENET_NETWORK_NEXT
{
    ENetHost * host;
    ENetPeer * currentPeer;

    if (peerCount > ENET_PROTOCOL_MAXIMUM_PEER_ID)
      return NULL;

#if ENET_NETWORK_NEXT
    next_assert( config );
    if ( !config )
    {
        return NULL;        
    }
#endif // #if ENET_NETWORK_NEXT

    host = (ENetHost *) enet_malloc (sizeof (ENetHost));
    if (host == NULL)
      return NULL;
    memset (host, 0, sizeof (ENetHost));

    host -> peers = (ENetPeer *) enet_malloc (peerCount * sizeof (ENetPeer));
    if (host -> peers == NULL)
    {
       enet_free (host);

       return NULL;
    }
    memset (host -> peers, 0, peerCount * sizeof (ENetPeer));

#if ENET_NETWORK_NEXT

    enet_list_clear( &host->receivePacketQueue );

    if ( config->client )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "creating network next client" );

        host->client = next_client_create( host, config->bind_address, enet_network_next_client_packet_received, NULL );

        if ( host->client == NULL )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "could not create network next client" );
            enet_free( host->peers );
            enet_free( host );
            return NULL;
        }

        host->address.host = 0x100007f;
        host->address.port = next_client_port( host->client );
    }
    else
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "creating network next server" );

        host->server = next_server_create( host, config->server_address, config->bind_address, config->server_datacenter, enet_network_next_server_packet_received, NULL );

        if ( !host->server )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "could not create network next server" );
            enet_free( host->peers );
            enet_free( host );
            return NULL;
        }

        struct next_address_t server_address = next_server_address( host->server );
        host->address = enet_address_from_next( &server_address );
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "enet host address is %x:%d", host->address.host, host->address.port );

#else // #if ENET_NETWORK_NEXT

    host -> socket = enet_socket_create (ENET_SOCKET_TYPE_DATAGRAM);
    if (host -> socket == ENET_SOCKET_NULL || (address != NULL && enet_socket_bind (host -> socket, address) < 0))
    {
       if (host -> socket != ENET_SOCKET_NULL)
         enet_socket_destroy (host -> socket);

       enet_free (host -> peers);
       enet_free (host);

       return NULL;
    }

    enet_socket_set_option (host -> socket, ENET_SOCKOPT_NONBLOCK, 1);
    enet_socket_set_option (host -> socket, ENET_SOCKOPT_BROADCAST, 1);
    enet_socket_set_option (host -> socket, ENET_SOCKOPT_RCVBUF, ENET_HOST_RECEIVE_BUFFER_SIZE);
    enet_socket_set_option (host -> socket, ENET_SOCKOPT_SNDBUF, ENET_HOST_SEND_BUFFER_SIZE);

    if (address != NULL && enet_socket_get_address (host -> socket, & host -> address) < 0)   
      host -> address = * address;

#endif // #if ENET_NETWORK_NEXT

    if (! channelLimit || channelLimit > ENET_PROTOCOL_MAXIMUM_CHANNEL_COUNT)
      channelLimit = ENET_PROTOCOL_MAXIMUM_CHANNEL_COUNT;
    else
    if (channelLimit < ENET_PROTOCOL_MINIMUM_CHANNEL_COUNT)
      channelLimit = ENET_PROTOCOL_MINIMUM_CHANNEL_COUNT;

    host -> randomSeed = (enet_uint32) (size_t) host;
    host -> randomSeed += enet_host_random_seed ();
    host -> randomSeed = (host -> randomSeed << 16) | (host -> randomSeed >> 16);
    host -> channelLimit = channelLimit;
    host -> incomingBandwidth = incomingBandwidth;
    host -> outgoingBandwidth = outgoingBandwidth;
    host -> bandwidthThrottleEpoch = 0;
    host -> recalculateBandwidthLimits = 0;
    host -> mtu = ENET_HOST_DEFAULT_MTU;
    host -> peerCount = peerCount;
    host -> commandCount = 0;
    host -> bufferCount = 0;
    host -> checksum = NULL;
    host -> receivedAddress.host = ENET_HOST_ANY;
    host -> receivedAddress.port = 0;
    host -> receivedData = NULL;
    host -> receivedDataLength = 0;
     
    host -> totalSentData = 0;
    host -> totalSentPackets = 0;
    host -> totalReceivedData = 0;
    host -> totalReceivedPackets = 0;

    host -> connectedPeers = 0;
    host -> bandwidthLimitedPeers = 0;
    host -> duplicatePeers = ENET_PROTOCOL_MAXIMUM_PEER_ID;
    host -> maximumPacketSize = ENET_HOST_DEFAULT_MAXIMUM_PACKET_SIZE;
    host -> maximumWaitingData = ENET_HOST_DEFAULT_MAXIMUM_WAITING_DATA;

    host -> compressor.context = NULL;
    host -> compressor.compress = NULL;
    host -> compressor.decompress = NULL;
    host -> compressor.destroy = NULL;

    host -> intercept = NULL;

    enet_list_clear (& host -> dispatchQueue);

    for (currentPeer = host -> peers;
         currentPeer < & host -> peers [host -> peerCount];
         ++ currentPeer)
    {
       currentPeer -> host = host;
       currentPeer -> incomingPeerID = currentPeer - host -> peers;
       currentPeer -> outgoingSessionID = currentPeer -> incomingSessionID = 0xFF;
       currentPeer -> data = NULL;

       enet_list_clear (& currentPeer -> acknowledgements);
       enet_list_clear (& currentPeer -> sentReliableCommands);
       enet_list_clear (& currentPeer -> sentUnreliableCommands);
       enet_list_clear (& currentPeer -> outgoingCommands);
       enet_list_clear (& currentPeer -> dispatchedCommands);

       enet_peer_reset (currentPeer);
    }

    return host;
}

/** Destroys the host and all resources associated with it.
    @param host pointer to the host to destroy
*/
void
enet_host_destroy (ENetHost * host)
{
    ENetPeer * currentPeer;

    if (host == NULL)
      return;

#if ENET_NETWORK_NEXT

    while ( !enet_list_empty( &host->receivePacketQueue ) )
    {
        struct ENetInternalPacket * packet = (struct ENetInternalPacket*) enet_list_remove( enet_list_begin( &host->receivePacketQueue ) );
        next_assert( packet );
        enet_free( packet );
    }

    enet_list_clear( &host->receivePacketQueue );

    if ( host->client )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "destroying network next client" );
        next_client_destroy( host->client );
        host->client = NULL;
    }

    if ( host->server )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "destroying network next server" );
        next_server_destroy( host->server );
        host->server = NULL;
    }

#else // #if ENET_NETWORK_NEXT

    enet_socket_destroy (host -> socket);

#endif // #if ENET_NETWORK_NEXT

    for (currentPeer = host -> peers;
         currentPeer < & host -> peers [host -> peerCount];
         ++ currentPeer)
    {
       enet_peer_reset (currentPeer);
    }

    if (host -> compressor.context != NULL && host -> compressor.destroy)
      (* host -> compressor.destroy) (host -> compressor.context);

    enet_free (host -> peers);
    enet_free (host);
}

enet_uint32
enet_host_random (ENetHost * host)
{
    /* Mulberry32 by Tommy Ettinger */
    enet_uint32 n = (host -> randomSeed += 0x6D2B79F5U);
    n = (n ^ (n >> 15)) * (n | 1U);
    n ^= n + (n ^ (n >> 7)) * (n | 61U);
    return n ^ (n >> 14);
}

/** Initiates a connection to a foreign host.
    @param host host seeking the connection
    @param address destination for the connection
    @param channelCount number of channels to allocate
    @param data user data supplied to the receiving host 
    @returns a peer representing the foreign host on success, NULL on failure
    @remarks The peer returned will have not completed the connection until enet_host_service()
    notifies of an ENET_EVENT_TYPE_CONNECT event for the peer.
*/
ENetPeer *
enet_host_connect (ENetHost * host, const ENetAddress * address, size_t channelCount, enet_uint32 data)
{
    ENetPeer * currentPeer;
    ENetChannel * channel;
    ENetProtocol command;

    if (channelCount < ENET_PROTOCOL_MINIMUM_CHANNEL_COUNT)
      channelCount = ENET_PROTOCOL_MINIMUM_CHANNEL_COUNT;
    else
    if (channelCount > ENET_PROTOCOL_MAXIMUM_CHANNEL_COUNT)
      channelCount = ENET_PROTOCOL_MAXIMUM_CHANNEL_COUNT;

    for (currentPeer = host -> peers;
         currentPeer < & host -> peers [host -> peerCount];
         ++ currentPeer)
    {
       if (currentPeer -> state == ENET_PEER_STATE_DISCONNECTED)
         break;
    }

    if (currentPeer >= & host -> peers [host -> peerCount])
      return NULL;

    currentPeer -> channels = (ENetChannel *) enet_malloc (channelCount * sizeof (ENetChannel));
    if (currentPeer -> channels == NULL)
      return NULL;
    currentPeer -> channelCount = channelCount;
    currentPeer -> state = ENET_PEER_STATE_CONNECTING;
    currentPeer -> address = * address;
    currentPeer -> connectID = enet_host_random (host);

    if (host -> outgoingBandwidth == 0)
      currentPeer -> windowSize = ENET_PROTOCOL_MAXIMUM_WINDOW_SIZE;
    else
      currentPeer -> windowSize = (host -> outgoingBandwidth /
                                    ENET_PEER_WINDOW_SIZE_SCALE) * 
                                      ENET_PROTOCOL_MINIMUM_WINDOW_SIZE;

    if (currentPeer -> windowSize < ENET_PROTOCOL_MINIMUM_WINDOW_SIZE)
      currentPeer -> windowSize = ENET_PROTOCOL_MINIMUM_WINDOW_SIZE;
    else
    if (currentPeer -> windowSize > ENET_PROTOCOL_MAXIMUM_WINDOW_SIZE)
      currentPeer -> windowSize = ENET_PROTOCOL_MAXIMUM_WINDOW_SIZE;
         
    for (channel = currentPeer -> channels;
         channel < & currentPeer -> channels [channelCount];
         ++ channel)
    {
        channel -> outgoingReliableSequenceNumber = 0;
        channel -> outgoingUnreliableSequenceNumber = 0;
        channel -> incomingReliableSequenceNumber = 0;
        channel -> incomingUnreliableSequenceNumber = 0;

        enet_list_clear (& channel -> incomingReliableCommands);
        enet_list_clear (& channel -> incomingUnreliableCommands);

        channel -> usedReliableWindows = 0;
        memset (channel -> reliableWindows, 0, sizeof (channel -> reliableWindows));
    }
        
    command.header.command = ENET_PROTOCOL_COMMAND_CONNECT | ENET_PROTOCOL_COMMAND_FLAG_ACKNOWLEDGE;
    command.header.channelID = 0xFF;
    command.connect.outgoingPeerID = ENET_HOST_TO_NET_16 (currentPeer -> incomingPeerID);
    command.connect.incomingSessionID = currentPeer -> incomingSessionID;
    command.connect.outgoingSessionID = currentPeer -> outgoingSessionID;
    command.connect.mtu = ENET_HOST_TO_NET_32 (currentPeer -> mtu);
    command.connect.windowSize = ENET_HOST_TO_NET_32 (currentPeer -> windowSize);
    command.connect.channelCount = ENET_HOST_TO_NET_32 (channelCount);
    command.connect.incomingBandwidth = ENET_HOST_TO_NET_32 (host -> incomingBandwidth);
    command.connect.outgoingBandwidth = ENET_HOST_TO_NET_32 (host -> outgoingBandwidth);
    command.connect.packetThrottleInterval = ENET_HOST_TO_NET_32 (currentPeer -> packetThrottleInterval);
    command.connect.packetThrottleAcceleration = ENET_HOST_TO_NET_32 (currentPeer -> packetThrottleAcceleration);
    command.connect.packetThrottleDeceleration = ENET_HOST_TO_NET_32 (currentPeer -> packetThrottleDeceleration);
    command.connect.connectID = currentPeer -> connectID;
    command.connect.data = ENET_HOST_TO_NET_32 (data);
 
    enet_peer_queue_outgoing_command (currentPeer, & command, NULL, 0, 0);

#if ENET_NETWORK_NEXT

    if ( host->client )
    {
        // puking noises... =p
        struct next_address_t server_address;
        enet_address_to_next( address, &server_address );
        char server_address_string[NEXT_MAX_ADDRESS_STRING_LENGTH];
        next_address_to_string( &server_address, server_address_string );
        next_client_open_session( host->client, server_address_string );
    }

#endif // #if ENET_NETWORK_NEXT

    return currentPeer;
}

/** Queues a packet to be sent to all peers associated with the host.
    @param host host on which to broadcast the packet
    @param channelID channel on which to broadcast
    @param packet packet to broadcast
*/
void
enet_host_broadcast (ENetHost * host, enet_uint8 channelID, ENetPacket * packet)
{
    ENetPeer * currentPeer;

    for (currentPeer = host -> peers;
         currentPeer < & host -> peers [host -> peerCount];
         ++ currentPeer)
    {
       if (currentPeer -> state != ENET_PEER_STATE_CONNECTED)
         continue;

       enet_peer_send (currentPeer, channelID, packet);
    }

    if (packet -> referenceCount == 0)
      enet_packet_destroy (packet);
}

/** Sets the packet compressor the host should use to compress and decompress packets.
    @param host host to enable or disable compression for
    @param compressor callbacks for for the packet compressor; if NULL, then compression is disabled
*/
void
enet_host_compress (ENetHost * host, const ENetCompressor * compressor)
{
    if (host -> compressor.context != NULL && host -> compressor.destroy)
      (* host -> compressor.destroy) (host -> compressor.context);

    if (compressor)
      host -> compressor = * compressor;
    else
      host -> compressor.context = NULL;
}

/** Limits the maximum allowed channels of future incoming connections.
    @param host host to limit
    @param channelLimit the maximum number of channels allowed; if 0, then this is equivalent to ENET_PROTOCOL_MAXIMUM_CHANNEL_COUNT
*/
void
enet_host_channel_limit (ENetHost * host, size_t channelLimit)
{
    if (! channelLimit || channelLimit > ENET_PROTOCOL_MAXIMUM_CHANNEL_COUNT)
      channelLimit = ENET_PROTOCOL_MAXIMUM_CHANNEL_COUNT;
    else
    if (channelLimit < ENET_PROTOCOL_MINIMUM_CHANNEL_COUNT)
      channelLimit = ENET_PROTOCOL_MINIMUM_CHANNEL_COUNT;

    host -> channelLimit = channelLimit;
}


/** Adjusts the bandwidth limits of a host.
    @param host host to adjust
    @param incomingBandwidth new incoming bandwidth
    @param outgoingBandwidth new outgoing bandwidth
    @remarks the incoming and outgoing bandwidth parameters are identical in function to those
    specified in enet_host_create().
*/
void
enet_host_bandwidth_limit (ENetHost * host, enet_uint32 incomingBandwidth, enet_uint32 outgoingBandwidth)
{
    host -> incomingBandwidth = incomingBandwidth;
    host -> outgoingBandwidth = outgoingBandwidth;
    host -> recalculateBandwidthLimits = 1;
}

void
enet_host_bandwidth_throttle (ENetHost * host)
{
    enet_uint32 timeCurrent = enet_time_get (),
           elapsedTime = timeCurrent - host -> bandwidthThrottleEpoch,
           peersRemaining = (enet_uint32) host -> connectedPeers,
           dataTotal = ~0,
           bandwidth = ~0,
           throttle = 0,
           bandwidthLimit = 0;
    int needsAdjustment = host -> bandwidthLimitedPeers > 0 ? 1 : 0;
    ENetPeer * peer;
    ENetProtocol command;

    if (elapsedTime < ENET_HOST_BANDWIDTH_THROTTLE_INTERVAL)
      return;

    host -> bandwidthThrottleEpoch = timeCurrent;

    if (peersRemaining == 0)
      return;

    if (host -> outgoingBandwidth != 0)
    {
        dataTotal = 0;
        bandwidth = (host -> outgoingBandwidth * elapsedTime) / 1000;

        for (peer = host -> peers;
             peer < & host -> peers [host -> peerCount];
            ++ peer)
        {
            if (peer -> state != ENET_PEER_STATE_CONNECTED && peer -> state != ENET_PEER_STATE_DISCONNECT_LATER)
              continue;

            dataTotal += peer -> outgoingDataTotal;
        }
    }

    while (peersRemaining > 0 && needsAdjustment != 0)
    {
        needsAdjustment = 0;
        
        if (dataTotal <= bandwidth)
          throttle = ENET_PEER_PACKET_THROTTLE_SCALE;
        else
          throttle = (bandwidth * ENET_PEER_PACKET_THROTTLE_SCALE) / dataTotal;

        for (peer = host -> peers;
             peer < & host -> peers [host -> peerCount];
             ++ peer)
        {
            enet_uint32 peerBandwidth;
            
            if ((peer -> state != ENET_PEER_STATE_CONNECTED && peer -> state != ENET_PEER_STATE_DISCONNECT_LATER) ||
                peer -> incomingBandwidth == 0 ||
                peer -> outgoingBandwidthThrottleEpoch == timeCurrent)
              continue;

            peerBandwidth = (peer -> incomingBandwidth * elapsedTime) / 1000;
            if ((throttle * peer -> outgoingDataTotal) / ENET_PEER_PACKET_THROTTLE_SCALE <= peerBandwidth)
              continue;

            peer -> packetThrottleLimit = (peerBandwidth * 
                                            ENET_PEER_PACKET_THROTTLE_SCALE) / peer -> outgoingDataTotal;
            
            if (peer -> packetThrottleLimit == 0)
              peer -> packetThrottleLimit = 1;
            
            if (peer -> packetThrottle > peer -> packetThrottleLimit)
              peer -> packetThrottle = peer -> packetThrottleLimit;

            peer -> outgoingBandwidthThrottleEpoch = timeCurrent;

            peer -> incomingDataTotal = 0;
            peer -> outgoingDataTotal = 0;

            needsAdjustment = 1;
            -- peersRemaining;
            bandwidth -= peerBandwidth;
            dataTotal -= peerBandwidth;
        }
    }

    if (peersRemaining > 0)
    {
        if (dataTotal <= bandwidth)
          throttle = ENET_PEER_PACKET_THROTTLE_SCALE;
        else
          throttle = (bandwidth * ENET_PEER_PACKET_THROTTLE_SCALE) / dataTotal;

        for (peer = host -> peers;
             peer < & host -> peers [host -> peerCount];
             ++ peer)
        {
            if ((peer -> state != ENET_PEER_STATE_CONNECTED && peer -> state != ENET_PEER_STATE_DISCONNECT_LATER) ||
                peer -> outgoingBandwidthThrottleEpoch == timeCurrent)
              continue;

            peer -> packetThrottleLimit = throttle;

            if (peer -> packetThrottle > peer -> packetThrottleLimit)
              peer -> packetThrottle = peer -> packetThrottleLimit;

            peer -> incomingDataTotal = 0;
            peer -> outgoingDataTotal = 0;
        }
    }

    if (host -> recalculateBandwidthLimits)
    {
       host -> recalculateBandwidthLimits = 0;

       peersRemaining = (enet_uint32) host -> connectedPeers;
       bandwidth = host -> incomingBandwidth;
       needsAdjustment = 1;

       if (bandwidth == 0)
         bandwidthLimit = 0;
       else
       while (peersRemaining > 0 && needsAdjustment != 0)
       {
           needsAdjustment = 0;
           bandwidthLimit = bandwidth / peersRemaining;

           for (peer = host -> peers;
                peer < & host -> peers [host -> peerCount];
                ++ peer)
           {
               if ((peer -> state != ENET_PEER_STATE_CONNECTED && peer -> state != ENET_PEER_STATE_DISCONNECT_LATER) ||
                   peer -> incomingBandwidthThrottleEpoch == timeCurrent)
                 continue;

               if (peer -> outgoingBandwidth > 0 &&
                   peer -> outgoingBandwidth >= bandwidthLimit)
                 continue;

               peer -> incomingBandwidthThrottleEpoch = timeCurrent;
 
               needsAdjustment = 1;
               -- peersRemaining;
               bandwidth -= peer -> outgoingBandwidth;
           }
       }

       for (peer = host -> peers;
            peer < & host -> peers [host -> peerCount];
            ++ peer)
       {
           if (peer -> state != ENET_PEER_STATE_CONNECTED && peer -> state != ENET_PEER_STATE_DISCONNECT_LATER)
             continue;

           command.header.command = ENET_PROTOCOL_COMMAND_BANDWIDTH_LIMIT | ENET_PROTOCOL_COMMAND_FLAG_ACKNOWLEDGE;
           command.header.channelID = 0xFF;
           command.bandwidthLimit.outgoingBandwidth = ENET_HOST_TO_NET_32 (host -> outgoingBandwidth);

           if (peer -> incomingBandwidthThrottleEpoch == timeCurrent)
             command.bandwidthLimit.incomingBandwidth = ENET_HOST_TO_NET_32 (peer -> outgoingBandwidth);
           else
             command.bandwidthLimit.incomingBandwidth = ENET_HOST_TO_NET_32 (bandwidthLimit);

           enet_peer_queue_outgoing_command (peer, & command, NULL, 0, 0);
       } 
    }
}
    
/** @} */
