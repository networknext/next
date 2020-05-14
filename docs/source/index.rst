
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

The SDK consists of two main components. A client and a server. These components work together to monitor a player's connection and accelerate it when we find a route across Network Next.

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

.. code-block:: c++

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );
    if ( client == NULL )
    {
	    printf( "error: failed to create client\n" );
	    return 1;
    }

Internally, the client creates and manages a UDP socket. In this example the socket is bound to IPv4 address 0.0.0.0 (any interface), and port 0, letting the system choose the port number.

Now, open a session between the client and the server:

.. code-block:: c++

	next_client_open_session( client, "127.0.0.0:50000" );

In this example, the server is running on 127.0.0.0 (localhost) on port 50000.

You can send packets from the client to the server like this:

.. code-block:: c++

	uint8_t packet_data[32];
	memset( packet_data, 0, sizeof( packet_data ) );
	next_client_send_packet( client, packet_data, sizeof(packet_data) );

Now make sure the client is updated once every time your game frame updates:

.. code-block:: c++

	next_client_update_session( client );

This is necessary to keep the main thread interface for the client in sync with the background thread that does all the work.

Finally, when you have finished your session with the server, you close it:

.. code-block:: c++

	next_client_close_session( client );

From this point you can open another session, or you can destroy the client:

.. code-block:: c++

	next_client_destroy( client );

Finally, before your client application terminates, please shut down the network next SDK:

.. code-block:: c++

	next_term();

.. toctree::
   :maxdepth: 2
   :caption: Contents:
