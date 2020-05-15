
Complex Client Example
----------------------

In this example we build the kitchen sink version of a client where we show off all the features.

We demonstrate:

- Setting the network next log level
- Setting a custom log function
- Setting a custom assert handler
- Setting a custom allocator
- Querying the port the client socket is bound to running on
- Getting statistics from the client and displaying them periodically

This is going to be a huge example, so let's get started!

First, as with the upgraded client example, we start by defining our key configuration variables:

.. code-block:: c++

	const char * bind_address = "0.0.0.0:0";
	const char * server_address = "127.0.0.1:50000";
	const char * customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";

Next, we dive right in and define a custom allocator class that tracks all allocations made, and checks them for leaks when it shuts down:

.. code-block:: c++

	struct AllocatorEntry
	{
	    // ...
	};

	class Allocator
	{
	    int64_t num_allocations;
	    next_mutex_t mutex;
	    std::map<void*, AllocatorEntry*> entries;

	public:

	    Allocator()
	    {
	        int result = next_mutex_create( &mutex );
	        next_assert( result == NEXT_OK );
	        num_allocations = 0;
	    }

	    ~Allocator()
	    {
	        next_mutex_destroy( &mutex );
	        next_assert( num_allocations == 0 );
	        next_assert( entries.size() == 0 );
	    }

	    void * Alloc( size_t size )
	    {
	        next_mutex_guard( &mutex );
	        void * pointer = malloc( size );
	        next_assert( pointer );
	        next_assert( entries[pointer] == NULL );
	        AllocatorEntry * entry = new AllocatorEntry();
	        entries[pointer] = entry;
	        num_allocations++;
	        return pointer;
	    }

	    void Free( void * pointer )
	    {
	        next_mutex_guard( &mutex );
	        next_assert( pointer );
	        next_assert( num_allocations > 0 );
	        std::map<void*, AllocatorEntry*>::iterator itor = entries.find( pointer );
	        next_assert( itor != entries.end() );
	        entries.erase( itor );
	        num_allocations--;
	        free( pointer );
	    }
	};

IMPORTANT: Since this allocator will be called from multiple threads, it must be thread safe. This is done by using the platform independent mutex supplied by the Network Next SDK.

There are three types of allocations done by the Network Next SDK:

1. Global allocations
2. Per-client allocations
3. Per-server allocations

Each of these situations corresponds to what is called a "context" in the Network Next SDK. 

A context is simply a void* to a type that you define which is passed in to malloc and free callbacks that we call to perform allocations on behalf of the SDK. The context passed is gives you the flexibility to have a specific memory pool for Network Next (most common), or even to have a completely different allocation pool for each client and server instance. That's what we're going to do in this example.

Let's define a base context that will be used for global allocations:

.. code-block:: c++

	struct Context
	{
	    Allocator * allocator;
	};

And a per-client context that is binary compatible with the base context, to be used for per-client allocations:

.. code-block:: c++

	struct ClientContext
	{
	    Allocator * allocator;
	    uint32_t client_data;
	};

As you can see, the client context can contain additional information aside from the allocator. The context is not *just* passed into allocator callbacks, but all callbacks from the client and server, so you can use it to integrate with your own client and server objects in your game. 

Here we just set a dummy uint32_t and check that the value to verify it's being passed through correctly. For example, in the received packet callback, we have access to the client context and check the value is what we expect:

.. code-block:: c++

	void client_packet_received( next_client_t * client, void * _context, const uint8_t * packet_data, int packet_bytes )
	{
	    (void) client;

	    ClientContext * context = (ClientContext*) _context;

	    next_assert( context );
	    next_assert( context->allocator != NULL );
	    next_assert( context->client_data == 0x12345 );

	    next_printf( NEXT_LOG_LEVEL_INFO, "client received packet from server (%d bytes)", packet_bytes );

	    verify_packet( packet_data, packet_bytes );
	}

Next we define malloc and free functions to pass in to the SDK. These same functions are used for global, per-client and per-server allocations. The only difference is the context passed in to each.

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

