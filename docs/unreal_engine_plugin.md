<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Unreal Engine Plugin

Network Next comes with a drop-in Unreal Engine plugin that replaces the NetDriver.

## 1. Copy the entire **NetworkNext** folder into your **Plugins** directory.

_The Unreal Engine is still being updated, but will be placed under the `~/next/unreal` directory when it is ready (soon!)_

## 2. Add the following to the bottom of **DefaultEngine.ini**

```
[/Script/Engine.Engine]
!NetDriverDefinitions=ClearArray
+NetDriverDefinitions=  (DefName="GameNetDriver",DriverClassName="/Script/NetworkNext.NetworkNextNetDriver",DriverClassNameFallback="/Script/NetworkNext.NetworkNextNetDriver")

[/Script/NetworkNext.NetworkNextNetDriver]
ServerBackendHostname=server-dev.[yourdomain.com]
BuyerPublicKey="<your buyer public key>"
NetConnectionClassName="/Script/NetworkNext.NetworkNextConnection"
```

## 3. Scan all ini files in your game project in case somewhere else is clobbering the NetDriver setting. 

If this is the case the Network Next plugin will not work. 

This is a common failure point during integration for our Unreal Engine customers.

## 4. Edit your game mode blueprint to exec **UpgradePlayer** in response to the **OnPostLogin**

## 5. Set environment variables on the server

```
export NEXT_SERVER_ADDRESS=10.2.100.23        # change to the public IP of your server
export NEXT_DATACENTER=cloud                  # autodetects datacenter in GCP or AWS
export NEXT_BUYER_PRIVATE_KEY=<your buyer private key>
```

## 6. Verify your sessions and servers show up in the portal

When everything is working correctly, you will see your client sessions showing up in the portal under "Sessions", and your servers will show up under "Servers".

Because the Network Next SDK takes great care to fall back to no acceleration, make sure to verify that you see your sessions and servers in the portal, post integration, so that you know Network Next has not fallen back to a non-accelerated state.

Congratulations! You have successfully integrated Network Next with your Unreal Game!

[Back to main documentation](README.md)
