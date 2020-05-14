
Network Next SDK
================

Introduction
------------

Network Next is the marketplace for premium network transit. 

Our technology monitors your player's network connections and runs bids on our marketplace to find routes across private networks with better performance than public networks.

When we find a route that meets your network optimization requirements, we steer player traffic across these private networks. Otherwise, players traverse the public network directly between the client and server.

If at any point Network Next is down, our SDK falls back to the public network, without any disruption to your players.

Usage
-----

To use Network Next you must integrate our SDK with your game. 

The SDK consists of two main components. A client and a server. These components work together to monitor a player's connection and accelerate them when we find a route across Network Next.

To get started, let's look at an example of creating a client:

First, initialize the SDK:

.. code-black:: c++

	next_init( NULL, NULL ); 

Next, define a function to be called when packets are received:

.. code-block:: c++

	void client_packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
	{
		printf( "client received packet (%d bytes)", packet_bytes );
	}

Now create a network next client:

-- code-block:: c++

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );
    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

In this example we bound to IPv4 any address 0.0.0.0, with a port number of zero which makes the system select the port number to use for your client socket.

.. toctree::
   :maxdepth: 2
   :caption: Contents:
