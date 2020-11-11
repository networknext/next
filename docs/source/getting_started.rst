
Getting Started
===============

1. Sign up and download the SDK
-------------------------------

Sign up and create an account at https://portal.networknext.com by clicking "Get Access".

Once you've signed up, a "Downloads" page will show up. Here you can download the latest SDK.

Download the SDK and unzip it. 

2. Run the keygen
-----------------

There is a keygen tool under the "keygen" directory.

Run the keygen to generate a keypair for your company. 

You'll see something like this:

.. code-block:: console

	Welcome to Network Next!

	This is your public key:

	    OGivr2IM0k7oLTQ3lmGXVnZJpxDRPFsZrKxYLn7fQAosTpQAfs464w==

	This is your private key:

	    OGivr2IM0k4lCfbM/VZCVK99KkDSCbzi8fzM2WnZCQb7R6k4UHc51+gtNDeWYZdWdkmnENE8WxmsrFguft9ACixOlAB+zjrj

	IMPORTANT: Save your private key in a secure place and don't share it with anybody, not even us!

3. Enter your public key in the portal
--------------------------------------

Go back to the portal and copy your public key into the game settings page to associate the keypair with your account:

.. image:: images/game_settings_public_key.png

4. Set the public key in your client
------------------------------------

For example, in *upgraded_client.cpp* example, replace the test customer public key with your own:

.. code-block:: c++

	const char * customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";

5. Set the private key in your server
-------------------------------------

For example, you can change the code in the *upgraded_server.cpp* example:

.. code-block:: c++

	const char * customer_private_key = "OGivr2IM0k4lCfbM/VZCVK99KkDSCbzi8fzM2WnZCQb7R6k4UHc51+gtNDeWYZdWdkmnENE8WxmsrFguft9ACixOlAB+zjrj";

Or pass it in with an environment variable:

.. code-block:: console

	export NEXT_CUSTOMER_PRIVATE_KEY=OGivr2IM0k4lCfbM/VZCVK99KkDSCbzi8fzM2WnZCQb7R6k4UHc51+gtNDeWYZdWdkmnENE8WxmsrFguft9ACixOlAB+zjrj

6. Build and run your client and server
---------------------------------------

Now you should now be able to run the upgraded client and server and see the session show up in the portal. 

Make sure to run the server on a public IP address because it will not work if you are behind NAT.

Welcome to Network Next!
