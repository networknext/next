
globals
=======

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

next_config_t
-------------

.. code-block:: c++

	struct next_config_t
	{
	    char hostname[256];
	    char customer_public_key[256];
	    char customer_private_key[256];
	    int socket_send_buffer_size;
	    int socket_receive_buffer_size;
	    bool disable_network_next;
	    bool disable_tagging;
	};

next_default_config
-------------------

next_init
---------

next_term
---------

next_time
---------

next_sleep
----------

next_printf
-----------

next_assert
-----------

next_quiet
----------

next_log_level
--------------

next_log_function
-----------------

next_assert_function
--------------------

next_allocator
--------------

next_user_id_string
-------------------

next_is_network_next_packet
---------------------------

