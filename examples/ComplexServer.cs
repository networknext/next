using System;
using System.Runtime.InteropServices;
using System.Collections.Generic;
using System.Text;
using UnityEngine;
using UnityEngine.Assertions;
using AOT;
using NetworkNext;

namespace Server {

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

	public class ComplexServer : MonoBehaviour
	{
		// Constants
		const string bindAddress = "0.0.0.0:50000";
		const string serverAddress = "127.0.0.1:50000";
		const string serverDatacenter = "local";
		const string serverBackendHostname = "prod.spacecats.net";
		const string customerPrivateKey = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"; // Replace with the private key from your account
		const double deltaTime = 0.25;

		// ----------------------------------------------------------

		// Enums and Structs

		enum Color { red, green, blue, black, white, yellow, orange };

		[StructLayout (LayoutKind.Sequential)]
		public struct Context
		{
			public IntPtr AllocatorGCH;
		}

		[StructLayout (LayoutKind.Sequential)]
		public struct ServerContext
		{
			public IntPtr AllocatorGCH;
			public uint ServerData;
			public IntPtr ClientDataMapGCH;
		}

		// ----------------------------------------------------------

		// Global variables
		IntPtr server;
		IntPtr globalCtxPtr;
		IntPtr serverCtxPtr;

		double accumulator = 0.0;

		// ----------------------------------------------------------

		// Delegate functions

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

		// Define packet receive callback function
		[MonoPInvokeCallback(typeof(NextServerPacketReceivedCallback))]
		public void ServerPacketReceived(IntPtr serverPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
		{
			// Unmarshal the context pointer into the client context to access its fields
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

		// ----------------------------------------------------------

		// Utility functions

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

		static void PrintServerStats(IntPtr serverPtr, ClientDataMap clientDataMap)
		{
			Next.NextAssert(serverPtr != IntPtr.Zero);
			Next.NextAssert(clientDataMap != null);

			Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("{0} connected clients", clientDataMap.GetNumClients()));

			bool showDetailedStats = true;

			if (!showDetailedStats)
			{
				return;
			}

			StringBuilder sb = new StringBuilder();

			Next.NextAddress[] addresses = clientDataMap.GetClientAddresses();

			for (int i = 0; i < addresses.Length; i++)
			{
				Next.NextAddress address = addresses[i];

				// Create IntPtr for each address
				IntPtr addrPtr = Marshal.AllocHGlobal(Marshal.SizeOf(address));
				Marshal.StructureToPtr(address, addrPtr, false);

				// Get the stats for this address
				Next.ServerStats stats;
				Next.NextServerStats(serverPtr, addrPtr, out stats);

				// Release memory for the address pointer
				Marshal.FreeHGlobal(addrPtr);

				sb.Append("================================================================\n");
				sb.AppendFormat("address = {0}\n", Next.NextAddressToString(ref address));

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

				sb.AppendFormat("session id = {0}\n", stats.SessionID.ToString());
				sb.AppendFormat("platform id = {0} ({1})\n", platform, stats.PlatformID);

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
					sb.AppendFormat("committed = {0}\n", stats.Committed.ToString());
					sb.AppendFormat("multipath = {0}\n", stats.Multipath.ToString());
					sb.AppendFormat("reported = {0}\n", stats.Reported.ToString());
				}

				sb.AppendFormat("fallback to direct = {0}\n", stats.FallbackToDirect.ToString());

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

				if (!stats.FallbackToDirect)
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
				
				if (stats.NumTags > 0)
				{
					sb.Append("tags = [");
					for (int j = 0; j < stats.NumTags; j++)
					{
						if (j < stats.NumTags - 1)
						{
							sb.AppendFormat("{0},", stats.Tags[j].ToString());
						}
						else
						{
							sb.AppendFormat("{0}", stats.Tags[j].ToString());
						}
					}
					sb.Append("]\n");
				}

				sb.Append("================================================================\n");
			}

			if (sb.Length > 0)
			{
				Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, sb.ToString());
			}
		}

		// ----------------------------------------------------------

		// Start is called before the first frame update
		void Start()
		{
			// Allow all logging messages to be displayed
			Next.NextLogLevel(Next.NEXT_LOG_LEVEL_DEBUG);

			// Assign our custom logging function
			Next.NextLogFunction(UnityLogger);

			// Assign our custom assert function
			Next.NextAssertFunction(AssertFunction);

			// Assign our custom allocation functions
			Next.NextAllocator(MallocFunction, FreeFunction);

			// Get the default configuration
			Next.NextConfig config;
			Next.NextDefaultConfig(out config);

			// Assign our private key and the server backend hostname to the configuration
			config.CustomerPrivateKey = customerPrivateKey;
			config.ServerBackendHostname = serverBackendHostname;

			// Create a global context for any global allocations, using GCHandle for read/write fields
			Context globalCtx = new Context();
			Allocator globalCtxAllocator = new Allocator();
			GCHandle globalCtxAllocatorGCH = GCHandle.Alloc(globalCtxAllocator);
			globalCtx.AllocatorGCH = GCHandle.ToIntPtr(globalCtxAllocatorGCH);  

			// Marshal the global context into a pointer
			globalCtxPtr = Marshal.AllocHGlobal(Marshal.SizeOf(globalCtx));
			Marshal.StructureToPtr(globalCtx, globalCtxPtr, false);

			// Initialize Network Next
			if (Next.NextInit(globalCtxPtr, ref config) != Next.NEXT_OK)
			{
				Debug.LogError("error: could not initialize network next");
				this.gameObject.SetActive(false);
				return;
			}

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

			ushort serverPort = Next.NextServerPort(server);

			Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server port is {0}", serverPort.ToString()));
		}

		// Update is called once per frame
		void Update()
		{
			Next.NextServerUpdate(server);

			// Unmarshal the context pointer into the server context to access client data map
			ServerContext ctx = (ServerContext)Marshal.PtrToStructure(serverCtxPtr, typeof(ServerContext));

			Next.NextAssert(!ctx.Equals(default(ServerContext)));
			
			GCHandle clientDataMapGCH = GCHandle.FromIntPtr(ctx.ClientDataMapGCH);
			ClientDataMap clientDataMap = (ClientDataMap)clientDataMapGCH.Target;
			
			clientDataMap.UpdateClientTimeouts(Next.NextTime());

			accumulator += deltaTime;

			if (accumulator > 10.0)
			{
				PrintServerStats(server, clientDataMap);
				accumulator = 0.0;
			}

			Next.NextSleep(deltaTime);
		}

		// OnApplicationQuit is called when the application quits or when playmode is stopped in the editor
		// These actions should be done in Destroy() rather than when the application quits
		void OnApplicationQuit()
		{
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

			// Shut down the SDK
			Next.NextTerm();
		}
	}
}
