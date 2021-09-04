# Complex Client Example

In this example we build the kitchen sink version of a client in Unity where we show off all the features :)

We demonstrate:
- Setting the network next log level
- Setting a custom log function
- Setting a custom assert handler
- Setting a custom allocator
- Querying the port the client socket is bound to
- Getting statistics from the client

This is going to be a huge example, so let's get started!

First, we start by defining our key configuration variables:
```csharp
const string bindAddress = "0.0.0.0:0";
const string serverAddress = "127.0.0.1:50000";
const string customerPublicKey = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";
```

Next, we dive right in and define a custom allocator class that tracks all allocations made, and checks them for leaks when it shuts down:
```csharp
// Empty structs use 1 byte
struct AllocatorEntry
{
    // ...
}

[StructLayout (LayoutKind.Sequential)]
public class Allocator
{
    // Constants
    public const int MAX_ALLOCATIONS = 2048;

    // Global vars
    public int numAllocations;
    public Next.NextMutex mutex;
    [MarshalAs(UnmanagedType.ByValArray, SizeConst=MAX_ALLOCATIONS)]
    public IntPtr[] entries;

    public Allocator()
    {
        // Intialize globals
        int result = Next.NextMutexCreate(out mutex);
        Next.NextAssert(result == Next.NEXT_OK);
        numAllocations = 0;
        entries = new IntPtr[MAX_ALLOCATIONS];
    }

    ~Allocator()
    {
        // Cleanup
        Next.NextMutexDestroy(ref mutex);
        Next.NextAssert(numAllocations == 0);
        Next.NextAssert(Array.Exists(entries, element => element == null || element.Equals(IntPtr.Zero)));
    }

    public IntPtr Alloc(int bytes)
    {
        Next.NextMutexAcquire(ref mutex);

        // Verify we are not over the max number of allocations
        if (numAllocations >= MAX_ALLOCATIONS)
        {
            Next.NextPrintf(Next.NEXT_LOG_LEVEL_WARN, String.Format("exceeded max allocations {0}, cannot allocate further", numAllocations));
            Next.NextMutexRelease(ref mutex);
            return IntPtr.Zero;
        }

        // Create a pointer for an allocator entry and allocate memory
        AllocatorEntry entry = new AllocatorEntry();
        IntPtr entryPtr = Marshal.AllocHGlobal(bytes);
        Marshal.StructureToPtr(entry, entryPtr, false);

        Next.NextAssert(entryPtr != IntPtr.Zero);
        Next.NextAssert(!Array.Exists(entries, element => element.Equals(entryPtr)));

        // Add the pointer to the array and increment globals
        for (int i = 0; i < entries.Length; i++)
        {
            if (entries[i] == null || entries[i].Equals(IntPtr.Zero))
            {
                entries[i] = entryPtr;
                break;
            }
        }
        numAllocations++;

        Next.NextMutexRelease(ref mutex);
        
        return entryPtr;
    }

    public void Free(IntPtr entryPtr)
    {
        Next.NextMutexAcquire(ref mutex);

        Next.NextAssert(entryPtr != IntPtr.Zero);
        Next.NextAssert(numAllocations > 0);
        Next.NextAssert(Array.Exists(entries, element => element.Equals(entryPtr)));

        // Remove the pointer and free its memory
        for (int i = 0; i < entries.Length; i++)
        {
            if (entries[i] == entryPtr)
            {
                entries[i] = IntPtr.Zero;
                break;
            }
        }
        numAllocations--;
        Marshal.FreeHGlobal(entryPtr);

        Next.NextMutexRelease(ref mutex);
    }
}
```

Because this class will be passed between managed and unmanaged code, all instance variable data types must be compatible with P/Invoke interop marshaling. As a result, we opt to use an array with a fixed size compared to a hash map since it works nicely with P/Invoke. However, you can certainly create your own custom hash map class that can be marshaled.

IMPORTANT: Since this allocator will be called from multiple threads, it must be thread safe. This is done by using the platform independent mutex supplied by the Network Next SDK.

