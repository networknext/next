<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run your own client and server

## 1. Read SDK documentation

The Network Next SDK documentation lives here: [Network Next SDK Docs](https://network-next-sdk.readthedocs-hosted.com/en/latest/)

Take some time to read through the documentation for the example programs, and choose one of the "Upgraded" or "Complex" example programs as your starting point for your own client and server.

The upgraded example is the minimal example of a client and server with acceleration.

The complex example shows off additional features like custom allocators, custom assert functions and custom logging that are usually expected and required when you are developing console games.

## 2. Build your own client and server

I have created a test project for you under `example`. 

This project is a copy of the SDK project with client.cpp and server.cpp files based on upgraded_server.cpp/upgraded_client.cpp. If you want to start with the complex example, copy those files over and rename as client.cpp and server.cpp instead.

Customize your client.cpp to replace the buyer public key with your public key:

```
const char * buyer_public_key = "fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA==";
```

This public key is how the client handshakes with your server and establishes a secure connection. It is safe for this public key to be embedded in your client executable and known by your players.

Now for the server.cpp, we will _not_ set the buyer private key in the source code, as this is bad practice to commit secrets to your code repository. Instead, we will pass in the datacenter and private key using environment variables later.

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

And then open the generated solution file.

Build your client and server. They will be placed under `example/bin`

## 3. Run the client and server locally

Run `./example/bin/client` and `./example/bin/server`.

You will see that the client and server connect, but no acceleration is performed. The session will also not show up in your portal.

This is because the server.cpp has the datacenter is set to 'local'. When you integrate Network Next with your game, by default set the "local" datacenter too, and when you run local playtests in your LAN at your office, or run a local client and server testing, Network Next will not get in your way.

## 4. Run the server in google cloud

Manually create a VM in google cloud. n1-standard-2 type with Ubuntu 22.04 LTS is fine. 

Create the VM in the datacenter "us-central-1", or in Network Next datacenter names: "google.iowa.1".

On this VM, install premake5, then copy across the example source code, and build it:

```console
premake5 gmake
make -j
```

Set the datacenter with environment variables. This is how Network Next knows how to accelerate traffic to the correct location where your server is:

```console
export NEXT_DATACENTER=google.iowa.1
```

You could also set the datacenter to "cloud" and Network Next will autodetect where in Google, or AWS your server is running for you:

```console
export NEXT_DATACENTER=cloud
```

Set your buyer private key via environment variable:

```console
export NEXT_BUYER_PRIVATE_KEY="<your buyer private key>"
```

The Network Next SDK will now pick up your buyer private key from this environment var, so you don't need to check it in your codebase. The private key links your server to the buyer you created in the previous step.

Run your server in the VM:

```console
./bin/server
```

You should see something like this:

```
...
```

Go to the portal, and you should see a server running under your new buyer account, alongside the test server:

...

