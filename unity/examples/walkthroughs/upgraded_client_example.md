# Upgraded Client Example

In this example we setup a client in Unity for monitoring and acceleration by Network Next.

First, define configuration values for the client:
```csharp
const string bindAddress = "0.0.0.0:0";
const string serverAddress = "127.0.0.1:50000";
const string customerPublicKey = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";
```

These include the bind address for the client socket, the server address to connect to, and the test customer public key we're using in this example. A customer public key is required to enable acceleration by Network Next.

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

Initialize the SDK, this time passing the custom configuration struct:
```csharp
// Initialize Network Next
if (Next.NextInit(IntPtr.Zero, ref config) != Next.NEXT_OK)
{
    Debug.LogError("error: could not initialize network next");
    this.gameObject.SetActive(false);
    return;
}
```

This activates the customer public key so it's used by the client.

Network Next needs a customer public key to monitor and accelerate players. Without a customer public key, Network Next just sends player traffic across the public internet.

Next, define a function to be called when packets are received:
```csharp
// Define packet receive callback function
[MonoPInvokeCallback(typeof(NextClientPacketReceivedCallback))]
static void ClientPacketReceived(IntPtr clientPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
{
    Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("client received packet from server ({0} bytes)", packetBytes));
}
```

Create the client:
```csharp
// Create the packet received callback
NextClientPacketReceivedCallback recvCallBack = new NextClientPacketReceivedCallback(ClientPacketReceived);

// Create a pointer to the client (store as global var)
client = Next.NextClientCreate(IntPtr.Zero, bindAddress, recvCallBack, null);
if (client == IntPtr.Zero)
{
    Debug.LogError("error: failed to create client");
    this.gameObject.SetActive(false);
    return;
}
```

Open a session between the client and the server:
```csharp
// Open a session to the server
Next.NextClientOpenSession(client, serverAddress);
```

Note that `Next.NextClientCreate()` returns an `IntPtr`. This is intentional, since the `next_client_t` struct in C++ is complex, and we have no need to marshal and unmarshal the various struct fields between managed and unmanaged code using P/Invoke. Instead, anything SDK related with the client will accept an `IntPtr` to pass to the unmanaged C++ code.

We recommend keeping the client pointer as a global variable so that it is accessible from various Unity functions, like `Update()` and `Destroy()`.

Now you can send packets to the server like this:
```csharp
byte[] packetData = {1};
int packetBytes = packetData.Length;
Next.NextClientSendPacket(client, packetData, packetBytes);
```

Make sure the client is updated once a frame (best done in Unity's `Update()` loop):
```csharp
Next.NextClientUpdate(client);
```

When you have finished your session with the server, close it:
```csharp
// Close the session
Next.NextClientCloseSession(client);
```

When you have finished using your client, destroy it (Unity's `OnDestroy()` function is a good place to do this):
```csharp
Next.NextClientDestroy(client);
```

Finally, before your application terminates, please shut down the SDK (Unity's `OnApplicationQuit()` is an appropriate place to do this):
```csharp
Next.NextTerm();
```
