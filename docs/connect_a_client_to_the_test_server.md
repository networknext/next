<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Connect a client to the test server

There is a test server already running in the dev environment.

You can see this test server running in google cloud if you look at VMs:

<img width="1428" alt="test server" src="https://github.com/user-attachments/assets/691d712d-1b7d-45f8-a86b-bc49d0e5c1db" />

Make sure source is build locally:

	make

Now can connect a client to this test server like this:

	run client

Wait a few seconds and you should see your client session live in the portal.

	(screenshot)

Don't be concerned if it takes a while for the session counts in the top/right to update. They update once per-minute, and this is an intentional feature to make the session counting tractible at high session counts (like 10-20M CCU).

Next, click on "Servers" in the nav menu, and you can see the test server now has one session:

	(screenshot)

Click on "Relays" in the nav menu, and you'll see which relays are carrying sessions:

	(screenshot)

Click on "Sessions" in the nav menu then, click on your client session in the session list, and you will see real-time statistics for your session:

	(screenshot)

The blue line is latency in milliseconds before acceleration, orange line is the conservative predicted latency after acceleration, and the green is the actual accelerated latency.

Don't be concerned if you see the accelerated latency is higher than the non-accelerated latency. By default the test client is configured to always go across network next, even if no acceleration is found. 

We will fix this in the next step.

Up next: [Run your own client and server](run_your_own_client_and_server.md).
