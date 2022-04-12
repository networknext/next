<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Unreal Engine 4 Plugin

This repository contains the UE4 plugin for Network Next.

It's tested working with Unreal Engine 4.25.

This repo includes PC, Mac and Linux support. 

Console support is available by request at support@networknext.com

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
