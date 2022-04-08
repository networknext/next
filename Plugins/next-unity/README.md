<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Unity Plugin

This repository contains the Unity Native Plugin for Network Next SDK version 4.20.0.

It's tested and works with Unity Engine 2018.4.36f1 and 2020.3.14f1.

Supported Platforms:
- Windows
- MacOS
- Linux

# Installation

Import the plugin while the Unity Editor is open by double-clicking `next-unity.unitypackage`, or by navigating to Assets > Import Package > Custom Package > `next-unity.unitypackage`.
Please remember to close and reopen the editor for the plugin to load.

# Overview

This plugin can be used in two ways:
1. [`Next.cs`](https://github.com/networknext/unity/blob/main/Plugins/next-unity/Next.cs) is a C# wrapper for Network Next's C++ SDK. This gives you access to every SDK function and allows you to setup a traditional client and server in Unity.
2. [`NextClientTransport.cs`](https://github.com/networknext/unity/blob/main/Plugins/next-unity/NextClientTransport.cs) and [`NextServerTransport.cs`](https://github.com/networknext/unity/blob/main/Plugins/next-unity/NextServerTransport.cs) enable you to integrate Network Next in UNET (Unity's deprecated multiplayer solution). The transports utilize the `next-unity` plugin to open a separate socket for Network Next communications to run in parallel with UNET.

Documentation for the C++ SDK is available [here](https://network-next-sdk.readthedocs-hosted.com/en/latest/). You can follow along with the [reference](https://network-next-sdk.readthedocs-hosted.com/en/latest/reference.html), and the accompanying C# documentation is viewable in `Next.cs`.

Examples and walkthroughs for this plugin are available in the [Github repository](https://github.com/networknext/unity).

# Getting Started

### Traditional Setup
1. Follow steps 1 - 5 in the  Getting Started section of the [SDK documentation](https://network-next-sdk.readthedocs-hosted.com/en/latest/getting_started.html) to setup your account and generate a public and private key pair.
2. Replace the public and private keys in the `UpgradedClient.cs` and `UpgradedServer.cs` examples with your newly generated keys.
3. In the editor, create two game objects called `UpgradedServer` and `UpgradedClient` and add the respective scripts as components in the Inspector.
4. Deactivate the `UpgradedClient` game object and press play. You should see logs in the Console indicating the server has started.
5. During play mode, activate the `UpgradedClient` game object. The client will connect to the server and begin exchanging packets, and the session will be visible on the [portal](https://portal.networknext.com/).

### UNET Setup
1. Follow steps 1 - 5 in the  Getting Started section of the [SDK documentation](https://network-next-sdk.readthedocs-hosted.com/en/latest/getting_started.html) to setup your account and generate a public and private key pair.
2. Replace the public and private keys in the `UpgradedUNETClient.cs` and `UpgradedNetworkManager.cs` examples with your newly generated keys.
3. In the editor, create a `Player` game object (i.e. Cube) and add the `UpgradedUNETClient.cs` script as a component. Add UNET's `NetworkIdentity` script as another component and check the "Local Player Authority" box. Make `Player` a prefab and remove it from the scene.
4. Create a `NetworkManager` game object and add the `UpgradedNetworkManager.cs` script as a component. In the inspector under the Spawn Info tab, set the "Player Prefab" as the `Player` prefab from step 3. Make sure the "Auto Create Player" box is unchecked. Finally add UNET's `NetworkManagerHUD` script as another component and check the "Show Runtime GUI" box.
5. Build the project by going to File > Build Settings. Add the current scene to the "Scenes In Build" section, and ensure your "Target Platform" is correct. In the "Player Settings..." section, change the "Fullscreen Mode" to Windowed with a default screen width and height of 800 and 600, and check the "Use Player Log" box. Finally, back in the Build Settings, press Build And Run. Keep this window on the side.
6. Return to the editor and press play. Start a LAN server, and in the windowed build from step 5, select LAN Client. The client will connect to the server over both UNET and Network Next and begin exchanging packets (visible in the Console), and the session will be visible on the [portal](https://portal.networknext.com/).

### Welcome to Network Next!
Congratulations, your account is now fully setup.

Please reach out to support@networknext.com and we’ll get your sessions accelerated!
