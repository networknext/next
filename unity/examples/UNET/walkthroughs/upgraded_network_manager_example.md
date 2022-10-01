# Upgraded Network Manager Example

In this example we setup a server for monitoring and acceleration in Unity using UNET and Network Next by extending the `NetworkManager` class.

First, define configuration values for the server:
```csharp
const string bindIP = "0.0.0.0";
const int bindPort = 50000;
const string serverIP = "127.0.0.1";
const int serverPort = 50000;
const int unetPort = 7777;
const string serverDatacenter = "local";
const string serverBackendHostname = "prod.spacecats.net";
const string customerPrivateKey = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn";
```

This includes the test customer private key we're using in this example. A customer private key is required on the server to enable acceleration by Network Next.

Like in the simple network manager example, set up a custom logging function to view the output of the Network Next SDK in the Unity console:
```csharp
// Create custom logging function to output to Unity console
[MonoPInvokeCallback(typeof(NextLogFunction))]
static void UnityLogger(int level, IntPtr formatPtr, IntPtr argsPtr)
{
    // Unmarshal the log message into a string
    string argsStr = Marshal.PtrToStringAnsi(argsPtr);

    if (level != Next.NEXT_LOG_LEVEL_NONE)
    {
        // Log to Unity console
        Debug.Log(argsStr);
    }
}
```

Assign the logging function in the manager's `Start()` function:
```csharp
// Assign our custom logging function
Next.NextLogFunction(UnityLogger);
```

You can also bypass the need for a Network Manager HUD by starting the server in `Start()` itself:
```csharp
#if UNITY_SERVER
// Set the UNET IP, port, and bind address
this.networkAddress = serverIP;
this.networkPort = unetPort;
this.serverBindAddress = String.Format("{0}:{1}", bindIP, unetPort);

// Bypass the Network Manager HUD and start the server
this.StartServer();
#endif // #if UNITY_SERVER
```

Now we define a function to be called when packets are received from clients.

Here is one that reflects the packet back to the client that sent it, and upgrades the client that sent the packet for monitoring and acceleration by Network Next:
```csharp
// Define packet receive callback function
[MonoPInvokeCallback(typeof(NextServerPacketReceivedCallback))]
public void ServerPacketReceived(IntPtr serverPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
{
    // Unmarshal the packet data into byte[]
    byte[] packetData = new byte[packetBytes];
    Marshal.Copy(packetDataPtr, packetData, 0, packetBytes);

    serverTransport.NextServerSendPacket(fromPtr, packetData, packetBytes);
    Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server received packet from client ({0} bytes)", packetBytes));

    if (!serverTransport.NextServerSessionUpgraded(fromPtr))
    {
        string userIDString = "12345";
        serverTransport.NextServerUpgradeSession(fromPtr, userIDString);
    }
}
```
Generally you would *not* want to upgrade every client session you receive a packet from. This is just done to make this example easy to implement.

Instead, you should only upgrade sessions that have passed whatever security and protocol level checks you have in your game so you are 100% confident this is a real player joining your game.

Also notice that you can pass in a user ID as a string to the upgrade call:
```csharp
serverTransport.NextServerUpgradeSession(fromPtr, userIDString);
```
This user ID is very important because it allows you to look up users by that ID in our portal. Please make sure you set the user ID to however you uniquely identify users in your game. For example, PSN IDs, Steam IDs and so on. For privacy reasons, this user ID is hashed before sending to our backend.


Now, in the manager's `OnStartServer()` callback, initialize a configuration struct to defaults, then copy the hostname and the customer private key on top:
```csharp
// Get the default configuration
Next.NextConfig config;
Next.NextDefaultConfig(out config);

// Assign our private key and the server backend hostname to the configuration
config.CustomerPrivateKey = customerPrivateKey;
config.ServerBackendHostname = serverBackendHostname;
```

IMPORTANT: Generally speaking it's bad form to include a private key in your codebase like this, it's done here only to make this example easy to use. In production environments, we strongly recommend passing in "" for your customer private key, and setting it via the environment variable: *NEXT_CUSTOMER_PRIVATE_KEY* which overrides the value specified in code. In Unity, you can set environment variables in the project settings (see [here](https://support.unity.com/hc/en-us/articles/360044824951-I-need-to-start-Unity-with-an-environment-variable-s-set-how-can-I-do-that-)), or use [`System.Environment.ExpandEnvironment`](https://docs.microsoft.com/en-us/dotnet/api/system.environment.expandenvironmentvariables?view=net-5.0).

Create the Next Server Transport.
```csharp
// Create the packet received callback
NextServerPacketReceivedCallback recvCallBack = new NextServerPacketReceivedCallback(ServerPacketReceived);

// Create the NextServerTransport, which sets up the Network Next server on its own socket independent of UNET,
// and save it as an instance var for use in other callbacks
serverTransport = new NextServerTransport(IntPtr.Zero, ref config, serverIP, serverPort, bindIP, bindPort, serverDatacenter, recvCallBack, null);
```

Lastly, set the manager's active transport to Next Server Transport:
```csharp
// Set the NextServerTransport as the active transport
activeTransport = serverTransport;
serverTransport.Init();
```

It's worth mentioning that the `server` instance variable within `NextServerTransport` is publicly acessible after initialization. This gives you the freedom to use Network Next server related functions outside the scope of the transport. However, the transport has safety measures in place for all Network Next functions to ensure the server can be used. We recommend using the transport to call Network Next server functions to prevent unknown crashes.

Also, recognize that `server` is an `IntPtr`. This is intentional, since the `next_server_t` struct in C++ is complex, and we have no need to marshal and unmarshal the various struct fields between managed and unmanaged code using P/Invoke. Instead, anything SDK related with the server will accept an `IntPtr` to pass to the unmanaged C++ code.

In the manager's `Update()` loop, make sure the server gets updated every frame:
```csharp
serverTransport?.NextServerUpdate();
```

When you have finished using your server, flush and destroy it (the `OnStopServer()` callback is a good place to do this):
```csharp
// Flush the server
serverTransport.NextServerFlush();

// Destroy the server
serverTransport.NextServerDestroy();
```

Finally, before your application terminates, please shut down the transport (Unity's `OnApplicationQuit()` is an appropriate place to do this):
```csharp
// Shutdown the transport and Network Next SDK
serverTransport?.Shutdown();
```