There are three types of allocations done by the Network Next SDK:
1. Global allocations
2. Per-client allocations
3. Per-server allocations

Each of these situations corresponds to what is called a "context" in the Network Next SDK. 

A context is simply a custom struct that you define which is passed in to malloc and free callbacks that we call to perform allocations on behalf of the SDK. The context passed is gives you the flexibility to have a specific memory pool for Network Next (most common), or even to have a completely different allocation pool used for each client and server instance. That's what we're going to do in this example.

Let's define a base context that will be used for global allocations:
```csharp
[StructLayout (LayoutKind.Sequential)]
public struct Context
{
    public IntPtr AllocatorGCH;
}
```

And a per-client context that is binary compatible with the base context, to be used for per-client allocations:
```csharp
[StructLayout (LayoutKind.Sequential)]
public struct ClientContext
{
    public IntPtr AllocatorGCH;
    public uint ClientData;
    public IntPtr LastPacketReceiveTimeGCH;
}
```
As you can see, the client context can contain additional information aside from the allocator. The context is not *just* passed into allocator callbacks, but to all callbacks from the client and server, so you can use it to integrate with your own client and server objects in your game. 

Here we just put a dummy `uint` in the client context and check its value to verify it's being passed through correctly. For example, in the received packet callback, we have access to the client context and check the dummy value is what we expect:
```csharp
// Define packet receive callback function
[MonoPInvokeCallback(typeof(NextClientPacketReceivedCallback))]
static void ClientPacketReceived(IntPtr clientPtr, IntPtr ctxPtr, IntPtr packetDataPtr, int packetBytes)
{
    // Unmarshal the context pointer into the client context to access its fields
    ClientContext ctx = (ClientContext)Marshal.PtrToStructure(ctxPtr, typeof(ClientContext));

    Next.NextAssert(!ctx.Equals(default(ClientContext)));
    
    GCHandle allocatorGCH = GCHandle.FromIntPtr(ctx.AllocatorGCH);
    Allocator allocator = (Allocator)allocatorGCH.Target; 

    Next.NextAssert(allocator != null);
    Next.NextAssert(ctx.ClientData == 0x12345);
    
    // Unmarshal the packet data into byte[]
    byte[] packetData = new byte[packetBytes];
    Marshal.Copy(packetDataPtr, packetData, 0, packetBytes);

    if (VerifyPacket(packetData, packetBytes))
    {
        // Update the last packet receive time
        GCHandle lastPacketReceiveTimeGCH = GCHandle.FromIntPtr(ctx.LastPacketReceiveTimeGCH);
        LastPacketReceiveTime lastPacketReceiveTime = (LastPacketReceiveTime)lastPacketReceiveTimeGCH.Target;

        lastPacketReceiveTime.SetTime(Next.NextTime());
    }
}
```
We also mark the last packet receive time for the client. That way if the client hasn't received a packet in a long time, we can time the client out.

