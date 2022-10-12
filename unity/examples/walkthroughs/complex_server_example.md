# Complex Server Example

In this example we build the kitchen sink version of a server in Unity where we show off all the features :)

We demonstrate:

- Setting the Network Next log level
- Setting a custom log function
- Setting a custom assert handler
- Setting a custom allocator

In this example, everything is as per the complex client example: setting up the allocator, a global context, override functions for malloc and free, custom log function and a custom assert function.

When creating a server, we create it with a server context as follows:
```csharp
[StructLayout (LayoutKind.Sequential)]
public struct ServerContext
{
    public IntPtr AllocatorGCH;
    public uint ServerData;
    public IntPtr ClientDataMapGCH;
}
```

It uses the same `AllocatorGCH` as in the complex client example, but this time it also includes `ServerData`, a read only field, and `ClientDataMapGCH`, a `IntPtr` to a `GCHandle` of the `ClientDataMap` class, which is responsible for tracking ongoing client sessions.
```csharp
[StructLayout (LayoutKind.Sequential)]
public struct ClientData
{
    public ulong SessionID;
    public Next.NextAddress Address;
    public double LastPacketReceiveTime;
}

// Mimics a map using an array, easy to store as a pointer in context
[StructLayout (LayoutKind.Sequential)]
public class ClientDataMap
{
    // Constants
    public const int MAX_CLIENTS = 512;

    // Global vars
    public Next.NextMutex mutex;
    public int numClients;
    [MarshalAs(UnmanagedType.ByValArray, SizeConst=MAX_CLIENTS)]
    public IntPtr[] clientMap;

    public ClientDataMap()
    {
        // Intialize globals
        int result = Next.NextMutexCreate(out mutex);
        Next.NextAssert(result == Next.NEXT_OK);
        numClients = 0;
        clientMap = new IntPtr[MAX_CLIENTS];
    }

    ~ClientDataMap()
    {
        // Cleanup
        Next.NextMutexDestroy(ref mutex);

        foreach (IntPtr dataPtr in clientMap)
        {
            if (dataPtr != IntPtr.Zero && dataPtr != null)
            {
                Marshal.FreeHGlobal(dataPtr);
                numClients--;
            }
        }

        Next.NextAssert(numClients == 0);
        Next.NextAssert(Array.Exists(clientMap, element => element == null || element.Equals(IntPtr.Zero)));

    }

    public bool IsExistingSession(ref Next.NextAddress address)
    {
        Next.NextMutexAcquire(ref mutex);

        foreach (IntPtr dataPtr in clientMap)
        {
            if (dataPtr != IntPtr.Zero && dataPtr != null)
            {
                ClientData data = (ClientData)Marshal.PtrToStructure(dataPtr, typeof(ClientData));

                if (Next.NextAddressEqual(ref address, ref data.Address))
                {
                    Next.NextMutexRelease(ref mutex);
                    return true;
                }
            }
        }

        Next.NextMutexRelease(ref mutex);
        return false;
    }

    public bool AddNewSession(ClientData data)
    {
        Next.NextMutexAcquire(ref mutex);

        if (numClients >= MAX_CLIENTS)
        {
            Next.NextMutexRelease(ref mutex);

            Next.NextPrintf(Next.NEXT_LOG_LEVEL_WARN, String.Format("server has reached max number of sessions ({0}) could not add {1} [{2}]", MAX_CLIENTS, Next.NextAddressToString(ref data.Address), data.SessionID.ToString()));
            return false;
        }

        for (int i = 0; i < clientMap.Length; i++)
        {
            if (clientMap[i] == IntPtr.Zero || clientMap[i] == null)
            {
                // Create a new pointer with the data and store it in the array
                IntPtr dataPtr = Marshal.AllocHGlobal(Marshal.SizeOf(data));
                Marshal.StructureToPtr(data, dataPtr, false);
                clientMap[i] = dataPtr;
                numClients++;

                Next.NextMutexRelease(ref mutex);
                return true;
            }
        }

        Next.NextMutexRelease(ref mutex);

        Next.NextPrintf(Next.NEXT_LOG_LEVEL_ERROR, String.Format("server could not add new session for {0} [{1}]", Next.NextAddressToString(ref data.Address), data.SessionID.ToString()));
        return false;
    }

    public bool RemoveSession(ref Next.NextAddress address)
    {
        Next.NextMutexAcquire(ref mutex);

        for (int i = 0; i < clientMap.Length; i++)
        {
            if (clientMap[i] != IntPtr.Zero && clientMap[i] != null)
            {
                ClientData data = (ClientData)Marshal.PtrToStructure(clientMap[i], typeof(ClientData));

                if (Next.NextAddressEqual(ref address, ref data.Address))
                {
                    // Free the pointer and set it to IntPtr.Zero
                    Marshal.FreeHGlobal(clientMap[i]);
                    clientMap[i] = IntPtr.Zero;
                    numClients--;

                    Next.NextMutexRelease(ref mutex);
                    return true;
                }
            }
        }

        Next.NextMutexRelease(ref mutex);

        Next.NextPrintf(Next.NEXT_LOG_LEVEL_ERROR, String.Format("server could not find session to remove: {0}", Next.NextAddressToString(ref address)));
        return false;
    }

    public bool UpdateLastPacketReceiveTime(ref Next.NextAddress address, double lastPacketReceiveTime)
    {
        Next.NextMutexAcquire(ref mutex);

        for (int i = 0; i < clientMap.Length; i++)
        {
            if (clientMap[i] != IntPtr.Zero && clientMap[i] != null)
            {
                ClientData data = (ClientData)Marshal.PtrToStructure(clientMap[i], typeof(ClientData));

                if (Next.NextAddressEqual(ref address, ref data.Address))
                {
                    // Update the time
                    data.LastPacketReceiveTime = lastPacketReceiveTime;

                    // Free the previous pointer
                    Marshal.FreeHGlobal(clientMap[i]);

                    // Create a new pointer with the latest time
                    IntPtr dataPtr = Marshal.AllocHGlobal(Marshal.SizeOf(data));
                    Marshal.StructureToPtr(data, dataPtr, false);
                    clientMap[i] = dataPtr;

                    Next.NextMutexRelease(ref mutex);
                    return true;
                }
            }
        }

        Next.NextMutexRelease(ref mutex);

        Next.NextPrintf(Next.NEXT_LOG_LEVEL_ERROR, String.Format("server could not find client address in map: {0}", Next.NextAddressToString(ref address)));
        return false;
    }

    public void UpdateClientTimeouts(double currentTime)
    {
        Next.NextMutexAcquire(ref mutex);

        for (int i = 0; i < clientMap.Length; i++)
        {
            if (clientMap[i] != IntPtr.Zero && clientMap[i] != null)
            {
                ClientData data = (ClientData)Marshal.PtrToStructure(clientMap[i], typeof(ClientData));

                if (data.LastPacketReceiveTime + 5.0 < currentTime)
                {
                    Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("client disconnected: {0} [{1}]", Next.NextAddressToString(ref data.Address), data.SessionID.ToString()));

                    // Free the pointer and set it to IntPtr.Zero
                    Marshal.FreeHGlobal(clientMap[i]);
                    clientMap[i] = IntPtr.Zero;
                    numClients--;
                }
            }
        }

        Next.NextMutexRelease(ref mutex);
    }

    public int GetNumClients()
    {
        int clients;

        Next.NextMutexAcquire(ref mutex);
        clients = numClients;
        Next.NextMutexRelease(ref mutex);

        return clients;
    }

    public Next.NextAddress[] GetClientAddresses()
    {
        List<Next.NextAddress> addresses = new List<Next.NextAddress>();

        Next.NextMutexAcquire(ref mutex);

        for (int i = 0; i < clientMap.Length; i++)
        {
            if (clientMap[i] != IntPtr.Zero && clientMap[i] != null)
            {
                ClientData data = (ClientData)Marshal.PtrToStructure(clientMap[i], typeof(ClientData));

                addresses.Add(data.Address);
            }
        }

        Next.NextMutexRelease(ref mutex);

        return addresses.ToArray();
    }
}
```
We create this convenience class to monitor sessions since we cannot marshal a generic hash map using P/Invoke. We also must ensure that the data types of all instance variables are compatible with P/Invoke interop marshaling, which is why we opt to use an array of fixed size to hold pointers to session data.

