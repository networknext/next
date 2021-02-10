
Getting Started
===============

1. Sign up to the portal
------------------------

Sign up and create an account at https://portal.networknext.com by clicking "Get Access".

.. image:: images/get_access.png

2. Download the SDK
-------------------

Once you've signed up you get access to the "Downloads" page. Here you can download the latest SDK.

.. image:: images/downloads.png

3. Generate a keypair
---------------------

Inside the SDK there is a keygen tool under the "keygen" directory.

Run the keygen to generate a keypair for your company.

You'll see something like this:

.. code-block:: console

	Welcome to Network Next!

	This is your public key:

	    OGivr2IM0k7oLTQ3lmGXVnZJpxDRPFsZrKxYLn7fQAosTpQAfs464w==

	This is your private key:

	    OGivr2IM0k4lCfbM/VZCVK99KkDSCbzi8fzM2WnZCQb7R6k4UHc51+gtNDeWYZdWdkmnENE8WxmsrFguft9ACixOlAB+zjrj

	IMPORTANT: Save your private key in a secure place and don't share it with anybody, not even us!

4. Setup your company in the portal
-----------------------------------

Go to the "Settings" page in the portal and enter your company name and company code:

.. image:: images/settings.png

5. Enter your public key in the portal
--------------------------------------

Once your company is setup, the "Game Configuration" tab is revealed.

Select it and enter your public key to link your keypair with your account:

.. image:: images/public_key.png

6. Set your public key on the client
------------------------------------

The client needs only your public key to link with your account. It's safe to share your public key outside your company.

In the *upgraded_client.cpp* example, replace the test customer public key with your own:

.. code-block:: c++

	const char * customer_public_key = "OGivr2IM0k7oLTQ3lmGXVnZJpxDRPFsZrKxYLn7fQAosTpQAfs464w==";

7. Set your private key on the server
-------------------------------------

The server needs your private key to link with your account. It's very important that nobody outside your company knows your private key.

Change the code in the *upgraded_server.cpp* example to set your private key:

.. code-block:: c++

	const char * customer_private_key = "OGivr2IM0k4lCfbM/VZCVK99KkDSCbzi8fzM2WnZCQb7R6k4UHc51+gtNDeWYZdWdkmnENE8WxmsrFguft9ACixOlAB+zjrj";

Or pass it in with an environment variable:

.. code-block:: console

	export NEXT_CUSTOMER_PRIVATE_KEY=OGivr2IM0k4lCfbM/VZCVK99KkDSCbzi8fzM2WnZCQb7R6k4UHc51+gtNDeWYZdWdkmnENE8WxmsrFguft9ACixOlAB+zjrj

8. Build and run a server
-------------------------

Follow :doc:`building_the_sdk` to build the example programs.

IMPORTANT: Make sure to run the server on a public IP address somewhere because it won't work if it's behind NAT.

The server output should look like this:

.. code-block:: console

	root@linux:~/sdk# ./bin/upgraded_server
	0.000317: info: customer private key override
	0.000342: info: found valid customer private key
	0.000347: info: override next hostname: 'prod.spacecats.net'
	0.000368: info: server sdk version is 4.0.8
	0.000377: info: server address override: '173.255.241.176:50000'
	0.000387: info: server datacenter is 'linode.fremont'
	0.000419: info: server bound to 0.0.0.0:50000
	0.001378: info: server started on 173.255.241.176:50000
	0.001445: info: server resolving backend hostname 'prod.spacecats.net'
	0.001572: info: server increased thread priority
	0.103092: info: server resolved backend hostname to 34.121.72.52:40000
	1.085716: info: server received init response from backend
	1.085754: info: welcome to network next :)

9. Build and run a client
-------------------------

Modify the client source code to open a session to your server IP address:

.. code-block:: c++

    const char * server_address = "173.255.241.176:50000";

Then build and run the client. 

The client should connect to your server and exchange packets:

.. code-block:: c++

	glenn@macbook sdk % ./bin/upgraded_client
	0.010058: info: found valid customer public key
	0.010086: info: client sdk version is 4.0.8
	0.010204: info: client bound to 0.0.0.0:59617
	0.010396: info: client increased thread priority
	0.106846: info: client opened session to 173.255.241.176:50000
	0.517579: info: client received packet from server (32 bytes)
	0.517605: info: client upgraded to session a496cc24b0b5e804
	0.601095: info: client direct route
	0.768587: info: client received packet from server (32 bytes)
	1.018871: info: client received packet from server (32 bytes)
	1.273413: info: client received packet from server (32 bytes)
	1.526227: info: client received packet from server (32 bytes)
	1.781426: info: client received packet from server (32 bytes)
	2.036010: info: client received packet from server (32 bytes)
	2.288679: info: client received packet from server (32 bytes)
	...

10. See your session in the portal
----------------------------------

Go to the portal at https://portal.networknext.com

You should see a dot on your map at your location:

.. image:: images/map.png

And in the sessions page you should see your session:

.. image:: images/sessions.png

Click on your session id to drill in and see your latency, jitter and packet loss in real-time:

.. image:: images/session_drill_in.png

Welcome to Network Next!
------------------------

Congratulations, your account is now fully setup.

Please reach out to support@networknext.com and we'll get your sessions accelerated!
