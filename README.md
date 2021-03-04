<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Unreal Engine 4 Plugin

This repository contains the UE4 plugin for Network Next.

# Usage

1. Drop the "NetworkNext" folder under your engine "Plugins" directory.

2. Add the following to Config/DefaultEngine.ini

    [/Script/Engine.Engine]
    !NetDriverDefinitions=ClearArray
    +NetDriverDefinitions=    (DefName="GameNetDriver",DriverClassName="/Script/NetworkNext.NetworkNextNetDriver",DriverClassNameFallback="/Script/NetworkNext.NetworkNextNetDriver")

    [/Script/NetworkNext.NetworkNextNetDriver]
    NextHostname=prod.spacecats.net
    CustomerPublicKey="M/NxwbhSaPjUHES+kePTWD9TFA0bga1kubG+3vg0rTx/3sQoFgMB1w=="
    CustomerPrivateKey="M/NxwbhSaPiXITC+B4jYjdo1ahjj5NEmLaBZPPCIKL4b7c1KeQ8hq9QcRL6R49NYP1MUDRuBrWS5sb7e+DStPH/exCgWAwHX"
    NetConnectionClassName="/Script/NetworkNext.NetworkNextConnection"
