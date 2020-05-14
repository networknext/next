
Network Next SDK
================

Introduction
------------

Network Next is the marketplace for premium network transit. We steer player traffic across private network with lower latency, packet loss and jitter than public networks.

To use Network Next you must integrate this SDK with your game. The SDK consists of two main components. A client and a server. These components work together to monitor a player's connection and accelerate them when we find a route across Network Next that meets your optimization requirements.

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
