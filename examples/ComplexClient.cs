using System;
using System.Runtime.InteropServices;
using System.Text;
using UnityEngine;
using UnityEngine.Assertions;
using AOT;
using NetworkNext;

namespace Client {

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
	public class LastPacketReceiveTime
	{
		public double time;

		public LastPacketReceiveTime(double startTime)
		{
			time = startTime;
		}

		public double GetTime()
		{
			return time;
		}

		public void SetTime(double newTime)
		{
			time = newTime;
		}
	}

	public class ComplexClient : MonoBehaviour
	{
		// Constants
		const string bindAddress = "0.0.0.0:0";
		const string serverAddress = "127.0.0.1:50000";
		const string customerPublicKey = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="; // Replace with the public key from your account
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
		public struct ClientContext
		{
			public IntPtr AllocatorGCH;
			public uint ClientData;
			public IntPtr LastPacketReceiveTimeGCH;
		}

		// ----------------------------------------------------------

		// Global variables
		
		IntPtr client;
		IntPtr globalCtxPtr;
		IntPtr clientCtxPtr;

		double accumulator = 0.0;
		bool reported = false;

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

		// ----------------------------------------------------------

		// Utility functions

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

		static void PrintClientStats(IntPtr client)
		{
			bool showDetailedStats = true;
			if (!showDetailedStats)
			{
				return;
			}

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

		// Updates the client timeout
		static void UpdateClientTimeout(IntPtr ctxPtr)
		{
			// Unmarshal the context pointer into the client context to access its fields
			ClientContext ctx = (ClientContext)Marshal.PtrToStructure(ctxPtr, typeof(ClientContext));

			Next.NextAssert(!ctx.Equals(default(ClientContext)));

			GCHandle lastPacketReceiveTimeGCH = GCHandle.FromIntPtr(ctx.LastPacketReceiveTimeGCH);
			LastPacketReceiveTime lastPacketReceiveTime = (LastPacketReceiveTime)lastPacketReceiveTimeGCH.Target;

			if (lastPacketReceiveTime.GetTime() + 5.0 < Next.NextTime())
			{
				Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, "client connection timed out");

				#if UNITY_EDITOR
					UnityEditor.EditorApplication.isPlaying = false;
				#else
					Application.Quit();
				#endif // #if UNITY_EDITOR
			}
		}

		// ----------------------------------------------------------

		// Unity GameObject functions

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

			// Log the client port
			ushort clientPort = Next.NextClientPort(client);
			Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, "client port is ", clientPort.ToString());

			// Open a session to the server
			Next.NextClientOpenSession(client, serverAddress);
		}

		// Update is called once per frame
		void Update()
		{
			Next.NextClientUpdate(client);

			int clientState = Next.NextClientState(client);
			if (clientState == Next.NEXT_CLIENT_STATE_ERROR)
			{
				Debug.LogError("error: client is in an error state");
				this.gameObject.SetActive(false);
				return;
			}

			// Create a packet to send to the server
			int packetBytes;
			byte[] packetData = GeneratePacket(out packetBytes);

			// Send the packet to the server
			Next.NextClientSendPacket(client, packetData, packetBytes);

			if (Next.NextTime() > 60.0 && !reported)
			{
				Next.NextClientReportSession(client);
				reported = true;
			}

			accumulator += deltaTime;

			if (accumulator > 10.0)
			{
				PrintClientStats(client);
				accumulator = 0.0;
			}

			UpdateClientTimeout(clientCtxPtr);

			Next.NextSleep(deltaTime);
		}

		// OnApplicationQuit is called when the application quits or when playmode is stopped in the editor
		// These actions should be done in Destroy() rather than when the application quits
		void OnApplicationQuit()
		{
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

			// Shut down the SDK
			Next.NextTerm();
		}
	}
}
