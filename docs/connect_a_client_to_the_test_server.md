<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Connect a client to the test server

There is a test server already running in the dev environment.

You can see this test server running in google cloud if you look at VMs:

<img width="1428" alt="test server" src="https://github.com/user-attachments/assets/691d712d-1b7d-45f8-a86b-bc49d0e5c1db" />

First, make sure source is built locally:

```
cd ~/next
make
```

Now can connect a client to this test server like this:

```
run client
```

Wait a few seconds and you should see your client session live in the portal:

<img width="1456" alt="client session" src="https://github.com/user-attachments/assets/2874825f-b8f8-4484-a3ad-9c899a3f4311" />

Don't be concerned if it takes a while for the session counts in the top/right to update. They update once per-minute, and this is an intentional feature to make the session counting tractible at high session counts (like 10-20M CCU).

Next, click on "Servers" in the nav menu, and you can see the test server now has one session:

<img width="1456" alt="test server carrying session" src="https://github.com/user-attachments/assets/863c2104-8ced-4a55-b52e-58a1dec4d0d5" />

Click on "Relays" in the nav menu, and you'll see which relays are carrying your session:

<img width="1456" alt="relays carrying session" src="https://github.com/user-attachments/assets/931ab4d9-c492-4765-9355-1a8347040e8d" />

Click on "Sessions" in the nav menu then and then click on your client session in the session list, and you will see real-time statistics for your session:

<img width="1457" alt="client session details" src="https://github.com/user-attachments/assets/53abb823-72c6-4627-9a71-b0939ff97cad" />

The blue line is latency in milliseconds before acceleration, orange line is the conservative predicted latency after acceleration, and the green is the actual accelerated latency.

Don't be concerned if you see the accelerated latency is higher than the non-accelerated latency. By default the test client is configured to always go across network next, even if no acceleration is found. 

We will fix this in the next step.

Up next: [Modify route shader for test buyer](modify_route_shader_for_test_buyer.md).