Now, we can create a server using the server with a server context:
```csharp
// Create server context, using GCHandle for read/write fields
ServerContext serverCtx = new ServerContext();

Allocator serverCtxAllocator = new Allocator();
GCHandle serverCtxAllocatorGCH = GCHandle.Alloc(serverCtxAllocator);
serverCtx.AllocatorGCH = GCHandle.ToIntPtr(serverCtxAllocatorGCH);

serverCtx.ServerData = 0x12345678;

ClientDataMap serverCtxClientDataMap = new ClientDataMap();
GCHandle serverCtxClientDataMapGCH = GCHandle.Alloc(serverCtxClientDataMap);
serverCtx.ClientDataMapGCH = GCHandle.ToIntPtr(serverCtxClientDataMapGCH);

// Marshal the server context into a pointer
serverCtxPtr = Marshal.AllocHGlobal(Marshal.SizeOf(serverCtx));
Marshal.StructureToPtr(serverCtx, serverCtxPtr, false);

// Create the packet received callback
NextServerPacketReceivedCallback recvCallBack = new NextServerPacketReceivedCallback(ServerPacketReceived);

// Create a pointer to the server (store as global var)
server = Next.NextServerCreate(serverCtxPtr, serverAddress, bindAddress, serverDatacenter, recvCallBack, null);
if (server == IntPtr.Zero)
{
    Debug.LogError("error: failed to create server");
    this.gameObject.SetActive(false);
    return;
}
```

