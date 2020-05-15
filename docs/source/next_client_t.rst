
next_client_t
=============

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

Curabitur pretium tincidunt lacus. Nulla gravida orci a odio. Nullam varius, turpis et commodo pharetra, est eros bibendum elit, nec luctus magna felis sollicitudin mauris. Integer in mauris eu nibh euismod gravida. Duis ac tellus et risus vulputate vehicula. Donec lobortis risus a elit. Etiam tempor. Ut ullamcorper, ligula eu tempor congue, eros est euismod turpis, id tincidunt sapien risus a quam. Maecenas fermentum consequat mi. Donec fermentum. Pellentesque malesuada nulla a mi. Duis sapien sem, aliquet nec, commodo eget, consequat quis, neque. Aliquam faucibus, elit ut dictum aliquet, felis nisl adipiscing sapien, sed malesuada diam lacus eget erat. Cras mollis scelerisque nunc. Nullam arcu. Aliquam consequat. Curabitur augue lorem, dapibus quis, laoreet et, pretium ac, nisi. Aenean magna nisl, mollis quis, molestie eu, feugiat in, orci. In hac habitasse platea dictumst.

(link to simple, upgraded, and complex client examples)

next_client_create
------------------

Creates an instance of a client, binding a socket to the specified address and port.

.. code-block:: c++

	next_client_t * next_client_create( void * context, 
	                                    const char * bind_address, 
	                                    void (*packet_received_callback)( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes ) );

**Parameters:**

	- **context** -- An optional pointer to context to be passed in any callbacks made from the client. Also passed to custom malloc and free functions called for allocations on behalf of this client instance.

	- **bind_address** -- An address string describing the bind address and port to bind to. Typically "0.0.0.0:0" is passed in, which binds to any IPv4 interface and lets the system pick a port. Alternatively, you can bind to a specific port "0.0.0.0:50000".

	- **packet_received_callback** -- Called from the same thread that calls *next_client_update*, whenever a packet is received from the server. Required.

**Return value:** 

	The newly created client instance, or NULL, if the client could not be created. 

	Typically, NULL is only returned when another socket is already bound on the same port, or if an invalid bind address is passed in.

**Example:**

First define a callback for received packets:

.. code-block:: c++

	void client_packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
	{
	    printf( "client received packet from server (%d bytes)\n", packet_bytes );
	}

Then, create a client:

.. code-block:: c++

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );
    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

next_client_destroy
-------------------

Destroys a client instance, and the socket it manages internally.

.. code-block:: c++

	void next_client_destroy( next_client_t * client );

**Parameters:**

	- **client** -- The client instance to destroy. Must be a valid client instance created by *next_client_create*. Do not pass in NULL.

**Return value:** 

	The newly created client instance, or NULL, if the client could not be created. 

**Example:**

.. code-block:: c++

	next_client_destroy( client );

next_client_port
----------------

Gets the port the client socket is bound to.

.. code-block:: c++

	uint16_t next_client_port( next_client_t * client );

**Return value:** 

	The actual port number the client socket is bound to.

	This makes it possible to look up what specific port the client is bound to when you bind to port zero and the system chooses a port.

**Example:**

.. code-block:: c++

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );
    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

    const uint16_t client_port = next_client_port( client );

    printf( "the client was bound to port %d\n", client_port );

next_client_open_session
------------------------

Opens a session between the client and a server.

.. code-block:: c++

	void next_client_open_session( next_client_t * client,
	                               const char * server_address )

**Parameters:**

	- **client** -- The client instance.

	- **server_address** -- The address of the server that the client wants to connect to.

**Example:**

.. code-block:: c++

	next_client_open_session( client, "127.0.0.1:50000" );

next_client_close_session
-------------------------

Closes the session between the client and server.

.. code-block:: c++

	void next_client_close_session( next_client_t * client )

**Parameters:**

	- **client** -- The client instance.

**Example:**

.. code-block:: c++

	next_client_close_session( client );

next_client_is_session_open
---------------------------

Check if the client has a session open.

.. code-block:: c++

	bool next_client_is_session_open( next_client_t * client );

**Return value:** 

	True, if the client has an open session with a server, false otherwise.

**Example:**

.. code-block:: c++

    const bool session_open = next_client_session_open( client );

    printf( "session open = %s\n", session_open ? "yes" : "no" );

next_client_state
-----------------

Gets the state the client is in.

.. code-block:: c++

	int next_client_state( next_client_t * client );

**Return value:** 

	The client state is either:

		- NEXT_CLIENT_STATE_CLOSED
		- NEXT_CLIENT_STATE_OPEN
		- NEXT_CLIENT_STATE_ERROR

	The client is initially in closed state. After *next_client_open_session* the client is immediately in open state on success, or error state if something went wrong while opening the session, for example, an invalid server address was passed in.

**Example:**

.. code-block:: c++

    const char * state = "???";

    const int client_state = next_client_state( client );
    
    switch ( client_state )
    {
        case NEXT_CLIENT_STATE_CLOSED:
            state = "closed";
            break;

        case NEXT_CLIENT_STATE_OPEN:
            state = "open";
            break;

        case NEXT_CLIENT_STATE_ERROR:
            state = "error";
            break;

        default:
            break;
    }

    printf( "client state = %s (%d)\n", state, client_state );

next_client_update
------------------

Updates the client.

.. code-block:: c++

	int next_client_update( next_client_t * client );

Please call this every frame as it drives the packet received callback.

**Example:**

.. code-block:: c++

    while ( !quit )
    {
        next_client_update( client );

        // ... do stuff ...
        
        next_sleep( 1.0 / 60.0 );
    }

next_client_send_packet
-----------------------

Sends a packet to the server.

.. code-block:: c++

	void next_client_send_packet( next_client_t * client, const uint8_t * packet_data, int packet_bytes );

Depending on whether this player is accelerated or not, this packet will be sent direct across the public internet, or through Network Next's network of private networks.

**Example:**

.. code-block:: c++

    uint8_t packet_data[32];
    memset( packet_data, 0, sizeof( packet_data ) );

    while ( !quit )
    {
        next_client_update( client );

        next_client_send_packet( client, packet_data, sizeof(packet_data) );
        
        next_sleep( 1.0 / 60.0 );
    }

next_client_send_packet_direct
------------------------------

Sends a packet to the server, forcing the packet to be sent across the public internet.

.. code-block:: c++

	void next_client_send_packet_direct( next_client_t * client, const uint8_t * packet_data, int packet_bytes );

The packet will be sent unaccelerated across the public internet and will not count towards your Network Next bandwidth envelope.

This can be very useful when you need to send a burst of non-latency sensitive packets, for example, in a load screen.

Example:

.. code-block:: c++

    uint8_t packet_data[32];
    memset( packet_data, 0, sizeof( packet_data ) );

    while ( !quit )
    {
        next_client_update( client );

        next_client_send_packet_direct( client, packet_data, sizeof(packet_data) );
        
        next_sleep( 1.0 / 60.0 );
    }

next_client_flag_session
------------------------

Flag the session as problematic.

.. code-block:: c++

	void next_client_flag_session( next_client_t * client );

This feature was added to support our customers who let players flag bad play sessions. 

Call this function when your players complain, and it's sent to our backend so we can help you track down why!

**Example:**

.. code-block:: c++

    next_client_report_session( client );

next_client_session_id
----------------------

Gets the client session id.

.. code-block:: c++

	void next_client_flag_session( next_client_t * client );

The client session id is a random uint64_t generated by our backend per-unique session.

It is assigned only once the server has upgraded a player, so until that point, the client will have a session id of 0.

**Example:**

.. code-block:: c++

    const uint64_t session_id = next_client_session_id( client );

    printf( "session id = %" PRIx64 "\n", session_id );

next_client_stats
-----------------

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

next_client_set_user_flags
--------------------------

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
