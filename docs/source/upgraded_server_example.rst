
Upgraded Server Example
-----------------------

In this example we setup a server for monitoring and acceleration by network next.

First define configuration values for the server:

.. code-block:: c++

	const char * bind_address = "0.0.0.0:50000";
	const char * server_address = "127.0.0.1:50000";
	const char * server_datacenter = "local";
	const char * backend_hostname = "dev.networknext.com";
	const char * customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn";

This includes the test customer private key we're using in this example. A customer private key is required on the server to enable acceleration by Network Next.

Next, initialize a configuration struct to defaults, then copy the hostname and the customer private key on top.

.. code-block:: c++

    next_config_t config;
    next_default_config( &config );
    strncpy( config.hostname, backend_hostname, sizeof(config.hostname) - 1 );
    strncpy( config.customer_private_key, customer_private_key, sizeof(config.customer_private_key) - 1 );

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

Next, define a function to be called when packets are received from clients.

Here is one that reflects the packet back to the client that sent it, and upgrades the client that sent the packet for monitoring and acceleration by Network Next.

.. code-block:: c++

	next_server_send_packet( server, from, packet_data, packet_bytes );

	next_printf( NEXT_LOG_LEVEL_INFO, "server received packet from client (%d bytes)", packet_bytes );

	if ( !next_server_session_upgraded( server, from ) )
	{
	    const char * user_id_string = "12345";
	    next_server_upgrade_session( server, from, user_id_string, NULL );
	}

Generally you would *not* want to upgrade every client session you receive a packet from. This is just done to make this example easy to implement.

Instead, you should only upgrade sessions that have passed whatever security and protocol level checks you have in your game so you are 100% confident this is a real player joining your game.

This is important because you are paying to monitor and accelerate any players you upgrade.

Also notice that you can pass in a user id as a string to the upgrade call.

.. code-block:: c++

	next_server_upgrade_session( server, from, user_id_string, NULL );

This user id is very important because it allows you to look up users by that id in our portal. This user id is hashed before sending to our backend for privacy reasons.

Now, create the server. 

.. code-block:: c++

    next_server_t * server = next_server_create( NULL, server_address, bind_address, server_datacenter, server_packet_received );
    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }

Make sure the server gets updated every frame:

.. code-block:: c++

	next_server_update( server );

When you have finished using your server, please destroy it:

.. code-block:: c++

	next_server_destroy( server );

Before your application terminates, please shut down the SDK:

.. code-block:: c++

	next_term();