The overridden malloc and free functions are now called with the server context including our custom allocator:
```csharp
// Define custom malloc function
[MonoPInvokeCallback(typeof(NextMallocFunction))]
static IntPtr MallocFunction(IntPtr ctxPtr, ulong bytes)
{
    Context ctx = (Context)Marshal.PtrToStructure(ctxPtr, typeof(Context));

    Next.NextAssert(!ctx.Equals(default(Context)));

    GCHandle allocatorGCH = GCHandle.FromIntPtr(ctx.AllocatorGCH);
    Allocator allocator = (Allocator)allocatorGCH.Target;

    Next.NextAssert(allocator != null);

    return allocator.Alloc((int)bytes);
}

// Define custom free function
[MonoPInvokeCallback(typeof(NextFreeFunction))]
static void FreeFunction(IntPtr ctxPtr, IntPtr p)
{
    Context ctx = (Context)Marshal.PtrToStructure(ctxPtr, typeof(Context));

    Next.NextAssert(!ctx.Equals(default(Context)));

    GCHandle allocatorGCH = GCHandle.FromIntPtr(ctx.AllocatorGCH);
    Allocator allocator = (Allocator)allocatorGCH.Target;

    Next.NextAssert(allocator != null);

    allocator.Free(p);
}
```

And the packet received callback with the server context, allowing you to get a pointer to your own internal server data structure passed in to the packet received callback:
```csharp
// Define packet receive callback function
[MonoPInvokeCallback(typeof(NextServerPacketReceivedCallback))]
public void ServerPacketReceived(IntPtr serverPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
{
    // Unmarshal the context pointer into the server context to access its fields
    ServerContext ctx = (ServerContext)Marshal.PtrToStructure(ctxPtr, typeof(ServerContext));

    Next.NextAssert(!ctx.Equals(default(ServerContext)));

    GCHandle allocatorGCH = GCHandle.FromIntPtr(ctx.AllocatorGCH);
    Allocator allocator = (Allocator)allocatorGCH.Target;

    Next.NextAssert(allocator != null);
    Next.NextAssert(ctx.ServerData == 0x12345678);

    // Unmarshal the packet data into byte[]
    byte[] packetData = new byte[packetBytes];
    Marshal.Copy(packetDataPtr, packetData, 0, packetBytes);

    Next.NextServerSendPacket(serverPtr, fromPtr, packetData, packetBytes);

    Next.NextAddress fromAddress = Next.GetNextAddressFromPointer(fromPtr);

    GCHandle clientDataMapGCH = GCHandle.FromIntPtr(ctx.ClientDataMapGCH);
    ClientDataMap clientDataMap = (ClientDataMap)clientDataMapGCH.Target;

    if (clientDataMap.IsExistingSession(ref fromAddress))
    {
        // Update last packet receive time
        clientDataMap.UpdateLastPacketReceiveTime(ref fromAddress, Next.NextTime());
    }
    else
    {
        // Create the client data for the new session
        string userID = "user id can be any id that is unique across all users. we hash it before sending up to our backend";
        ulong sessionID = Next.NextServerUpgradeSession(serverPtr, fromPtr, userID);

        ClientData clientData = new ClientData();
        clientData.Address = fromAddress;
        clientData.SessionID = sessionID;
        clientData.LastPacketReceiveTime = Next.NextTime();

        if (clientDataMap.AddNewSession(clientData))
        {
            Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("client connected {0} [{1}]", Next.NextAddressToString(ref fromAddress), sessionID.ToString()));

            if (sessionID != 0)
            {
                string[] tags = new string[]{"pro", "streamer"};
                int numTags = 2;
                Next.NextServerTagSessionMultiple(serverPtr, fromPtr, tags, numTags);
            }
        }
    }
}
```

When you have finished using your server, flush and destroy it and free the memory allocated for all contexts (Unity's `OnDestroy()` function is a good place to do this):
```csharp
// Flush the server
Next.NextServerFlush(server);

// Destroy the server
Next.NextServerDestroy(server);

// Free the unmanaged memory from the context's fields and context iteself
ServerContext serverCtx = (ServerContext)Marshal.PtrToStructure(serverCtxPtr, typeof(ServerContext));
GCHandle clientDataMapGCH = GCHandle.FromIntPtr(serverCtx.ClientDataMapGCH);
clientDataMapGCH.Free();
GCHandle serverCtxAllocatorGCH = GCHandle.FromIntPtr(serverCtx.AllocatorGCH);
serverCtxAllocatorGCH.Free();
Marshal.FreeHGlobal(serverCtxPtr);

Context globalCtx = (Context)Marshal.PtrToStructure(globalCtxPtr, typeof(Context));
GCHandle globalCtxAllocatorGCH = GCHandle.FromIntPtr(globalCtx.AllocatorGCH);
globalCtxAllocatorGCH.Free();
Marshal.FreeHGlobal(globalCtxPtr);
```
Notice how we free the server context before the global context, and that the allocator is always freed prior to the contexts themselves.

Finally, before your application terminates, please shut down the SDK (Unity's `OnApplicationQuit()` is an appropriate place to do this):
```csharp
Next.NextTerm();
```
