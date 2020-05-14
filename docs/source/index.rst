
Network Next SDK
================

Introduction
------------

Network Next is the marketplace for premium network transit. 

Our technology monitors your player's network connections and runs bids on our marketplace to find routes across private networks with lower latency, packet loss and jitter than public networks.

When we find a route that meets your network optimization requirements, we steer player traffic across these private networks. Otherwise, players traverse the public network directly between the client and server.

If at any point Network Next is down, our SDK automatically falls back to providing unaccerelated direct connections between clients and servers.

Usage
-----

To use Network Next you must integrate our SDK with your game. 

The SDK consists of two main components. A client and a server. These components work together to monitor a player's connection and accelerate them when we find a route across Network Next that meets your requirements.

To get started, let's look at an example of creating a client and server:

Example
-------

.. code-block:: c++

	next_config_t config;
	next_default_config( &config );
	strncpy( config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1 );

.. toctree::
   :maxdepth: 2
   :caption: Contents:
