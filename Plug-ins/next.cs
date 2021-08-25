#define NEXT_EXPERIMENTAL

using System;
using System.Runtime.InteropServices;
using System.Runtime.CompilerServices;
using System.Text;
using System.Collections.Generic;
using AOT;

namespace NetworkNext {

	// Network Next delegate functions

	public delegate void NextClientPacketReceivedCallback(IntPtr client, IntPtr ctx_ptr, IntPtr packet_data_ptr, int packet_bytes);

	public delegate void NextWakeupCallback(IntPtr ctx_ptr);

	public delegate void NextServerPacketReceivedCallback(IntPtr server, IntPtr ctx_ptr, IntPtr from_ptr, IntPtr packet_data_ptr, int packet_bytes);

	// C# does not support c-style formatting, making the format parameter useless
	// Instead, the args parameter will have the string to log
	public delegate void NextLogFunction(int level, IntPtr format, IntPtr args);

	public delegate void NextAssertFunction(bool condition, [MarshalAs(UnmanagedType.LPStr)] string function, [MarshalAs(UnmanagedType.LPStr)] string file, int line);

	public delegate IntPtr NextMallocFunction(IntPtr ctx_ptr, ulong bytes);

	public delegate void NextFreeFunction(IntPtr ctx_ptr, IntPtr p);

	// ----------------------------------------------------------

	// Network Next class
	public static class Next
	{
		#if UNITY_IOS
			const string dll = "__Internal";
		#else 
			const string dll = "next-unity";
		#endif // #if UNITY_IOS

		// ----------------------------------------------------------

		// Network Next constants

		public const int NEXT_OK = 0;
		public const int NEXT_ERROR = -1;

		public const int NEXT_MTU = 1300;
		public const int NEXT_ETHERNET_HEADER_BYTES = 18;
		public const int NEXT_IPV4_HEADER_BYTES = 20;
		public const int NEXT_UDP_HEADER_BYTES = 8;
		public const int NEXT_HEADER_BYTES = 34;

		public const int NEXT_LOG_LEVEL_NONE = 0;
		public const int NEXT_LOG_LEVEL_ERROR = 1;
		public const int NEXT_LOG_LEVEL_INFO = 2;
		public const int NEXT_LOG_LEVEL_WARN = 3;
		public const int NEXT_LOG_LEVEL_DEBUG = 4;

		public const int NEXT_ADDRESS_NONE = 0;
		public const int NEXT_ADDRESS_IPV4 = 1;
		public const int NEXT_ADDRESS_IPV6 = 2;

		public const int NEXT_MAX_ADDRESS_STRING_LENGTH = 256;

		public const int NEXT_CONNECTION_TYPE_UNKNOWN = 0;
		public const int NEXT_CONNECTION_TYPE_WIRED = 1;
		public const int NEXT_CONNECTION_TYPE_WIFI = 2;
		public const int NEXT_CONNECTION_TYPE_CELLULAR = 3;
		public const int NEXT_CONNECTION_TYPE_MAX = 3;

		public const int NEXT_PLATFORM_UNKNOWN = 0;
		public const int NEXT_PLATFORM_WINDOWS = 1;
		public const int NEXT_PLATFORM_MAC = 2;
		public const int NEXT_PLATFORM_LINUX = 3;
		public const int NEXT_PLATFORM_SWITCH = 4;
		public const int NEXT_PLATFORM_PS4 = 5;
		public const int NEXT_PLATFORM_IOS = 6;
		public const int NEXT_PLATFORM_XBOX_ONE = 7;
		public const int NEXT_PLATFORM_XBOX_SERIES_X = 8;
		public const int NEXT_PLATFORM_PS5 = 9;
		public const int NEXT_PLATFORM_GDK = 10;
		public const int NEXT_PLATFORM_MAX = 10;

		public const int NEXT_MAX_TAGS = 8;

		public const int NEXT_DEFAULT_SOCKET_SEND_BUFFER_SIZE = 1000000;
		public const int NEXT_DEFAULT_SOCKET_RECEIVE_BUFFER_SIZE = 1000000;

		public const int NEXT_CLIENT_STATE_CLOSED = 0;
		public const int NEXT_CLIENT_STATE_OPEN = 1;                               
		public const int NEXT_CLIENT_STATE_ERROR = 2;

		public const int NEXT_SERVER_STATE_DIRECT_ONLY = 0;
		public const int NEXT_SERVER_STATE_INITIALIZING = 1;
		public const int NEXT_SERVER_STATE_INITIALIZED = 2;

		public const int NEXT_MUTEX_BYTES = 256;

		// ----------------------------------------------------------
		
		// Network Next Wrapper Utility functions

		// Gets the IPv4 address as a byte[]
		public static byte[] GetNextAddressIPV4(NextAddress address)
		{
			byte[] addr = new byte[]{address.IPV4_0, address.IPV4_1, address.IPV4_2, address.IPV4_3};
			return addr;
		}

		// Gets the IPv6 address as a ushort[]
		public static ushort[] GetNextAddressIPV6(NextAddress address)
		{
			ushort[] addr = new ushort[]{address.IPV6_0, address.IPV6_1, address.IPV6_2, address.IPV6_3, address.IPV6_4, address.IPV6_5, address.IPV6_6, address.IPV6_7};
			return addr;
		}

		// Converts an IntPtr to a NextAddress struct 
		public static NextAddress GetNextAddressFromPointer(IntPtr nextAddressPointer)
		{
			return (NextAddress)Marshal.PtrToStructure(nextAddressPointer, typeof(NextAddress));
		}

		public static ClientStats GetNextClientStatsFromPointer(IntPtr clientStatsPointer)
		{
			return (ClientStats)Marshal.PtrToStructure(clientStatsPointer, typeof(ClientStats));
		}

		// ----------------------------------------------------------

		// Network Next Config
		[StructLayout (LayoutKind.Sequential, CharSet=CharSet.Ansi)]
		public struct NextConfig
		{
			[MarshalAs(UnmanagedType.ByValTStr, SizeConst = NEXT_MAX_ADDRESS_STRING_LENGTH)]
			public string ServerBackendHostname;
			[MarshalAs(UnmanagedType.ByValTStr, SizeConst = NEXT_MAX_ADDRESS_STRING_LENGTH)]
			public string PingBackendHostname;
			[MarshalAs(UnmanagedType.ByValTStr, SizeConst = NEXT_MAX_ADDRESS_STRING_LENGTH)]
			public string CustomerPublicKey;
			[MarshalAs(UnmanagedType.ByValTStr, SizeConst = NEXT_MAX_ADDRESS_STRING_LENGTH)]
			public string CustomerPrivateKey;
			public int SocketSendBufferSize;
			public int SocketReceiveBufferSize;
			[MarshalAs(UnmanagedType.U1)]
			public bool DisableNetworkNext;
		}

		// Network Next global functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_default_config")]
		private static extern int next_default_config(out NextConfig config);

		public static void NextDefaultConfig(out NextConfig config)
		{
			Next.next_default_config(out config);
		}
		
		// Context is an IntPtr because it is user defined, and thus needs to be marshaled and unmarshaled by the user
		// Reference: https://docs.microsoft.com/en-us/dotnet/api/system.runtime.interopservices.marshal.ptrtostructure?view=net-5.0#System_Runtime_InteropServices_Marshal_PtrToStructure_System_IntPtr_System_Type_
		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_init", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_init(IntPtr ctx, ref NextConfig config);

		public static int NextInit(IntPtr ctx, ref NextConfig config)
		{
			// Minimum socket send and receive buffer must be greater than 0
			if (config.SocketSendBufferSize <= 0)
			{
				config.SocketSendBufferSize = NEXT_DEFAULT_SOCKET_SEND_BUFFER_SIZE;
			} 

			if (config.SocketReceiveBufferSize <= 0)
			{
				config.SocketReceiveBufferSize = NEXT_DEFAULT_SOCKET_RECEIVE_BUFFER_SIZE;
			}

			return next_init(ctx, ref config);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_term")]
		private static extern void next_term();

		public static void NextTerm()
		{
			Next.next_term();
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_time", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern double next_time();

		public static double NextTime() {
			return Next.next_time();
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_sleep", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_sleep(double time_seconds);

		public static void NextSleep(double timeSeconds) {
			Next.next_sleep(timeSeconds);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_printf", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_printf(int level, [MarshalAs(UnmanagedType.LPStr)] string format);

		public static void NextPrintf(int level, string format, params Object[] args) {
			// Create the string to pass
			StringBuilder sb = new StringBuilder(format);

			foreach (Object o in args)
			{
				sb.AppendFormat("{0} ", o);
			}

			Next.next_printf(level, sb.ToString());
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "default_assert_function", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void default_assert_function([MarshalAs(UnmanagedType.LPStr)] string condition, [MarshalAs(UnmanagedType.LPStr)] string function, [MarshalAs(UnmanagedType.LPStr)] string file, int line);

		// Implement this in C# because we cannot export macro functions
		public static void NextAssert(bool condition, NextAssertFunction assertFunction = null, [CallerMemberName] string function = "", [CallerFilePath] string file = "", [CallerLineNumber] int line = 0) {
			#if UNITY_ASSERTIONS
				if (!condition && assertFunction != null)
				{
					// Use custom assert function callback
					assertFunction(condition, function, file, line);
				}
				else if (!condition)
				{
					// Use default assert function
					Next.default_assert_function(condition.ToString(), function, file, line);
				}
			#endif // #if UNITY_ASSERTIONS
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_quiet", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_quiet(bool flag);

		public static void NextQuiet(bool flag) {
			Next.next_quiet(flag);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_log_level", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_log_level(int level);

		public static void NextLogLevel(int level) {
			Next.next_log_level(level);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_log_function", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_log_function(NextLogFunction function);

		public static void NextLogFunction(NextLogFunction function) {
			Next.next_log_function(function);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_assert_function", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_assert_function(NextAssertFunction function);

		public static void NextAssertFunction(NextAssertFunction function) {
			Next.next_assert_function(function);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_allocator", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_allocator(NextMallocFunction malloc_function, NextFreeFunction free_function);

		public static void NextAllocator(NextMallocFunction malloc_function, NextFreeFunction free_function) {
			Next.next_allocator(malloc_function, free_function);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_user_id_string", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_user_id_string(ulong user_id, StringBuilder buffer);

		// Equivalent of String.Format("{0:x}", user_id);
		// Make sure the capacity is at least 16, otherwise string will be cut short
		public static string NextUserIDString(ulong user_id, int capacity) {
			StringBuilder buffer = new StringBuilder(capacity);
			next_user_id_string(user_id, buffer);
			return buffer.ToString();
		}
	
		// ----------------------------------------------------------

		// Network Next address
		[StructLayout (LayoutKind.Explicit)]
		public struct NextAddress
		{
			[FieldOffset(0)]
			public byte IPV4_0;
			[FieldOffset(1)]
			public byte IPV4_1;
			[FieldOffset(2)]
			public byte IPV4_2;
			[FieldOffset(3)]
			public byte IPV4_3;
			[FieldOffset(0)]
			public ushort IPV6_0;
			[FieldOffset(2)]
			public ushort IPV6_1;
			[FieldOffset(4)]
			public ushort IPV6_2;
			[FieldOffset(6)]
			public ushort IPV6_3;
			[FieldOffset(8)]
			public ushort IPV6_4;
			[FieldOffset(10)]
			public ushort IPV6_5;
			[FieldOffset(12)]
			public ushort IPV6_6;
			[FieldOffset(14)]
			public ushort IPV6_7;		
			[FieldOffset(16)]
			public ushort Port;
			[FieldOffset(18)]
			public byte Type;
		}

		// Network Next address functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_address_parse", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_address_parse(out NextAddress address, [MarshalAs(UnmanagedType.LPStr)] string address_string);

		public static int NextAddressParse(out NextAddress address, string address_string) {
			return Next.next_address_parse(out address, address_string);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_address_to_string", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_address_to_string(ref NextAddress address, StringBuilder buffer);

		public static string NextAddressToString(ref NextAddress address) {
			StringBuilder buffer = new StringBuilder(NEXT_MAX_ADDRESS_STRING_LENGTH);
			next_address_to_string(ref address, buffer);
			return buffer.ToString();
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_address_equal", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern bool next_address_equal(ref NextAddress a, ref NextAddress b);

		public static bool NextAddressEqual(ref NextAddress a, ref NextAddress b) {
			return Next.next_address_equal(ref a, ref b);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_address_anonymize", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_address_anonymize(ref NextAddress address);

		public static void NextAddressAnonymize(ref NextAddress address) {
			Next.next_address_anonymize(ref address);
		}

		// ----------------------------------------------------------

		// Network Next client stats
		[StructLayout (LayoutKind.Sequential)]
		public struct ClientStats
		{
			public int PlatformID;
			public int ConnectionType;
			[MarshalAs(UnmanagedType.U1)]
			public bool Next;
			[MarshalAs(UnmanagedType.U1)]
			public bool Upgraded;
			[MarshalAs(UnmanagedType.U1)]
			public bool Committed;
			[MarshalAs(UnmanagedType.U1)]
			public bool Multipath;
			[MarshalAs(UnmanagedType.U1)]
			public bool Reported;
			[MarshalAs(UnmanagedType.U1)]
			public bool FallbackToDirect;
			[MarshalAs(UnmanagedType.U1)]
			public bool HighFrequencyPings;
			public float DirectRTT;
			public float DirectJitter;
			public float DirectPacketLoss;
			public float NextRTT;
			public float NextJitter;
			public float NextPacketLoss;
			public float NextKbpsUp;
			public float NextKbpsDown;
			public ulong PacketsSentClientToServer;
			public ulong PacketsSentServerToClient;
			public ulong PacketsLostClientToServer;
			public ulong PacketsLostServerToClient;
			public ulong PacketsOutOfOrderClientToServer;
			public ulong PacketsOutOfOrderServerToClient;
			public float JitterClientToServer;
			public float JitterServerToClient;
		}

		// Network Next client functions

		// Context is an IntPtr because it is user defined, and thus needs to be marshaled and unmarshaled by the user
		// Reference: https://docs.microsoft.com/en-us/dotnet/api/system.runtime.interopservices.marshal.ptrtostructure?view=net-5.0#System_Runtime_InteropServices_Marshal_PtrToStructure_System_IntPtr_System_Type_
		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_create", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_client_create(IntPtr ctx, [MarshalAs(UnmanagedType.LPStr)] string bind_address, NextClientPacketReceivedCallback packet_received_callback, NextWakeupCallback wakeup_callback);

		public static IntPtr NextClientCreate(IntPtr ctx, string bind_address, NextClientPacketReceivedCallback packet_received_callback, NextWakeupCallback wakeup_callback)
		{
			return Next.next_client_create(ctx, bind_address, packet_received_callback, wakeup_callback);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_destroy", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_destroy(IntPtr client);

		public static void NextClientDestroy(IntPtr client)
		{
			Next.next_client_destroy(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_port", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern ushort next_client_port(IntPtr client);

		public static ushort NextClientPort(IntPtr client)
		{
			return Next.next_client_port(client);
		}
		
		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_open_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_open_session(IntPtr client, [MarshalAs(UnmanagedType.LPStr)] string server_address_string);

		public static void NextClientOpenSession(IntPtr client, string server_address_string)
		{
			Next.next_client_open_session(client, server_address_string);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_close_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_close_session(IntPtr client);

		public static void NextClientCloseSession(IntPtr client)
		{
			Next.next_client_close_session(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_is_session_open", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern bool next_client_is_session_open(IntPtr client);

		public static bool NextClientIsSessionOpen(IntPtr client)
		{
			return Next.next_client_is_session_open(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_state", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_client_state(IntPtr client);

		public static int NextClientState(IntPtr client)
		{
			return Next.next_client_state(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_update", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_update(IntPtr client);

		public static void NextClientUpdate(IntPtr client)
		{
			Next.next_client_update(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_send_packet", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_send_packet(IntPtr client, byte[] packet_data, int packet_bytes);

		public static void NextClientSendPacket(IntPtr client, byte[] packet_data, int packet_bytes)
		{
			Next.next_client_send_packet(client, packet_data, packet_bytes);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_send_packet_direct", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_send_packet_direct(IntPtr client, byte[] packet_data, int packet_bytes);

		public static void NextClientSendPacketDirect(IntPtr client, byte[] packet_data, int packet_bytes)
		{
			Next.next_client_send_packet_direct(client, packet_data, packet_bytes);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_report_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_report_session(IntPtr client);

		public static void NextClientReportSession(IntPtr client)
		{
			Next.next_client_report_session(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_session_id", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern ulong next_client_session_id(IntPtr client);

		public static ulong NextClientSessionID(IntPtr client)
		{
			return Next.next_client_session_id(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_stats", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_client_stats(IntPtr client);

		public static ClientStats NextClientStats(IntPtr client)
		{	
			IntPtr clientStatsPtr = Next.next_client_stats(client);
			return Next.GetNextClientStatsFromPointer(clientStatsPtr);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_server_address", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_client_server_address(IntPtr client);

		public static NextAddress NextClientServerAddress(IntPtr client)
		{	
			IntPtr serverAddrPtr = Next.next_client_server_address(client);
			return Next.GetNextAddressFromPointer(serverAddrPtr);
		}

		// ----------------------------------------------------------

		// Network Next server stats
		[StructLayout (LayoutKind.Sequential)]
		public struct ServerStats
		{
			public NextAddress Address;
			public ulong SessionID;
			public ulong UserHash;
			public int PlatformID;
			public int ConnectionType;
			[MarshalAs(UnmanagedType.U1)]
			public bool Next;
			[MarshalAs(UnmanagedType.U1)]
			public bool Committed;
			[MarshalAs(UnmanagedType.U1)]
			public bool Multipath;
			[MarshalAs(UnmanagedType.U1)]
			public bool Reported;
			[MarshalAs(UnmanagedType.U1)]
			public bool FallbackToDirect;
			public float DirectRTT;
			public float DirectJitter;
			public float DirectPacketLoss;
			public float NextRTT;
			public float NextJitter;
			public float NextPacketLoss;
			public float NextKbpsUp;
			public float NextKbpsDown;
			public ulong PacketsSentClientToServer;
			public ulong PacketsSentServerToClient;
			public ulong PacketsLostClientToServer;
			public ulong PacketsLostServerToClient;
			public ulong PacketsOutOfOrderClientToServer;
			public ulong PacketsOutOfOrderServerToClient;
			public float JitterClientToServer;
			public float JitterServerToClient;
			public int NumTags;
			[MarshalAs(UnmanagedType.ByValArray, SizeConst = NEXT_MAX_TAGS)]
			public ulong[] Tags; 
		}

		// Network Next server functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_create", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_server_create(IntPtr ctx, [MarshalAs(UnmanagedType.LPStr)] string server_address, [MarshalAs(UnmanagedType.LPStr)] string bind_address, [MarshalAs(UnmanagedType.LPStr)] string datacenter, NextServerPacketReceivedCallback packet_received_callback, NextWakeupCallback wakeup_callback);

		public static IntPtr NextServerCreate(IntPtr ctx, string server_address, string bind_address, string datacenter, NextServerPacketReceivedCallback packet_received_callback, NextWakeupCallback wakeup_callback)
		{
			return Next.next_server_create(ctx, server_address, bind_address, datacenter, packet_received_callback, wakeup_callback);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_destroy", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_destroy(IntPtr server);

		public static void NextServerDestroy(IntPtr server)
		{
			Next.next_server_destroy(server);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_port", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern ushort next_server_port(IntPtr server);

		public static ushort NextServerPort(IntPtr server)
		{
			return Next.next_server_port(server);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_state", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_server_state(IntPtr server);

		public static int NextServerState(IntPtr server)
		{
			return Next.next_server_state(server);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_update", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_update(IntPtr server);

		public static void NextServerUpdate(IntPtr server)
		{
			Next.next_server_update(server);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_upgrade_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern ulong next_server_upgrade_session(IntPtr server, IntPtr address, [MarshalAs(UnmanagedType.LPStr)] string user_id);

		public static ulong NextServerUpgradeSession(IntPtr server, IntPtr address, string user_id)
		{
			return Next.next_server_upgrade_session(server, address, user_id);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_tag_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_tag_session(IntPtr server, IntPtr address, [MarshalAs(UnmanagedType.LPStr)] string tag);

		public static void NextServerTagSession(IntPtr server, IntPtr address, string tag)
		{
			Next.next_server_tag_session(server, address, tag);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_tag_session_multiple", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_tag_session_multiple(IntPtr server, IntPtr address, string[] tags, int num_tags);

		public static void NextServerTagSessionMultiple(IntPtr server, IntPtr address, string[] tags, int num_tags)
		{
			if (tags.Length > NEXT_MAX_TAGS) {
				string[] maxTags = new string[NEXT_MAX_TAGS];
				for (int i = 0; i < NEXT_MAX_TAGS; i++) {
					maxTags[i] = tags[i];
				}
				Next.next_server_tag_session_multiple(server, address, maxTags, NEXT_MAX_TAGS);
				return;
			}

			Next.next_server_tag_session_multiple(server, address, tags, num_tags);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_session_upgraded", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern bool next_server_session_upgraded(IntPtr server, IntPtr address);

		public static bool NextServerSessionUpgraded(IntPtr server, IntPtr address)
		{
			return Next.next_server_session_upgraded(server, address);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_send_packet", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_send_packet(IntPtr server, IntPtr to_address, byte[] packet_data, int packet_bytes);

		public static void NextServerSendPacket(IntPtr server, IntPtr to_address, byte[] packet_data, int packet_bytes)
		{
			Next.next_server_send_packet(server, to_address, packet_data, packet_bytes);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_send_packet_direct", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_send_packet_direct(IntPtr server, IntPtr to_address, byte[] packet_data, int packet_bytes);

		public static void NextServerSendPacketDirect(IntPtr server, IntPtr to_address, byte[] packet_data, int packet_bytes)
		{
			Next.next_server_send_packet_direct(server, to_address, packet_data, packet_bytes);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_stats", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_stats(IntPtr server, IntPtr address, out ServerStats stats);

		public static void NextServerStats(IntPtr server, IntPtr address, out ServerStats stats)
		{
			Next.next_server_stats(server, address, out stats);
		}

		// ----------------------------------------------------------

		// Network Next mutex
		[StructLayout (LayoutKind.Sequential)]
		public struct NextMutex
		{
			[MarshalAs(UnmanagedType.ByValArray, SizeConst = NEXT_MUTEX_BYTES)]
			public byte[] Dummy;
		}

		// Network Next mutex functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_mutex_create", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_mutex_create(out NextMutex mutex);

		public static int NextMutexCreate(out NextMutex mutex)
		{
			return Next.next_mutex_create(out mutex);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_mutex_destroy", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_mutex_destroy(ref NextMutex mutex);

		public static void NextMutexDestroy(ref NextMutex mutex)
		{
			Next.next_mutex_destroy(ref mutex);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_mutex_acquire", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_mutex_acquire(ref NextMutex mutex);

		public static void NextMutexAcquire(ref NextMutex mutex)
		{
			Next.next_mutex_acquire(ref mutex);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_mutex_release", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_mutex_release(ref NextMutex mutex);

		public static void NextMutexRelease(ref NextMutex mutex)
		{
			Next.next_mutex_release(ref mutex);
		}

		// ----------------------------------------------------------

		// Network Next experimental functions (untested)

		#if (NEXT_EXPERIMENTAL)

			public const double NEXT_PING_DURATION = 10.0;
			public const int NEXT_MAX_PING_TOKENS = 256;
			public const int NEXT_MAX_PING_TOKEN_BYTES = 256;

			[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_customer_id", CharSet = CharSet.Ansi, ExactSpelling = true)]
			private static extern ulong next_customer_id();

			public static ulong NextCustomerID()
			{
				return Next.next_customer_id();
			}

			[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_customer_private_key", CharSet = CharSet.Ansi, ExactSpelling = true)]
			private static extern byte[] next_customer_private_key();

			public static byte[] NextCustomerPrivateKey()
			{
				return Next.next_customer_private_key();
			}

			[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_customer_public_key", CharSet = CharSet.Ansi, ExactSpelling = true)]
			private static extern byte[] next_customer_public_key();

			public static byte[] NextCustomerPublicKey()
			{
				return Next.next_customer_public_key();
			}

			[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_generate_ping_token", CharSet = CharSet.Ansi, ExactSpelling = true)]
			private static extern void next_generate_ping_token(ulong customer_id, byte[] customer_private_key, NextAddress client_address, string datacenter_name, string user_id, out byte[] out_ping_token_data, out int out_ping_token_bytes);

			public static void NextGeneratePingToken(ulong customer_id, byte[] customer_private_key, NextAddress client_address, string datacenter_name, string user_id, out byte[] out_ping_token_data, out int out_ping_token_bytes)
			{
				Next.next_generate_ping_token(customer_id, customer_private_key, client_address, datacenter_name, user_id, out out_ping_token_data, out out_ping_token_bytes);
			}

			[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_validate_ping_token", CharSet = CharSet.Ansi, ExactSpelling = true)]
			private static extern bool next_validate_ping_token(ulong customer_id, byte[] customer_private_key, NextAddress client_address, byte[] ping_token_data, int ping_token_bytes);

			public static bool NextValidatePingToken(ulong customer_id, byte[] customer_private_key, NextAddress client_address, byte[] ping_token_data, int ping_token_bytes)
			{
				return Next.next_validate_ping_token(customer_id, customer_private_key, client_address, ping_token_data, ping_token_bytes);
			}

			public const int NEXT_PING_STATE_RESOLVING_HOSTNAME = 0;
			public const int NEXT_PING_STATE_SENDING_PINGS = 1;
			public const int NEXT_PING_STATE_FINISHED = 2;
			public const int NEXT_PING_STATE_ERROR = 3;

			[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_ping_create", CharSet = CharSet.Ansi, ExactSpelling = true)]
			private static extern IntPtr next_ping_create(IntPtr ctx, string bind_address, ref byte[] ping_token_data, ref int ping_token_bytes, int num_ping_tokens);

			public static IntPtr NextPingCreate(IntPtr ctx, string bind_address, ref byte[] ping_token_data, ref int ping_token_bytes, int num_ping_tokens)
			{
				return Next.next_ping_create(ctx, bind_address, ref ping_token_data, ref ping_token_bytes, num_ping_tokens);
			}

			[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_ping_destroy", CharSet = CharSet.Ansi, ExactSpelling = true)]
			private static extern void next_ping_destroy(IntPtr ping);

			public static void NextPingDestroy(IntPtr ping)
			{
				Next.next_ping_destroy(ping);
			}

			[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_ping_update", CharSet = CharSet.Ansi, ExactSpelling = true)]
			private static extern void next_ping_update(IntPtr ping);

			public static void NextPingUpdate(IntPtr ping)
			{
				Next.next_ping_update(ping);
			}

			[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_ping_state", CharSet = CharSet.Ansi, ExactSpelling = true)]
			private static extern int next_ping_state(IntPtr ping);

			public static int NextPingState(IntPtr ping)
			{
				return Next.next_ping_state(ping);
			}

		#endif // #if NEXT_EXPERIMENTAL 

		// ----------------------------------------------------------

		// Network Next unit test functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_test", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_test();

		public static void NextTest()
		{
			Next.next_test();
		}
	}
}