Notice how the `Allocator` and `LastPacketReceiveTime` fields have a `GCH` postfix, which represents [`GCHandle`](https://docs.microsoft.com/en-us/dotnet/api/system.runtime.interopservices.gchandle?view=net-5.0). In the contexts, the allocator performs read and write operations, which means the `Allocator` class instance should *not* be garbage collected while in use by unmanaged code. By using `GCHandle`, we can prevent the class instance from being collected by the garbage collector. This also applies to the `LastPacketReceiveTime` class, since we need to get and set the time of the last packet received for timeout purposes. However, we do not need to do this for read-only fields like `ClientData` since they will not be garbage collected.

Moving past allocations for the moment, we set up a callback for our own custom logging function, but compared to the previous examples, we add colours and more informative messages:

```csharp
// Determines the log type from the level 
static string LogLevelString(int level)
{
    if (level == Next.NEXT_LOG_LEVEL_ERROR) {
        return "error";
    } else if (level == Next.NEXT_LOG_LEVEL_INFO) {
        return "info";
    } else if (level == Next.NEXT_LOG_LEVEL_WARN) {
        return "warn";
    } else if (level == Next.NEXT_LOG_LEVEL_DEBUG) {
        return "debug";
    } else {
        return "???";
    }
}

enum Color { red, green, blue, black, white, yellow, orange };

// Define custom logging function to output to Unity console
[MonoPInvokeCallback(typeof(NextLogFunction))]
static void UnityLogger(int level, IntPtr formatPtr, IntPtr argsPtr)
{
    // Unmarshal the log message into a string
    string argsStr = Marshal.PtrToStringAnsi(argsPtr);

    // Choose a colour for the log depending on the log level
    Color c;
    if (level == Next.NEXT_LOG_LEVEL_ERROR) {
        c = Color.red;
    } else if (level == Next.NEXT_LOG_LEVEL_INFO) {
        c = Color.green;
    } else if (level == Next.NEXT_LOG_LEVEL_WARN) {
        c = Color.yellow;
    } else if (level == Next.NEXT_LOG_LEVEL_DEBUG) {
        c = Color.orange;
    } else {
        c = Color.white;
    }

    if (level != Next.NEXT_LOG_LEVEL_NONE)
    {
        // Log to Unity console
        Debug.Log(String.Format("<color={0}>{1}: {2}: {3}</color>", 
            c.ToString(), Next.NextTime().ToString("F2"), LogLevelString(level), argsStr)
        );
    }
}
```

There are four different log levels in Network Next:

1. NEXT_LOG_LEVEL_NONE (0)
2. NEXT_LOG_LEVEL_ERROR (1)
3. NEXT_LOG_LEVEL_INFO (2)
4. NEXT_LOG_LEVEL_WARN (3)
5. NEXT_LOG_LEVEL_DEBUG (4)

The default log level is NEXT_LOG_LEVEL_INFO, which shows both info and error logs. This is a good default, as these messages are infrequent. Warnings can be more frequent, and aren't important enough to be errors, so are off by default. Debug logs are incredibly spammy and should only be turned on when debugging a specific issue in the Network Next SDK.

How you handle each of these log levels in the log function callback is up to you. We just pass them in, but depending on the log level we will not call the callback unless the level of the log is <= the current log level value set.

Finally, there is a small feature where a log with NEXT_LOG_LEVEL_NONE is used to indicate an unadorned regular printf. This is useful for console platforms like XBoxOne where hoops need to be jumped through just to get text printed to stdout. This is used by our unit tests and by the default assert handler function in the Network Next SDK. However, this only applies when a custom logging function is not set.

Next we define a custom assert handler:
```csharp
// Define custom assert function
[MonoPInvokeCallback(typeof(NextAssertFunction))]
static void AssertFunction(bool condition, string function, string file, int line)
{
    #if UNITY_EDITOR
        // Stops the editor cleanly
        Debug.LogError(String.Format("assert failed: ({0}), function {1}, file {2}, line {3}", condition, function, file, line));
        Assert.IsFalse(condition, String.Format("assert failed: ({0}), function {1}, file {2}, line {3}", condition, function, file, line));

        UnityEditor.EditorApplication.isPlaying = false;
    #else
        Application.Quit();
    #endif // #if UNITY_EDITOR 
}
```
Here we print out the assert message and force the editor to stop. Again, typically you would override this to point to your own assert handler in your game. The code above gracefully stops play mode when an assertion fails, or quits the game when not using the editor. Setting your own assert handler is wise, since relying on the default assert handler will cause the editor to crash most of the time.

Also, it's worth noting that `Next.NextAssert()` only runs when the `UNITY_ASSERTIONS` define is true, which is the default when running in the editor.

Now instead of sending zero byte packets, let's send some packets with real intent.
```csharp
// Generates a valid Network Next packet
static byte[] GeneratePacket(out int packetBytes) {
    var rand = new System.Random();

    packetBytes = 1 + (rand.Next() % Next.NEXT_MTU);

    byte[] packetData = new byte[packetBytes];

    int start = packetBytes % 256;
    for (int i = 0; i < packetBytes; i++) {
        packetData[i] = (byte)((start + i) % 256);
    }

    return packetData;
}

// Validates a packet is a Network Next packet
static bool VerifyPacket(byte[] packetData, int packetBytes) {
    int start = packetBytes % 256;
    for (int i = 0; i < packetBytes; i++) {
        if (packetData[i] != (byte)((start + i) % 256)) {
            String msg = String.Format("{0}: {1} != {2} ({3})", i, packetData[i], (start + i) % 256, packetBytes);
            Debug.LogError(msg);
            return false;
        }
    }

    return true;
}
```
The functions above generate packets of random length from 1 to the maximum size packet that can be sent across Network Next – NEXT_MTU (1300 bytes). These packets have contents that can be inferred by the size of the packet, making it possible for us to test a packet and with high probability, ensure that the packet has not been incorrectly truncated or padded, and that it contains the exact bytes sent.

Now we are ready to set a custom log level, set our custom log function, allocators and assert handler. 

Before initializing Network Next, do this:
```csharp
// Allow all logging messages to be displayed
Next.NextLogLevel(Next.NEXT_LOG_LEVEL_DEBUG);

// Assign our custom logging function
Next.NextLogFunction(UnityLogger);

// Assign our custom assert function
Next.NextAssertFunction(AssertFunction);

// Assign our custom allocation functions
Next.NextAllocator(MallocFunction, FreeFunction);
```

Next, create a global context and pass it in to `Next.NextInit()` to be used for any global allocations made by the Network Next SDK:
```csharp
// Create a global context for any global allocations, using GCHandle for read/write fields
Context globalCtx = new Context();
Allocator globalCtxAllocator = new Allocator();
GCHandle globalCtxAllocatorGCH = GCHandle.Alloc(globalCtxAllocator);
globalCtx.AllocatorGCH = GCHandle.ToIntPtr(globalCtxAllocatorGCH);  

// Marshal the global context into a pointer
globalCtxPtr = Marshal.AllocHGlobal(Marshal.SizeOf(globalCtx));
Marshal.StructureToPtr(globalCtx, globalCtxPtr, false);

// Get the default configuration
Next.NextConfig config;
Next.NextDefaultConfig(out config);

// Assign our public key to the configuration
config.CustomerPublicKey = customerPublicKey;

// Initialize Network Next
if (Next.NextInit(globalCtxPtr, ref config) != Next.NEXT_OK)
{
    Debug.LogError("error: could not initialize network next");
    this.gameObject.SetActive(false);
    return;
}
```
Recall how we must use `GCHandle` to keep the `globalCtxAllocator` class instance alive so that it can be unmarshaled and accessed in our malloc and free functions.

Next, create a per-client context and pass it in as the context when creating the client:
```csharp
// Create client context, using GCHandle for read/write fields
ClientContext clientCtx = new ClientContext();

Allocator clientCtxAllocator = new Allocator();
GCHandle clientCtxAllocatorGCH = GCHandle.Alloc(clientCtxAllocator);
clientCtx.AllocatorGCH = GCHandle.ToIntPtr(clientCtxAllocatorGCH);

clientCtx.ClientData = 0x12345;

LastPacketReceiveTime clientCtxLastPacketReceiveTime = new LastPacketReceiveTime(Next.NextTime());
GCHandle clientCtxLastPacketReceiveTimeGCH = GCHandle.Alloc(clientCtxLastPacketReceiveTime);
clientCtx.LastPacketReceiveTimeGCH = GCHandle.ToIntPtr(clientCtxLastPacketReceiveTimeGCH);

// Marshal the client context into a pointer
clientCtxPtr = Marshal.AllocHGlobal(Marshal.SizeOf(clientCtx));
Marshal.StructureToPtr(clientCtx, clientCtxPtr, false);

// Create the packet received callback
NextClientPacketReceivedCallback recvCallBack = new NextClientPacketReceivedCallback(ClientPacketReceived);

// Create a pointer to the client (store as global var)
client = Next.NextClientCreate(clientCtxPtr, bindAddress, recvCallBack, null);
if (client == IntPtr.Zero)
{
    Debug.LogError("error: failed to create client");
    this.gameObject.SetActive(false);
    return;
}
```
Now when the client makes any allocations, and when it calls callbacks like `ClientPacketReceived()` it will pass in the client context as an `IntPtr` that we can unmarshal.

Since we are binding the client to port 0, the system will choose the actual port number. We can retrieve this port number as follows and print it out for posterity:
```csharp
// Log the client port
ushort clientPort = Next.NextClientPort(client);
Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, "client port is ", clientPort.ToString());
```        

Finally, the client has been extended to print out all the useful stats you can retrieve from a network next client, once every ten seconds:
```csharp
// Update is called once per frame
void Update()
{
    Next.NextClientUpdate(client);

    // ...

    // Create a packet to send to the server
    int packetBytes;
    byte[] packetData = GeneratePacket(out packetBytes);

    // Send the packet to the server
    Next.NextClientSendPacket(client, packetData, packetBytes);

    // ...

    accumulator += deltaTime;

    if (accumulator > 10.0)
    {
        PrintClientStats(client);
        accumulator = 0.0;
    }
    
    // ...
}

static void PrintClientStats(IntPtr client)
{
    // ...

    StringBuilder sb = new StringBuilder("================================================================\n");

    Next.ClientStats stats = Next.NextClientStats(client);

    string platform = "unknown";
    switch (stats.PlatformID)
    {
        case Next.NEXT_PLATFORM_WINDOWS:
            platform = "windows";
            break;

        case Next.NEXT_PLATFORM_MAC:
            platform = "mac";
            break;

        case Next.NEXT_PLATFORM_LINUX:
            platform = "linux";
            break;

        case Next.NEXT_PLATFORM_SWITCH:
            platform = "nintendo switch";
            break;

        case Next.NEXT_PLATFORM_PS4:
            platform = "ps4";
            break;

        case Next.NEXT_PLATFORM_PS5:
            platform = "ps5";
            break;

        case Next.NEXT_PLATFORM_IOS:
            platform = "ios";
            break;

        case Next.NEXT_PLATFORM_XBOX_ONE:
            platform = "xbox one";
            break;

        case Next.NEXT_PLATFORM_XBOX_SERIES_X:
            platform = "xbox series x";
            break;

        default:
            break;
    }

    string stateStr = "???";
    int state = Next.NextClientState(client);
    switch (state)
    {
        case Next.NEXT_CLIENT_STATE_CLOSED:
            stateStr = "closed";
            break;

        case Next.NEXT_CLIENT_STATE_OPEN:
            stateStr = "open";
            break;

        case Next.NEXT_CLIENT_STATE_ERROR:
            stateStr = "error";
            break;

        default:
            break;
    }

    sb.AppendFormat("state = {0} ({1})\n", stateStr, state);
    sb.AppendFormat("session id = {0}\n", Next.NextClientSessionID(client));
    sb.AppendFormat("platform id = {0} ({1})\n", platform, (int)stats.PlatformID);

    string connection = "unknown";

    switch (stats.ConnectionType)
    {
        case Next.NEXT_CONNECTION_TYPE_WIRED:
            connection = "wired";
            break;

        case Next.NEXT_CONNECTION_TYPE_WIFI:
            connection = "wifi";
            break;

        case Next.NEXT_CONNECTION_TYPE_CELLULAR:
            connection = "cellular";
            break;

        default:
            break;
    }

    sb.AppendFormat("connection type = {0} ({1})\n", connection, stats.ConnectionType);

    if (!stats.FallbackToDirect)
    {
        sb.AppendFormat("upgraded = {0}\n", stats.Upgraded.ToString());
        sb.AppendFormat("committed = {0}\n", stats.Committed.ToString());
        sb.AppendFormat("multipath = {0}\n", stats.Multipath.ToString());
        sb.AppendFormat("reported = {0}\n", stats.Reported.ToString());
    }

    sb.AppendFormat("fallback to direct = {0}\n", stats.FallbackToDirect.ToString());
    
    sb.AppendFormat("high frequency pings = {0}\n", stats.HighFrequencyPings.ToString());

    sb.AppendFormat("direct rtt = {0}ms\n", stats.DirectRTT.ToString("F"));
    sb.AppendFormat("direct jitter = {0}ms\n", stats.DirectJitter.ToString("F"));
    sb.AppendFormat("direct packet loss = {0}%\n", stats.DirectPacketLoss.ToString("F1"));

    if (stats.Next)
    {
        sb.AppendFormat("next rtt = {0}ms\n", stats.NextRTT.ToString("F"));
        sb.AppendFormat("next jitter = {0}ms\n", stats.NextJitter.ToString("F"));
        sb.AppendFormat("next packet loss = {0}%\n", stats.NextPacketLoss.ToString("F1"));
        sb.AppendFormat("next bandwidth up = {0}kbps\n", stats.NextKbpsUp.ToString("F1"));
        sb.AppendFormat("next bandwidth down = {0}kbps\n", stats.NextKbpsDown.ToString("F1"));
    }

    if (stats.Upgraded && !stats.FallbackToDirect)
    {
        sb.AppendFormat("packets sent client to server = {0}\n", stats.PacketsSentClientToServer);
        sb.AppendFormat("packets sent server to client = {0}\n", stats.PacketsSentServerToClient);
        sb.AppendFormat("packets lost client to server = {0}\n", stats.PacketsLostClientToServer);
        sb.AppendFormat("packets lost server to client = {0}\n", stats.PacketsLostServerToClient);
        sb.AppendFormat("packets out of order client to server = {0}\n", stats.PacketsOutOfOrderClientToServer);
        sb.AppendFormat("packets out of order server to client = {0}\n", stats.PacketsOutOfOrderServerToClient);
        sb.AppendFormat("jitter client to server = {0}\n", stats.JitterClientToServer.ToString("F"));
        sb.AppendFormat("jitter server to client = {0}\n", stats.JitterServerToClient.ToString("F"));
    }

    sb.AppendFormat("================================================================\n");

    Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, sb.ToString());
}
```

When you have finished your session with the server, close it:
```csharp
// Close the session
Next.NextClientCloseSession(client);
```

When you have finished using your client, destroy it and free the memory allocated for all contexts (Unity's `Destroy()` function is a good place to do this):
```csharp
// Destroy the client
Next.NextClientDestroy(client);

// Free the unmanaged memory from the context's allocators and context itself
ClientContext clientCtx = (ClientContext)Marshal.PtrToStructure(clientCtxPtr, typeof(ClientContext));
GCHandle clientCtxAllocatorGCH = GCHandle.FromIntPtr(clientCtx.AllocatorGCH);
clientCtxAllocatorGCH.Free();
GCHandle lastPacketReceiveTimeGCH = GCHandle.FromIntPtr(clientCtx.LastPacketReceiveTimeGCH);
lastPacketReceiveTimeGCH.Free();
Marshal.FreeHGlobal(clientCtxPtr);

Context globalCtx = (Context)Marshal.PtrToStructure(globalCtxPtr, typeof(Context));
GCHandle globalCtxAllocatorGCH = GCHandle.FromIntPtr(globalCtx.AllocatorGCH);
globalCtxAllocatorGCH.Free();
Marshal.FreeHGlobal(globalCtxPtr);
```
Notice how we free the client context before the global context, and that the allocator is always freed prior to the contexts themselves.

Finally, before your application terminates, please shut down the SDK (Unity's `OnApplicationQuit()` is an appropriate place to do this):
```csharp
Next.NextTerm();
```
