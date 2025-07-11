<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

# Unreal Engine 5 Plugin

This repository contains the UE5 plugin for Network Next.

The plugin supports Windows, Linux, Mac, PS4, PS5 and GDK (XBoxOne, SeriesX and PC).

It's tested working with Unreal Engine 5.3, 5.4 and 5.6.

# Usage

1. Copy the entire **NetworkNext** folder into your **Plugins** directory.

2. Add the following to the bottom of **DefaultEngine.ini**
  
```
[/Script/Engine.Engine]
!NetDriverDefinitions=ClearArray
+NetDriverDefinitions=(DefName="GameNetDriver",DriverClassName="/Script/NetworkNext.NetworkNextNetDriver",DriverClassNameFallback="/Script/NetworkNext.NetworkNextNetDriver")

[/Script/NetworkNext.NetworkNextNetDriver]
NetConnectionClassName="/Script/NetworkNext.NetworkNextConnection"
```
          
3. Replace the keypair values in **NetworkNextModule.cpp** with your own buyer keys.

4. Call FNetworkNextModule::UpdatePlayer from your game mode OnPostLogin function C++ to enable acceleration for that player.

5. Set environment variables on the server, so Network Next knows where your server is running.

```
export NEXT_SERVER_ADDRESS=10.2.100.23:7777        # change to the public IP:port of your server
export NEXT_DATACENTER=cloud                       # autodetects datacenter in GCP or AWS
```

6. If setting environment variables is not possible, you can pass the values in using command line parameters:

```
-serverAddress=10.2.100.23:7777
-datacenter=cloud
```
