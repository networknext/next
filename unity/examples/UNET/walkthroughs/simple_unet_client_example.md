# Simple UNET Client Example

In this example we setup the simplest possible client in Unity using UNET and Network Next.

First, create a custom logging function to view the output of the Network Next SDK in the Unity console:
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

Next, in the client's `Start()` function, assign the custom logging function:
```csharp
// Assign our custom logging function
Next.NextLogFunction(UnityLogger);
```

This is important! If we proceed with the rest of the example _without_ setting the custom logging function, we won't be able to view any of the SDK's `next_printf()` statements.

Now, create an empty configuration for intializing the SDK.
```csharp
// Create an empty configuration
Next.NextConfig config = new Next.NextConfig();
```

Next, define a callback function to be called when packets are received:
```csharp
// Define packet receive callback function
[MonoPInvokeCallback(typeof(NextClientPacketReceivedCallback))]
static void ClientPacketReceived(IntPtr clientPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
{
    Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("client received packet from server ({0} bytes)", packetBytes));
}
```

This is where you receive packets sent via Network Next instead of polling `NetworkEventType.DataEvent` from `NetworkTransport.Receive`.

Finally, create the Next Client Transport, which initializes and supports the Network Next SDK via the Unity plugin:
```csharp
// Create the packet received callback
NextClientPacketReceivedCallback recvCallBack = new NextClientPacketReceivedCallback(ClientPacketReceived);

// Create the NextClientTransport, which sets up the Network Next client on its own socket independent of UNET
clientTransport = new NextClientTransport(IntPtr.Zero, ref config, "0.0.0.0", 0, "127.0.0.1", 50000, recvCallBack, null);
clientTransport.Init();
```
In this case we bind the client to any IPv4 address and port zero, so the system selects a port to use.
We also provide the Network Next server IP and port for the client to connect to. Note that the Network Next server runs on its own socket independent of the UNET socket.

It's worth mentioning that the `client` instance variable within `NextClientTransport` is publicly acessible after initialization. This gives you the freedom to use Network Next client related functions outside the scope of the transport. However, the transport has safety measures in place for all Network Next functions to ensure the client can be used and is connected to the Network Next server. We recommend using the transport to call Network Next client functions to prevent unknown crashes.

Also, recognize that `client` is an `IntPtr`. This is intentional, since the `next_client_t` struct in C++ is complex, and we have no need to marshal and unmarshal the various struct fields between managed and unmanaged code using P/Invoke. Instead, anything SDK related with the client will accept an `IntPtr` to pass to the unmanaged C++ code.

With a client ready to go, open a session between the client and the server over UNET (i.e. port 7777):
```csharp
// Connect to the server via UNET
int hostID = 0;
int specialConnectionID = 0;
byte error;
int connectionID = clientTransport.Connect(hostID, "127.0.0.1", 7777, specialConnectionID, out error);
```

And then over Network Next (i.e. port 50000 from above):
```csharp
// Connect to the server via Network Next
clientTransport.NextClientOpenSession();
```

In Unity's `Update()` loop, we can send packets to the server like this:
```csharp
byte[] packetData = {1};
int packetBytes = packetData.Length;
clientTransport.NextClientSendPacket(packetData, packetBytes);
```

Make sure the client is updated once every frame:
```csharp
clientTransport.NextClientUpdate();
```
Doing so drives the packet received callback we set earlier.

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
