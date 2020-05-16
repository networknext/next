
environment
===========

You can fully configure Network Next SDK via environment variables.

Values set in the environment override values set by code.

NEXT_LOG_LEVEL
--------------

Set the network next log level.

Valid values:

 - **1** = NEXT_LOG_LEVEL_ERROR
 - **2** = NEXT_LOG_LEVEL_INFO (default)
 - **3** = NEXT_LOG_LEVEL_WARN
 - **4** = NEXT_LOG_LEVEL_DEBUG

Example:

.. code-block:: console

	$ set NEXT_LOG_LEVEL=4
	$ ./bin/simple_server
	(extreme spam follows...)