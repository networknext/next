<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run your own client and server

## 1. Read SDK documentation

The Network Next SDK documentation lives here: [Network Next SDK Docs](https://network-next-sdk.readthedocs-hosted.com/en/latest/)

Read through it, and choose one of the "Upgraded" or "Complex" example programs as your starting point for your own client and server.

The upgraded example is the minimal example of a client and server with acceleration.

The complex example shows off additional features like custom allocators, custom assert functions and detailed connection statistics.

## 2. Build your own client and server

I have created a test project for you under "example". 

This project is a copy of the SDK with client.cpp and server.cpp files based on upgraded_server.cpp/upgraded_client.cpp. If you want to start with the complex example, copy those files over and rename as client.cpp and server.cpp.

Customize your client.cpp to replace the buyer public key with your public key:

```
const char * buyer_public_key = "fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA==";
```

This public key is how the client handshakes with your server and establishes a secure connection. It is safe for this public key to be embedded in your client executable and known by your players.

Now for the server.cpp, we will _not_ set the buyer private key in the source code, as this is bad practice to commit secrets to your code repository. Instead, we will pass in the datacenter and private key using environment variables later.

Build the example locally:

For a unix inspired system and MacOS:

```
premake5 gmake
make
```

For Windows, something like:

```
premake5 vs2019
```

And then open the generated solution file.

Build your client and server.

