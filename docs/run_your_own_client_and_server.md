<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run your own client and server

## 1. Read SDK documentation

The Network Next SDK documentation lives here: [Network Next SDK Docs](https://network-next-sdk.readthedocs-hosted.com/en/latest/)

Take some time to read through the documentation for the example programs, and choose one of the "Upgraded" or "Complex" example programs as your starting point for your own client and server.

The upgraded example is the minimal example of a client and server with acceleration.

The complex example shows off additional features like custom allocators, custom assert functions and custom logging that are helpful when developing a console games.

## 2. Build your own client and server

There is a premake example project created for you already under `example`. 

This project is a copy of the SDK with client.cpp and server.cpp files based on upgraded_server.cpp/upgraded_client.cpp. 

If you want to start with the complex example, copy those files over and rename as client.cpp and server.cpp instead.

Customize client.cpp by replacing the buyer public key with your own:

```
const char * buyer_public_key = "<your new buyer public key here>";
```

This public key is how your client handshakes with your server and establishes a secure connection. It's safe to embed it in your client executable.

Now for the server.cpp, we will _not_ set the buyer private key in the source code, because it's bad security to commit secrets to your repository. Instead, we will pass in the private key using environment variables later on.

Build your client and server. They will be placed under `example/bin`

For Linux and MacOS:

```console
cd ~/next/example
premake5 gmake
make -j
```

For Windows, something like:

```console
premake5 vs2019
```

And then open the generated solution file and build all.

## 3. Run the client and server locally

Run `./bin/server`. You will see something like:

```console
gaffer@macbook example % ./bin/server
0.010114: info: platform is mac (wi-fi)
0.010352: info: server input datacenter is 'local' [249f1fb6f3a680e8]
0.010457: info: server bound to 0.0.0.0:50000
0.011503: info: server started on 127.0.0.1:50000
0.012162: info: server thread set to high priority
```

Now run the client, it will automatically connect to the server:

```console
gaffer@macbook example % ./bin/client
0.009847: info: platform is mac (wi-fi)
0.010066: info: found valid buyer public key: 'fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA=='
0.010293: info: client bound to 0.0.0.0:52247
0.010675: info: client thread set to high priority
0.113645: info: client opened session to 127.0.0.1:50000
0.521031: info: client received packet from server (32 bytes)
0.773261: info: client received packet from server (32 bytes)
1.028359: info: client received packet from server (32 bytes)
1.283558: info: client received packet from server (32 bytes)
1.538784: info: client received packet from server (32 bytes)
etc...
```

You will see that the client and server connect, but no acceleration is performed and your client session will _not_ show up in the portal. This is because the server.cpp has the datacenter set to 'local' by default. 

When you integrate Network Next with your game by default set the "local" datacenter there too, and when you run local playtests in your LAN at your office, or run a local server for testing, Network Next will not get in your way.

## 4. Run the server in google cloud

Manually create a VM in google cloud. n1-standard-2 type with Ubuntu 22.04 LTS is fine. 

Create the VM in the datacenter in region "us-central1 (Iowa)" and zone "us-central1-a"

Zip up the `~/next/example` directory on your local machine and copy it to google cloud storage:

```console
gsutil cp example.zip gs://[company_name]_network_next_dev
```

Then SSH into your VM, install some needed things, and copy the example zip file down and unzip it:

```console
sudo apt update && sudo apt install -y build-essential unzip
gsutil cp gs://[company_name]_network_next_dev/example.zip .
unzip example.zip
```

Install premake5 on the VM:

```console
wget https://github.com/premake/premake-core/releases/download/v5.0.0-beta2/premake-5.0.0-beta2-linux.tar.gz
tar -zxf *.tar.gz
```

Build the example code on the VM:

```console
premake5 gmake
make -j
```

Set the datacenter with environment variables on the VM. This is how the Network Next backend knows how to accelerate traffic to the datacenter where your server is located physically:

```console
export NEXT_DATACENTER=google.iowa.1
```

You can also set the datacenter to "cloud" and Network Next will automatically detect which Google or AWS datacenter your server is running in:

```console
export NEXT_DATACENTER=cloud
```

Set your buyer private key via environment variable:

```console
export NEXT_BUYER_PRIVATE_KEY="<your buyer private key>"
```

The Network Next SDK will pick up your buyer private key from this environment var and links your server to your buyer.

Make sure that UDP port 50000 is open in the firewall to receive packets. If you are not familiar with how to do this in Google Cloud, read this StackOverflow page: [https://stackoverflow.com/questions/21065922/how-to-open-a-specific-port-such-as-9090-in-google-compute-engine]

Set the server IP address in an environment var so Network Next knows which IP address your server is listening on:

```console
export NEXT_SERVER_ADDRESS=<your server external ip address>
```

Now run your server in the VM:

```console
./bin/server
```

You should see something like this:

```console
glenn@test-server-006:~/example$ ./bin/server
0.000204: info: platform is linux (wired)
0.000272: info: log level overridden to 4
0.000311: info: buyer private key override
0.000315: info: found valid buyer private key
0.000333: info: server address override: '34.67.212.136'
0.000376: info: server datacenter override 'cloud'
0.000387: info: server input datacenter is 'cloud' [9ebb5c9513bac4fe]
0.000428: info: server bound to 0.0.0.0:50000
0.003485: info: server initializing with backend
0.003641: info: server started on 34.67.212.136:50000
0.003792: info: server resolving backend hostname 'server-dev.virtualgo.net'
0.004722: info: server attempting to autodetect datacenter
0.004813: info: server autodetect datacenter: looking for curl
0.014795: info: server autodetect datacenter: curl exists
0.019241: info: server autodetect datacenter: running in google cloud
0.031601: info: server autodetect datacenter: google zone is "us-central1-a"
0.107748: info: server resolved backend hostname to 35.208.158.14:40000
0.172456: info: server autodetect datacenter: "us-central1-a" -> "google.iowa.1"
0.211762: info: server autodetected datacenter 'google.iowa.1' [aedb4f6e4bb13649]
0.214655: info: server received init response from backend
0.214675: info: welcome to network next :)
0.323384: info: server datacenter is 'google.iowa.1'
0.323412: info: server is ready to receive client connections
```

The critical thing that you must see is `welcome to network next :)`. 

This indicates that your server has successfully connected and authenticated with the Network Next backend.

Wait a minute for the portal to update, then verify that you see a server running under your new buyer account in the portal:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/526357e9-0842-428a-909d-966bd688a9e7">

## 5. Connect a client to your server in google cloud

Modify client.cpp so that it connects to the external IP address of your Google Cloud VM on port 50000:

```
const char * server_address = "34.67.212.136:50000";
```

Build and run your client:

```
make -j
./bin/client
```

Your client should connect to the server and have its connectioned 'upgraded' by the server. This starts the process of tracking your session in the portal, and deciding if it needs to be accelerated or not.

```
gaffer@macbook next % run client
0.003968: info: platform is mac (wi-fi)
0.004034: info: log level overridden to 4
0.004042: info: found valid buyer public key: 'fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA=='
0.004053: info: valid server backend public key
0.004056: info: valid relay backend public key
0.004117: info: client bound to 0.0.0.0:58230
0.105377: info: client opened session to 34.67.212.136:50000
0.603100: info: client upgraded to session ceddd63ffeb21499
```

## 6. See your client session in the portal

Now that your client has connected to the server and completed the upgrade process, you can see your session in the portal under your new buyer:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/ed330614-e22d-4c7e-ba72-95cbfaa21bc9">

Next step: [Integrate with your game](integrate_with_your_game.md)
