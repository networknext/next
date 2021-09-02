# Simple Server Example

In this example we setup the simplest possible server in Unity using Network Next.

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

Next, in the server's `Start()` function, allow for all logging statements and assign the custom logging function:
```csharp
// Allow all logging messages to be displayed (default is NEXT_LOG_LEVEL_INFO)
Next.NextLogLevel(Next.NEXT_LOG_LEVEL_DEBUG);

// Assign our custom logging function
Next.NextLogFunction(UnityLogger);
```

This is important! If we proceed with the rest of the example _without_ setting the custom logging function, we won't be able to view any of the SDK's `next_printf()` statements.

Now, initialize the SDK using an empty configuration:
```csharp
// Create an empty configuration
Next.NextConfig config = new Next.NextConfig();

// Initialize Network Next
if (Next.NextInit(IntPtr.Zero, ref config) != Next.NEXT_OK)
{
    Debug.LogError("error: could not initialize network next");
    this.gameObject.SetActive(false);
    return;
}
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

Now create the server.

In this example, we bind the server to port 50000 on 127.0.0.1 IPv4 address (localhost) and set the datacenter where your server is running to "local":
```csharp
// Create the packet received callback
NextServerPacketReceivedCallback recvCallBack = new NextServerPacketReceivedCallback(ServerPacketReceived);

// Create a pointer to the server (store as global var)
server = Next.NextServerCreate(IntPtr.Zero, "127.0.0.1:50000", "0.0.0.0:50000", "local", recvCallBack, null);
if (server == IntPtr.Zero)
{
    Debug.LogError("error: failed to create server");
    this.gameObject.SetActive(false);
    return;
}
```
Note that `Next.NextServerCreate()` returns an `IntPtr`. This is intentional, since the `next_server_t` struct in C++ is complex, and we have no need to marshal and unmarshal the various struct fields between managed and unmanaged code using P/Invoke. Instead, anything SDK related with the server will accept an `IntPtr` to pass to the unmanaged C++ code.

We recommend keeping the server pointer as a global variable so that it is accessible from various Unity functions, like `Update()` and `Destroy()`.

In Unity's `Update()` loop, make sure the server gets updated every frame:
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
