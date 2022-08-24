
Upgraded Client Example
-----------------------

In this example we setup a client for monitoring and acceleration by network next.

First define configuration values for the client:

.. code-block:: c++

	const char * bind_address = "0.0.0.0:0";
	const char * server_address = "127.0.0.1:50000";
	const char * customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";

These includes the bind address for the client socket, the server address to connect to, and the test customer public key we're using in this example. A customer public key is required to enable acceleration by Network Next.

Next, initialize a configuration struct to defaults, then copy the customer public key on top.

.. code-block:: c++

	next_config_t config;
	next_default_config( &config );
	strncpy( config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1 );

	if ( next_init( NULL, &config ) != NEXT_OK )
	{
	    printf( "error: could not initialize network next\n" );
	    return 1;
	}

Initialize the SDK, this time passing in the configuration struct. 

.. code-block:: c++

	if ( next_init( NULL, &config ) != NEXT_OK )
	{
	    printf( "error: could not initialize network next\n" );
	    return 1;
	}

This activates the customer public key so it's used by the client. 

Network Next needs a customer public key to monitor and accelerate players. Without a customer public key, Network Next just sends player traffic across the public internet.

Next, define a function to be called when packets are received:

.. code-block:: c++

	void client_packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
	{
	    printf( "client received packet (%d bytes)\n", packet_bytes );
	}

Create the client:

.. code-block:: c++

	next_client_t * client = next_client_create( NULL, bind_address, client_packet_received, NULL );
	if ( client == NULL )
	{
	    printf( "error: failed to create client\n" );
	    return 1;
	}

Open a session between the client and the server:

.. code-block:: c++

	next_client_open_session( client, server_address );

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
