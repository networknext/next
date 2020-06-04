
next_server_t
=============

The server side of a client/server connection.

To use a server, create it and it automatically starts accepting sessions from clients.

To upgrade a session for monitoring and *potential* acceleration, call *next_server_upgrade_session* on a client's address.

Packets received from clients are passed to you via a callback function, including the address of the client that sent the packet.

Make sure to pump the server update once per frame.

**Examples:**

-   :doc:`simple_server_example`
-   :doc:`upgraded_server_example`
-   :doc:`complex_server_example`

next_server_create
------------------

Creates an instance of a server, binding a socket to the specified address and port.

.. code-block:: c++

	next_server_t * next_server_create( void * context, 
	                                    const char * server_address, 
	                                    const char * bind_address, 
	                                    const char * datacenter, 
	                                    void (*packet_received_callback)( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes ) );

**Parameters:**

	- **context** -- An optional pointer to context passed to any callbacks made from the server.

	- **server_address** -- The public IP address and port that clients will connect to.

	- **bind_address** -- The address and port to bind to. Typically "0.0.0.0:[portnum]" is passed in, binding the server socket to any IPv4 interface on a specific port, for example: "0.0.0.0:50000".

	- **datacenter** -- The name of the datacenter that the game server is running in. Please pass in "local" until we work with you to determine the set of datacenters you host servers in.

	- **packet_received_callback** -- Called from the same thread that calls *next_server_update*, whenever a packet is received from a client. Required.

	- **wake_up_callback** -- Optional callback. Pass NULL if not used. Sets a callback function to be called from an internal network next thread when a packet is ready to be received for this server. Intended to let you set an event of your own creation when a packet is ready to receive, making it possible to use Network Next with applications built around traditional select or wait for multiple event style blocking socket loops. Call next_server_update to pump received packets to the packet_received_callback when you wake up on main thread from your event.

**Return value:** 

	The newly created server instance, or NULL, if the server could not be created. 

	Typically, NULL is only returned when another socket is already bound on the same port, or if an invalid server or bind address is passed in.

**Example:**

First define a callback for received packets:

.. code-block:: c++

	void server_packet_received( next_server_t * server, void * _context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
	{
	    char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
	    next_printf( NEXT_LOG_LEVEL_INFO, "server received packet from client %s (%d bytes)", next_address_to_string( from, address_buffer ), packet_bytes );
	}

Then, create a server:

.. code-block:: c++

    next_server_t * server = next_server_create( NULL, "0.0.0.0:0", server_packet_received );
    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }

next_server_destroy
-------------------

Destroys a server instance, and the socket it manages internally.

.. code-block:: c++

	void next_server_destroy( next_server_t * server );

**Parameters:**

	- **server** -- The server instance to destroy. Must be a valid server instance created by *next_server_create*. Do not pass in NULL.

**Example:**

.. code-block:: c++

	next_server_destroy( server );

next_server_port
----------------

Gets the port the server socket is bound to.

.. code-block:: c++

	uint16_t next_server_port( next_server_t * server );

**Return value:** 

	The port number the server socket is bound to.

**Example:**

.. code-block:: c++

    next_server_t * server = next_server_create( NULL, "127.0.0.1", "0.0.0.0:50000", "local", server_packet_received );
    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }

    const uint16_t server_port = next_server_port( client );

    printf( "the client is bound to port %d\n", server_port );

next_server_state
-----------------

Gets the state the server is in.

.. code-block:: c++

	int next_server_state( next_server_t * server );

**Parameters:**

	- **server** -- The server instance.

**Return value:** 

	The server state, which is one of the following:
	
		- NEXT_SERVER_STATE_DIRECT_ONLY
		- NEXT_SERVER_STATE_RESOLVING_HOSTNAME
		- NEXT_SERVER_STATE_INITIALIZING
		- NEXT_SERVER_STATE_INITIALIZED

	The server is initially in the direct only state. 

	If a valid customer private key is setup, the server will first try to resolve the backend hostname, which is "prod.networknext.com" by default.

	Once the backend hostname is resolved, the server initializes with the backend. When everything works, the server lands in the initialized state and is ready to accelerate players.

	If anything fails, the server falls back to the direct only state, and only serves up direct routes over the public internet.

**Example:**

.. code-block:: c++

    const char * state = "???";

    const int server_state = next_server_state( server );
    
    switch ( server_state )
    {
        case NEXT_SERVER_STATE_DIRECT_ONLY:
            state = "direct only";
            break;

        case NEXT_SERVER_STATE_RESOLVING_HOSTNAME:
            state = "resolving hostname";
            break;

        case NEXT_SERVER_STATE_INITIALIZING:
            state = "initializing";
            break;

        case NEXT_SERVER_STATE_INITIALIZED:
            state = "initialized";
            break;

        default:
            break;
    }

    printf( "server state = %s (%d)\n", state, server_state );

