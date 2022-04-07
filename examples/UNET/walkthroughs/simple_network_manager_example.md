# Simple Network Manager Example

In this example we setup the simplest possible server in Unity using UNET and Network Next by extending the `NetworkManager` class.

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

Next, in the manager's `Start()` function, assign the custom logging function:
```csharp
// Assign our custom logging function
Next.NextLogFunction(UnityLogger);
```

This is important! If we proceed with the rest of the example _without_ setting the custom logging function, we won't be able to view any of the SDK's `next_printf()` statements.

You can also bypass the need for a Network Manager HUD by starting the server in `Start()` itself:
```csharp
#if UNITY_SERVER
// Set the UNET IP, port, and bind address
this.networkAddress = "127.0.0.1";
this.networkPort = 7777;
this.serverBindAddress = "0.0.0.0:7777";

// Bypass the Network Manager HUD and start the server
this.StartServer();
#endif // #if UNITY_SERVER
```

Next, define a callback function to be called when packets are received from clients.

Here is one that reflects the packet back to the client that sent it:
```csharp
// Define packet receive callback function
[MonoPInvokeCallback(typeof(NextServerPacketReceivedCallback))]
public void ServerPacketReceived(IntPtr serverPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
{
    // Unmarshal the packet data into byte[]
    byte[] packetData = new byte[packetBytes];
    Marshal.Copy(packetDataPtr, packetData, 0, packetBytes);

    Next.NextServerSendPacket(serverPtr, fromPtr, packetData, packetBytes);
}
```

This is where you receive packets sent via Network Next instead of polling `NetworkEventType.DataEvent` from `NetworkTransport.Receive`.

Now, in the manager's `OnStartServer()` callback, create an empty configuration for intializing the SDK:
```csharp
// Create an empty configuration
Next.NextConfig config = new Next.NextConfig();
```

Create the Next Server Transport.
In this example, we bind the server to port 50000 on 127.0.0.1 IPv4 address (localhost) and set the datacenter where your server is running to "local":
```csharp
// Create the packet received callback
NextServerPacketReceivedCallback recvCallBack = new NextServerPacketReceivedCallback(ServerPacketReceived);

// Create the NextServerTransport, which sets up the Network Next server on its own socket independent of UNET,
// and save it as an instance var for use in other callbacks
serverTransport = new NextServerTransport(IntPtr.Zero, ref config, "127.0.0.1", 50000, "0.0.0.0", 50000, "local", recvCallBack, null);
```
Note that the Network Next server is running on a separate socket than UNET.
We also recommend storing the transport as an instance variable for use in other functions and callbacks.

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
