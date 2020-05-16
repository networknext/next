
Reference
---------

To use the Network Next SDK, you replace the socket on your client with *next_client_t* and the socket on your server with *next_server_t*.

Together the client and server provide an interface very similar to sendto and recvfrom, except that internally, they monitor the network performonce for your player, and automatically send your packets packets across Network Next when we find a route that meets your network optimization requirements.

If anything goes wrong, the client and server automatically fall back to sending packets across the public internet, without causing any disruption for your player.

.. toctree::
   :maxdepth: 1

   next_client_t
   next_server_t
   next_address_t
   globals