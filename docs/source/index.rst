
Network Next SDK
================


Introduction
------------

Network Next is the marketplace for premium network transit. 

Our technology monitors your player's network connections and runs bids on our marketplace to find routes across private networks with better performance than the public internet.

If at any point Network Next is down, our SDK falls back to the public internet, without any disruption to your players.


How it works
------------

The SDK has two main components. A client and a server. 

(diagrams etc.)

These components work together to monitor a player's connection and accelerate it when we find a route across Network Next.

(etc etc...)


Simple Client Example
---------------------

First, initialize the SDK:

.. code-block:: c++

	if ( next_init( NULL, NULL ) != NEXT_OK )
	{
	    printf( "error: could not initialize network next\n" );
	    return 1;
	}

Next, define a function to be called when packets are received:

.. code-block:: c++

	void client_packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
	{
	    printf( "client received packet (%d bytes)", packet_bytes );
	}

Create the client:

.. code-block:: c++

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );
    if ( client == NULL )
    {
	    printf( "error: failed to create client\n" );
	    return 1;
    }

Open a session between the client and the server:

.. code-block:: c++

	next_client_open_session( client, "127.0.0.0:50000" );

Now you can send packets to the server like this:

.. code-block:: c++

	uint8_t packet_data[32];
	memset( packet_data, 0, sizeof( packet_data ) );
	next_client_send_packet( client, packet_data, sizeof(packet_data) );

Make sure the client is updated once every frame:

.. code-block:: c++

	next_client_update( client );

When you have finished your session with the server, close it:

.. code-block:: c++

	next_client_close_session( client );

When you have finished using your client, destroy it:

.. code-block:: c++

	next_client_destroy( client );

Before your application terminates, shut down the SDK:

.. code-block:: c++

	next_term();


Simple Server Example
---------------------

Initialize the SDK:

.. code-block:: c++

	if ( next_init( NULL, NULL ) != NEXT_OK )
	{
	    printf( "error: could not initialize network next\n" );
	    return 1;
	}

Next, define a function to be called when packets are received. 

Here is one that reflects the packet back to the client that sent it:

.. code-block:: c++

	void server_packet_received( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
	{
	    next_server_send_packet( server, from, packet_data, packet_bytes );
	}

Now create the server. In this example, we bind the server to port 50000 on 127.0.0.1 IPv4 address (localhost) and set the datacenter where your server is running to "local":

.. code-block:: c++

    next_server_t * server = next_server_create( NULL, "127.0.0.1:50000", "0.0.0.0:50000", "local", server_packet_received );
    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }

Make sure the server gets updated every frame:

.. code-block:: c++

	next_server_update( server );

When you have finished using your server, destroy it:

.. code-block:: c++

	next_server_destroy( server );

Before your application terminates, shut down the SDK:

.. code-block:: c++

	next_term();

.. toctree::
   :maxdepth: 2
   :caption: Contents:
