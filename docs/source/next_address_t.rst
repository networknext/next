
next_address_t
==============

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

Curabitur pretium tincidunt lacus. Nulla gravida orci a odio. Nullam varius, turpis et commodo pharetra, est eros bibendum elit, nec luctus magna felis sollicitudin mauris. Integer in mauris eu nibh euismod gravida. Duis ac tellus et risus vulputate vehicula. Donec lobortis risus a elit. Etiam tempor. Ut ullamcorper, ligula eu tempor congue, eros est euismod turpis, id tincidunt sapien risus a quam. Maecenas fermentum consequat mi. Donec fermentum. Pellentesque malesuada nulla a mi. Duis sapien sem, aliquet nec, commodo eget, consequat quis, neque. Aliquam faucibus, elit ut dictum aliquet, felis nisl adipiscing sapien, sed malesuada diam lacus eget erat. Cras mollis scelerisque nunc. Nullam arcu. Aliquam consequat. Curabitur augue lorem, dapibus quis, laoreet et, pretium ac, nisi. Aenean magna nisl, mollis quis, molestie eu, feugiat in, orci. In hac habitasse platea dictumst.

next_client_create
------------------

Creates an instance of a client, binding a socket to the specified address and port.

.. code-block:: c++

	next_client_t * next_client_create( void * context, 
	                                    const char * bind_address, 
	                                    void (*packet_received_callback)( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes ) );

**Parameters:**

	- **context** -- An optional pointer to context passed to any callbacks made from the client.

	- **bind_address** -- An address string describing the bind address and port to bind to. Typically "0.0.0.0:0" is passed in, which binds to any IPv4 interface and lets the system pick a port. Alternatively, you can bind to a specific port, for example: "0.0.0.0:50000".

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

**Example:**

.. code-block:: c++

	next_client_destroy( client );

next_client_port
----------------

Gets the port the client socket is bound to.

.. code-block:: c++

	uint16_t next_client_port( next_client_t * client );

**Return value:** 

	The port number the client socket is bound to.

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
