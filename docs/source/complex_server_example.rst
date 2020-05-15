
Complex Server Example
----------------------

In this example we build the kitchen sink version of a server where we show off all the features.

We demonstrate:

- Setting the network next log level
- Setting a custom log function
- Setting a custom assert handler
- Setting a custom allocator

In this example, everything is the same as per the complex client example, specifically setting up the allocator, a global context, override functions for malloc and free, custom log function, custom assert function as in the previous example.

Now when creating a server, we create it with a server context as follows:

.. code-block:: c++

	Allocator server_allocator;
	ServerContext server_context;
	server_context.allocator = &server_allocator;
	server_context.server_data = 0x12345678;

	next_server_t * server = next_server_create( &server_context, server_address, bind_address, server_datacenter, server_packet_received );
	if ( server == NULL )
	{
	    printf( "error: failed to create server\n" );
	    return 1;
	}

And now this calls the overridden malloc and free functions with the server context:

.. code-block:: c++

	void * malloc_function( void * _context, size_t bytes )
	{
	    Context * context = (Context*) _context;
	    next_assert( context );
	    next_assert( context->allocator );
	    return context->allocator->Alloc( bytes );
	}

	void free_function( void * _context, void * p )
	{
	    Context * context = (Context*) _context;
	    next_assert( context );
	    next_assert( context->allocator );
	    return context->allocator->Free( p );
	}

more...