# Simple Client Example

In this example we setup the simplest possible client in Unity using Network Next.

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

Next, define a callback function to be called when packets are received:
```csharp
// Define packet receive callback function
[MonoPInvokeCallback(typeof(NextClientPacketReceivedCallback))]
static void ClientPacketReceived(IntPtr clientPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
{
    Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("client received packet from server ({0} bytes)", packetBytes));
}
```

Finally, create the client:
```csharp
// Create the packet received callback
NextClientPacketReceivedCallback recvCallBack = new NextClientPacketReceivedCallback(ClientPacketReceived);

// Create a pointer to the client (store as global var)
client = Next.NextClientCreate(IntPtr.Zero, "0.0.0.0:0", recvCallBack, null);
if (client == IntPtr.Zero)
{
    Debug.LogError("error: failed to create client");
    this.gameObject.SetActive(false);
    return;
}
```
In this case we bind the client to any IPv4 address and port zero, so the system selects a port to use.

Note that `Next.NextClientCreate()` returns an `IntPtr`. This is intentional, since the `next_client_t` struct in C++ is complex, and we have no need to marshal and unmarshal the various struct fields between managed and unmanaged code using P/Invoke. Instead, anything SDK related with the client will accept an `IntPtr` to pass to the unmanaged C++ code.

We recommend keeping the client pointer as a global variable so that it is accessible from various Unity functions, like `Update()` and `Destroy()`.

With a client ready to go, open a session between the client and the server:
```csharp
// Open a session to the server
Next.NextClientOpenSession(client, "127.0.0.1:50000");
```

In Unity's `Update()` loop, we can send packets to the server like this:
```csharp
byte[] packetData = {1};
int packetBytes = packetData.Length;
Next.NextClientSendPacket(client, packetData, packetBytes);
```

Make sure the client is updated once every frame:
```csharp
Next.NextClientUpdate(client);
```
Doing so drives the packet received callback we set earlier.

When you have finished using your client, destroy it (Unity's `Destroy()` function is a good place to do this):
```csharp
Next.NextClientDestroy(client);
```

Finally, before your application terminates, please shut down the SDK (Unity's `OnApplicationQuit()` is an appropriate place to do this):
```csharp
Next.NextTerm();
```
