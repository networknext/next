# Upgraded Server Example

In this example we setup a server in Unity for monitoring and acceleration by Network Next.

LIke the simple server example, start by setting up a custom logging function to view the output of the Network Next SDK in the Unity console:
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

Assign the logging function in `Start()`:
```csharp
// Allow all logging messages to be displayed (default is NEXT_LOG_LEVEL_INFO)
Next.NextLogLevel(Next.NEXT_LOG_LEVEL_DEBUG);

// Assign our custom logging function
Next.NextLogFunction(UnityLogger);
```

Now, define configuration values for the server:
```csharp
string bindAddress = "0.0.0.0:50000";
string serverAddress = "127.0.0.1:50000";
string serverDatacenter = "local";
string serverBackendHostname = "prod.spacecats.net";
string customerPrivateKey = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn";
```

This includes the test customer private key we're using in this example. A customer private key is required on the server to enable acceleration by Network Next.

Initialize a configuration struct to defaults, then copy the hostname and the customer private key on top:
```csharp
// Get the default configuration
Next.NextConfig config;
Next.NextDefaultConfig(out config);

// Assign our private key and the server backend hostname to the configuration
config.CustomerPrivateKey = customerPrivateKey;
config.ServerBackendHostname = serverBackendHostname;
```

IMPORTANT: Generally speaking it's bad form to include a private key in your codebase like this, it's done here only to make this example easy to use. In production environments, we strongly recommend passing in "" for your customer private key, and setting it via the environment variable: *NEXT_CUSTOMER_PRIVATE_KEY* which overrides the value specified in code. In Unity, you can set environment variables in the project settings (see [here](https://support.unity.com/hc/en-us/articles/360044824951-I-need-to-start-Unity-with-an-environment-variable-s-set-how-can-I-do-that-)), or use [`System.Environment.ExpandEnvironment`](https://docs.microsoft.com/en-us/dotnet/api/system.environment.expandenvironmentvariables?view=net-5.0).

Next we initialize the SDK, this time passing in the configuration struct:
```csharp
// Initialize Network Next
if (Next.NextInit(IntPtr.Zero, ref config) != Next.NEXT_OK)
{
    Debug.LogError("error: could not initialize network next");
    this.gameObject.SetActive(false);
    return;
}
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

    Next.NextServerSendPacket(serverPtr, fromPtr, packetData, packetBytes);
    Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server received packet from client ({0} bytes)", packetBytes));

    if (!Next.NextServerSessionUpgraded(serverPtr, fromPtr))
    {
        string userIDString = "12345";
        Next.NextServerUpgradeSession(serverPtr, fromPtr, userIDString);
    }
}
```

Generally you would *not* want to upgrade every client session you receive a packet from. This is just done to make this example easy to implement.

Instead, you should only upgrade sessions that have passed whatever security and protocol level checks you have in your game so you are 100% confident this is a real player joining your game.

Also notice that you can pass in a user ID as a string to the upgrade call:
```csharp
Next.NextServerUpgradeSession(serverPtr, fromPtr, userIDString);
```
This user id is very important because it allows you to look up users by that ID in our portal. Please make sure you set the user ID to however you uniquely identify users in your game. For example, PSN IDs, Steam IDs and so on. For privacy reasons, this user ID is hashed before sending to our backend.

Now, create the server:
```csharp
// Create the packet received callback
NextServerPacketReceivedCallback recvCallBack = new NextServerPacketReceivedCallback(ServerPacketReceived);

// Create a pointer to the server (store as global var)
server = Next.NextServerCreate(IntPtr.Zero, serverAddress, bindAddress, serverDatacenter, recvCallBack, null);
if (server == IntPtr.Zero)
{
    Debug.LogError("error: failed to create server");
    this.gameObject.SetActive(false);
    return;
}
```
Note that `Next.NextServerCreate()` returns an `IntPtr`. This is intentional, since the `next_server_t` struct in C++ is complex, and we have no need to marshal and unmarshal the various struct fields between managed and unmanaged code using P/Invoke. Instead, anything SDK related with the server will accept an `IntPtr` to pass to the unmanaged C++ code.

We recommend keeping the server pointer as a global variable so that it is accessible from various Unity functions, like `Update()` and `Destroy()`.

Make sure the server gets updated every frame (best done in Unity's `Update()` loop):
```csharp
Next.NextServerUpdate(server);
```

When you have finished using your server, destroy it (Unity's `Destroy()` function is a good place to do this):
```csharp
Next.NextServerDestroy(server);
```

Finally, before your application terminates, please shut down the SDK (Unity's `OnApplicationQuit()` is an appropriate place to do this):
```csharp
Next.NextTerm();
```
