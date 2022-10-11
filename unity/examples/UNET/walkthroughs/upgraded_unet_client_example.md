# Upgraded UNET Client Example

In this example we setup a client for monitoring and acceleration in Unity using UNET and Network Next.

First, define configuration values for the client:
```csharp
const string bindIP = "0.0.0.0";
const int bindPort = 0;
const string serverIP = "127.0.0.1";
const int serverPort = 50000;
const int unetPort = 7777;
const int hostID = 0;
const string customerPublicKey = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";
```

These include the bind address for the Network Next client socket, the server address to connect to, the Network Next and UNET ports to use, and the test customer public key we're using in this example. A customer public key is required to enable acceleration by Network Next.

Like in the simple client example, set up a custom logging function to view the output of the Network Next SDK in the Unity console:
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
// Assign our custom logging function
Next.NextLogFunction(UnityLogger);
```

Next, initialize a configuration struct to defaults, then copy the customer public key on top:
```csharp
// Get the default configuration
Next.NextConfig config;
Next.NextDefaultConfig(out config);

// Assign our public key to the configuration
config.CustomerPublicKey = customerPublicKey;
```

This activates the customer public key so it's used by the client.

Network Next needs a customer public key to monitor and accelerate players. Without a customer public key, Network Next just sends player traffic across the public internet.

Next, define a callback function to be called when packets are received:
```csharp
// Define packet receive callback function
[MonoPInvokeCallback(typeof(NextClientPacketReceivedCallback))]
static void ClientPacketReceived(IntPtr clientPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
{
    Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("client received packet from server ({0} bytes)", packetBytes));
}
```

Create the Next Client Transport:
```csharp
// Create the packet received callback
NextClientPacketReceivedCallback recvCallBack = new NextClientPacketReceivedCallback(ClientPacketReceived);

// Create the NextClientTransport, which sets up the Network Next client on its own socket independent of UNET
clientTransport = new NextClientTransport(IntPtr.Zero, ref config, bindIP, bindPort, serverIP, serverPort, recvCallBack, null);
clientTransport.Init();
```

It's worth mentioning that the `client` instance variable within `NextClientTransport` is publicly acessible after initialization. This gives you the freedom to use Network Next client related functions outside the scope of the transport. However, the transport has safety measures in place for all Network Next functions to ensure the client can be used and is connected to the Network Next server. We recommend using the transport to call Network Next client functions to prevent unknown crashes.

Also, recognize that `client` is an `IntPtr`. This is intentional, since the `next_client_t` struct in C++ is complex, and we have no need to marshal and unmarshal the various struct fields between managed and unmanaged code using P/Invoke. Instead, anything SDK related with the client will accept an `IntPtr` to pass to the unmanaged C++ code.

Open a session between the client and the server over UNET and Network Next:
```csharp
// Connect to the server via UNET
byte error;
connectionID = clientTransport.Connect(hostID, serverIP, unetPort, 0, out error);

// Connect to the server via Network Next
clientTransport.NextClientOpenSession();
```

You can send packets to the server like this:
```csharp
byte[] packetData = {1};
int packetBytes = packetData.Length;
clientTransport.NextClientSendPacket(packetData, packetBytes);
```

Make sure the client is updated once every frame (best done in Unity's `Update()` loop):
```csharp
clientTransport.NextClientUpdate();
```

When you have finished using your client, disconnect from the server and destroy the client (Unity's `OnDestroy()` function is a good place to do this):
```csharp
// Disconnect from the server and close the session
byte error;
clientTransport.Disconnect(hostID, connectionID, out error);

// Destroy the client
clientTransport.NextClientDestroy();
```

Finally, before your application terminates, please shut down the transport (Unity's `OnApplicationQuit()` is an appropriate place to do this):
```csharp
// Shutdown the transport and the Network Next SDK
clientTransport?.Shutdown();
```
