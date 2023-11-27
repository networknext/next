<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run your own client and server

## 1. Read SDK documentation

The Network Next SDK documentation lives here: [Network Next SDK Docs](https://network-next-sdk.readthedocs-hosted.com/en/latest/)

Read through it, and choose one of the "Upgraded" or "Complex" example programs as your starting point for your own client and server, at your choice. 

The upgraded example is the minimal example of a client and server with acceleration.

The complex example shows off additional features like custom allocators, custom assert functions and detailed connection statistics.

## 2. Setup your own project to build your client and server

Depending on your platform you could setup a makefile based around the SDK makefile, or a new Visual Studio project including the SDK source and include files directly. It is your choice how you want to build. 

You will want to make copies of complex_server.cpp/complex_client.cpp or upgraded_client.cpp/upgraded_server.cpp into your own project as client.cpp and server.cpp.

Lets move forward assuming you use the upgraded example.

## 3. 