next_server_upgrade_session
---------------------------

Upgrades a session for monitoring and *potential* acceleration by Network Next.

.. code-block:: c++

	uint64_t next_server_upgrade_session( next_server_t * server, 
	                                      const next_address_t * address, 
	                                      const char * user_id );

IMPORTANT: Make sure you only call this function when you are 100% sure this is a real player in your game.

**Parameters:**

	- **server** -- The server instance.

	- **address** -- The address of the client to be upgraded.

	- **user_id** -- The user id for the session. Pass in any unique per-user identifier you have.

**Return value:**

	The session id assigned the session that was upgraded.

**Example:**

The address struct is defined as follows:

.. code-block:: c++

	struct next_address_t
	{
	    union { uint8_t ipv4[4]; uint16_t ipv6[8]; } data;
	    uint16_t port;
	    uint8_t type;
	};

You can parse an address from a string like this:

.. code-block:: c++

	next_address_t address;
	if ( next_address_parse( &address, "127.0.0.1:50000" ) != NEXT_OK )
	{
	    printf( "error: failed to parse address\n" );
	}

The address struct is passed in when you receive packet from a client:

.. code-block:: c++

	void server_packet_received( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
	{
	    char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
	    next_printf( NEXT_LOG_LEVEL_INFO, "server received packet from client %s (%d bytes)", next_address_to_string( from, address_buffer ), packet_bytes );
	}

Once you have the address, upgrading a session is easy:

.. code-block:: c++

	next_server_upgrade_session( server, client_address, user_id );

next_server_tag_session
-----------------------

Tags a session for potentially different network optimization parameters.

.. code-block:: c++

	void next_server_tag_session( next_server_t * server, const next_address_t * address, const char * tag );

**Parameters:**

	- **server** -- The server instance.

	- **address** -- The address of the client to tag.

	- **tag** -- The tag to be applied to the client. Some ideas: "pro", "streamer" or "dev".

**Example:**

.. code-block:: c++

	next_server_tag_session( server, client_address, "pro" );

next_server_session_upgraded
----------------------------

Checks if a session has been upgraded.

.. code-block:: c++

	bool next_server_session_upgraded( next_server_t * server, const next_address_t * address );

**Parameters:**

	- **server** -- The server instance.

	- **address** -- The address of the client to check.

**Return value:**

	True if the session has been upgraded, false otherwise.

**Example:**

.. code-block:: c++

	const bool upgraded = next_server_session_upgraded( server, client_address );

	printf( "session upgraded = %s\n", upgraded ? "true" : "false" );

next_server_send_packet
-----------------------

Send a packet to a client.

.. code-block:: c++

	void next_server_send_packet( next_server_t * server, const next_address_t * to_address, const uint8_t * packet_data, int packet_bytes );

Sends a packet to a client. If the client is upgraded and accelerated by network next, the packet will be sent across our private network of networks.

Otherwise, the packet will be sent across the public internet.

**Parameters:**

	- **server** -- The server instance.

	- **to_address** -- The address of the client to send the packet to.

	- **packet_data** -- The packet data to send.

	- **packet_bytes** -- The size of the packet. Must be in the range 1 to NEXT_MTU (1300).

**Example:**

.. code-block:: c++

	uint8_t packet_data[32];
	memset( packet_data, 0, sizeof(packet_data) );
	next_server_send_packet( server, client_address, packet_data, sizeof(packet_data) );

next_server_send_packet_direct
------------------------------

Send a packet to a client, forcing the packet to be sent over the public internet.

.. code-block:: c++

	void next_server_send_packet_direct( next_server_t * server, const next_address_t * to_address, const uint8_t * packet_data, int packet_bytes );

This function is useful when you need to send non-latency sensitive packets to the client, for example, during a load screen.

Packets sent via this function do not apply to your network next bandwidth envelope.

**Parameters:**

	- **server** -- The server instance.

	- **to_address** -- The address of the client to send the packet to.

	- **packet_data** -- The packet data to send.

	- **packet_bytes** -- The size of the packet. Must be in the range 1 to NEXT_MTU (1300).

**Example:**

.. code-block:: c++

	uint8_t packet_data[32];
	memset( packet_data, 0, sizeof(packet_data) );
	next_server_send_packet_direct( server, client_address, packet_data, sizeof(packet_data) );
