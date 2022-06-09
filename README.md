<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Unreal Engine 4 Plugin (console)

This repository contains the UE4 plugin for Network Next.

It's tested working with Unreal Engine 4.25 and 4.27.

This repo includes PC, Mac, Linux and console support (PS4, PS5, XBox One, XBox Series X, Nintendo Switch).

# Usage

1. Copy the entire **NetworkNext** folder into your **Plugins** directory.

2. Add the following to the bottom of **DefaultEngine.ini**

        [/Script/Engine.Engine]
        !NetDriverDefinitions=ClearArray
        +NetDriverDefinitions=  (DefName="GameNetDriver",DriverClassName="/Script/NetworkNext.NetworkNextNetDriver",DriverClassNameFallback="/Script/NetworkNext.NetworkNextNetDriver")

        [/Script/NetworkNext.NetworkNextNetDriver]
        NextHostname=prod.spacecats.net
        CustomerPublicKey="M/NxwbhSaPjUHES+kePTWD9TFA0bga1kubG+3vg0rTx/3sQoFgMB1w=="
        CustomerPrivateKey="M/NxwbhSaPiXITC+B4jYjdo1ahjj5NEmLaBZPPCIKL4b7c1KeQ8hq9QcRL6R49NYP1MUDRuBrWS5sb7e+DStPH/exCgWAwHX"
        NetConnectionClassName="/Script/NetworkNext.NetworkNextConnection"

3. Run **keygen.exe** to generate your own customer keypair.

4. Replace the keypair values in **DefaultEngine.ini** with your own keys.

5. Edit your game mode blueprint to exec **UpgradePlayer** in response to the **OnPostLogin**

<img src="https://storage.googleapis.com/network-next-ue4/blueprint.jpg" alt="Network Next" width="600"/>

6. Set environment variables on the server, so Network Next knows where your server is running.

        export NEXT_SERVER_ADDRESS=10.2.100.23:7777        # change to the public IP:port of your server
        export NEXT_DATACENTER=cloud                       # autodetects datacenter in GCP or AWS

# If you are building for PS5

**PLATFORM_PS5** must be defined and there must be a platform socket subsystem definition for PS5 for the plugin to work. If you are using a stock build of the UE4 engine source, you must do the following:

1. Add the following to "\Engine\Source\Runtime\Sockets\Public\SocketSubsystem.h" within the "PLATFORM_SOCKETSUBSYSTEM" ifndef block:
```
#elif PLATFORM_PS5
        #define PLATFORM_SOCKETSUBSYSTEM FName(TEXT("PS5"))
```
2. Add the following to "\Engine\Source\Runtime\Core\Public\HAL\Platform.h" within the platform define section at the top of the file:
```
#if !defined(PLATFORM_PS5)
	#define PLATFORM_PS5 0
#endif
```
3. Add the **PLATFORM_PS5** definition "\Engine\Platforms\PS5\Source\Programs\UnrealBuildTool\UEBuildPS5.cs" by adding the following to the "SetUpEnvironment" function:
```
CompileEnvironment.Definitions.Add("PLATFORM_PS5=1");
```
If this step is not done correctly, you will see the following error: "Building unreal engine on PS5, but PLATFORM_PS5 is not defined! Please follow steps in README.md for PS5 platform setup!"
