<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# How to use Network Next

## 1. Direct mode

First, [build](https://github.com/networknext/sdk/blob/master/BUILDING.md) the Network Next SDK and run the client and server examples.

You should see something like:
```
Welcome to Network Next!

0.009249: info: server sdk version is 3.4.0
0.009276: info: server datacenter is 'local'
0.009317: info: server bound to port 32202
0.009872: info: server started on 127.0.0.1:32202
0.009882: info: server resolving backend hostname 'v3.networknext.com'
0.111349: info: server resolved backend hostname to 34.69.211.185:40000
```

and:
```

Welcome to Network Next!

0.009428: info: client sdk version is 3.4.0
0.009503: info: client bound to port 63585
0.110334: info: client opened session to 127.0.0.1:32202
```

At this point the client and server are exchanging packets, but they are not yet accelerated.

## 2. Create an account on our portal

Email sales@networknext.com to get access to our portal:

https://app.v3.networknext.com

## 3. Generate a private/public keypair

Once you have portal access, we need to generate a keypair so you can securely request accelerated routes for your players.

Run the keygen tool under the 'keygen' directory in the SDK:

You should see something like:
```
This is your public key:

    PuOtENNoEI7jewa7Kqvl3SHmRKnn6PzAquBhWO1qTRN0H8eJUzUXuA==

This is your private key:

    PuOtENNoEI77YEjPhWbcEGsG5L5dVE7if1KxwRpwvYbBSy5uL1OWJeN7Brsqq+XdIeZEqefo/MCq4GFY7WpNE3Qfx4lTNRe4
```

Paste the base64 public key into the web portal "Game Configuration" page and click save.

Next, copy the private key base64 into server.cpp:
```
static const char * customer_private_key = "PuOtENNoEI77YEjPhWbcEGsG5L5dVE7if1KxwRpwvYbBSy5uL1OWJeN7Brsqq+XdIeZEqefo/MCq4GFY7WpNE3Qfx4lTNRe4";
```

and the public key into client.cpp:
```
static const char * customer_public_key = "PuOtENNoEI7jewa7Kqvl3SHmRKnn6PzAquBhWO1qTRN0H8eJUzUXuA==";
```

## 4. Client and server running in 'upgraded' mode

Now rebuild and run the client and server.

You should now see something like:
```
Welcome to Network Next!

0.000179: info: server datacenter is 'local'
0.000494: info: server started on 127.0.0.1:32202
0.000521: info: server resolving hostname udp-v3
0.120935: info: server resolved hostname to 10.99.2.5:40001
2.058923: info: server upgraded client 192.168.65.3:44174 to session 7f127fe26281d1c4
```
and:
```
Welcome to Network Next!

0.015976: info: client opened session to 127.0.0.1:32202
0.064203: info: client upgraded to session 7f127fe26281d1c4
```

Your client has now been 'upgraded' and is ready to be accelerated by Network Next.

Please email sales@networknext.com for next steps.

