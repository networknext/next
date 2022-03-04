using System;
using System.Runtime.InteropServices;
using System.Runtime.CompilerServices;
using System.Text;
using System.Collections.Generic;
using AOT;

namespace NetworkNext {

	// Network Next delegate functions
	
	#region Delegates definition

	/**
	  * <summary>
	  *		Callback function when a Network Next client receives a packet.
	  * </summary>
	  * <param name="clientPtr">the pointer to the client object.</param>
	  * <param name="ctxPtr">the pointer to the client's context.</param>
	  * <param name="fromPtr">the pointer to the <see cref="NextAddress"/> that sent the packet to the client.</param>
	  * <param name="packetDataPtr">the pointer to the packet data received.</param>
	  * <param name="packetBytes">the number of bytes received in the packet data.</param>
	*/
	public delegate void NextClientPacketReceivedCallback(IntPtr clientPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes);

	/**
	  * <summary>
	  * 	Callback function when a packet is ready to be received by the Network Next client.
	  *     <remarks>
	  *			<para> 
	  * 			Intended to let you set an event of your own creation when a packet is ready to receive,
	  * 			making it possible to use Network Next with applications built around traditional select or wait
	  * 			for multiple event style blocking socket loops. Call NextClientUpdate to pump received packets 
	  * 			to the NextClientPacketReceivedCallback when you wake up on your main thread from your event.
	  *         </para>
	  *     </remarks>
	  * </summary>
	  * <param name="ctxPtr">the pointer to the client's context.</param>
	*/
	public delegate void NextWakeupCallback(IntPtr ctxPtr);

	/**
	  * <summary>
	  * 	Callback function when a Network Next server receives a packet.
	  * </summary>
	  * <param name="serverPtr">the pointer to the server object.</param>
	  * <param name="ctxPtr">the pointer to the server's context.</param>
	  * <param name="fromPtr">the pointer to the <see cref="NextAddress"/> that sent the packet to the server.</param>
	  * <param name="packetDataPtr">the pointer to the packet data received.</param>
	  * <param name="packetBytes">the number of bytes received in the packet data.</param>
	*/
	public delegate void NextServerPacketReceivedCallback(IntPtr serverPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes);

	/**
	* <summary>
	*	Callback function for defining custom logging for the Network Next SDK. Only logs <= the current log level are received.
	* 	<remarks>
	*   	<para>
	*			C# does not support C-style formatting, making the <c>format</c> parameter useless.
	* 			Instead, the <c>args</c> parameter will have the message to log.
	*		</para>
	*	</remarks>
	* </summary>
	* <param name="level">the log level of the message</param>
	* <list type="bullet">
	* 	<item>
	*		<description>
	*			<para
	*				><em>NEXT_LOG_LEVEL_NONE (0)
	*				</em>
	*			</para>
	*		</description>
	*	</item>
	* 	<item>
	*		<description>
	*			<para>
	*				<em>NEXT_LOG_LEVEL_ERROR (1)</em>
	*			</para>
	*		</description>
	*	</item>
	* 	<item>
	*		<description>
	*			<para>
	*				<em>NEXT_LOG_LEVEL_INFO (2)</em>
	*			</para>
	*		</description>
	*	</item>
	* 	<item>
	*		<description>
	*			<para>
	*				<em>NEXT_LOG_LEVEL_WARN (3)</em>
	*			</para>
	*		</description>
	*	</item>
	* 	<item>
	*		<description>
	*			<para>
	*				<em>NEXT_LOG_LEVEL_DEBUG (4)</em>
	*			</para>
	*		</description>
	*	</item>
	* </list>
	* <param name="format">the pointer to the ANSI string containing the C-style formatting.</param>
	* <param name="args">the pointer to a ANSI string containing the log message.</param>
	* <example>
	* This shows how to unmarshal the <c>args</c> pointer into a string.
	* <code>
	*		string argsStr = Marshal.PtrToStringAnsi(args);
	* </code>
	* </example>
	*/
	public delegate void NextLogFunction(int level, IntPtr format, IntPtr args);

	/**
	* <summary>
	* 	Callback function for custom assert handler.
	* </summary>
	* <param name="condition">the assert condition that failed.</param>
	* <param name="function">the function name where the assertion failed.</param>
	* <param name="file">the file name where the assertion failed.</param>
	* <param name="line">the line number where the assertion failed.<param>
	*/
	public delegate void NextAssertFunction(bool condition, [MarshalAs(UnmanagedType.LPStr)] string function, [MarshalAs(UnmanagedType.LPStr)] string file, int line);

	/**
	* <summary>
	*	Callback function for custom memory allocation.
	* </summary>
	* <param name="ctxPtr">the pointer to the provided context</param>
	* <param name="bytes">the number of bytes to allocate</param>
	* <returns> an <see cref="IntPtr"/> to the allocated object</returns>
	*/
	public delegate IntPtr NextMallocFunction(IntPtr ctxPtr, ulong bytes);

	/**
	* <summary>
	*	Callback function for custom memory free function.
	* </summary>
	* <param name="ctxPtr">the pointer to the provided context</param>
	* <param name="p">the pointer to the object to free</param>
	*/
	public delegate void NextFreeFunction(IntPtr ctxPtr, IntPtr p);

	#endregion // #region Delegates definition

	// ----------------------------------------------------------

	// Network Next class

	#region Next defintion

	/**
	* <summary>
	*	C# Wrapper class for Network Next SDK.
	* 	Utilizes P/Invoke to call out to unmanaged C++ code.
	*	<remarks>
	* 		Every function in next.h is accessible, and utility functions
	*		are provided to easily convert <see cref="IntPtr"/>s and
	*		struct fields.
	*	</remarks> 
	* </summary>
	*/
	public static class Next
	{
		#if UNITY_IOS
			const string dll = "__Internal";
		#else 
			const string dll = "next-unity";
		#endif // #if UNITY_IOS

		// ----------------------------------------------------------

		// Network Next constants

		#region Constants defintion

		public enum NEXT_BOOL : int
		{
			NEXT_FALSE = 0,
			NEXT_TRUE = 1,
		}

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

		public const int NEXT_MAX_MATCH_VALUES = 64;

		public const int NEXT_DEFAULT_SOCKET_SEND_BUFFER_SIZE = 1000000;
		public const int NEXT_DEFAULT_SOCKET_RECEIVE_BUFFER_SIZE = 1000000;

		public const int NEXT_CLIENT_STATE_CLOSED = 0;
		public const int NEXT_CLIENT_STATE_OPEN = 1;                               
		public const int NEXT_CLIENT_STATE_ERROR = 2;

		public const int NEXT_SERVER_STATE_DIRECT_ONLY = 0;
		public const int NEXT_SERVER_STATE_INITIALIZING = 1;
		public const int NEXT_SERVER_STATE_INITIALIZED = 2;

		public const int NEXT_MUTEX_BYTES = 256;

		#endregion // #region Constants definition

		// ----------------------------------------------------------
		
		// Network Next Wrapper Utility functions

		#region Utility functions

		/**
		* <summary>
		*	Gets the IPv4 address from <see cref="NextAddress"/>.
		* </summary>
		* <param name="address">the <see cref="NextAddress"/> to obtain the IPv4 from</param>
		* <returns>a <see cref="byte"/> array containing each component of the IPv4 address</returns>
		*/
		public static byte[] GetNextAddressIPV4(NextAddress address)
		{
			byte[] addr = new byte[]{address.IPV4_0, address.IPV4_1, address.IPV4_2, address.IPV4_3};
			return addr;
		}

		/**
		* <summary>
		*	Gets the IPv6 address from <see cref="NextAddress"/>.
		* </summary>
		* <param name="address">the <see cref="NextAddress"/> to obtain the IPv6 from</param>
		* <returns>a <see cref="ushort"/> array containing each component of the IPv6 address</returns>
		*/
		public static ushort[] GetNextAddressIPV6(NextAddress address)
		{
			ushort[] addr = new ushort[]{address.IPV6_0, address.IPV6_1, address.IPV6_2, address.IPV6_3, address.IPV6_4, address.IPV6_5, address.IPV6_6, address.IPV6_7};
			return addr;
		}

		/**
		* <summary>
		*	Gets a <see cref="NextAddress"/> struct from an <see cref="IntPtr"/>.
		* </summary>
		* <param name="nextAddressPointer">the pointer to the <see cref="NextAddress"/> struct</param>
		* <returns>a <see cref="NextAddress"/> struct from the pointer</returns>
		*/ 
		public static NextAddress GetNextAddressFromPointer(IntPtr nextAddressPointer)
		{
			return (NextAddress)Marshal.PtrToStructure(nextAddressPointer, typeof(NextAddress));
		}

		/**
		* <summary>
		*	Gets a <see cref="ClientStats"/> struct from an <see cref="IntPtr"/>.
		* </summary>
		* <param name="clientStatsPointer">the pointer to the <see cref="ClientStats"/> struct</param>
		* <returns>a <see cref="ClientStats"/> struct from the pointer</returns>
		*/ 
		public static ClientStats GetNextClientStatsFromPointer(IntPtr clientStatsPointer)
		{
			ClientStatsInternal internalStats = (ClientStatsInternal)Marshal.PtrToStructure(clientStatsPointer, typeof(ClientStatsInternal));
			ClientStats stats = new ClientStats();
			stats.PlatformID = internalStats.PlatformID;
			stats.ConnectionType = internalStats.ConnectionType;
			stats.Next = internalStats.Next == NEXT_BOOL.NEXT_TRUE;
			stats.Upgraded = internalStats.Upgraded == NEXT_BOOL.NEXT_TRUE;
			stats.Committed = internalStats.Committed == NEXT_BOOL.NEXT_TRUE;
			stats.Multipath = internalStats.Multipath == NEXT_BOOL.NEXT_TRUE;
			stats.Reported = internalStats.Reported == NEXT_BOOL.NEXT_TRUE;
			stats.FallbackToDirect = internalStats.FallbackToDirect == NEXT_BOOL.NEXT_TRUE;
			stats.HighFrequencyPings = internalStats.HighFrequencyPings == NEXT_BOOL.NEXT_TRUE;
			stats.DirectMinRTT = internalStats.DirectMinRTT;
			stats.DirectMinRTT = internalStats.DirectMaxRTT;
			stats.DirectPrimeRTT = internalStats.DirectPrimeRTT;
			stats.DirectJitter = internalStats.DirectJitter;
			stats.DirectPacketLoss = internalStats.DirectPacketLoss;
			stats.NextRTT = internalStats.NextRTT;
			stats.NextJitter = internalStats.NextJitter;
			stats.NextPacketLoss = internalStats.NextPacketLoss;
			stats.NextKbpsUp = internalStats.NextKbpsUp;
			stats.NextKbpsDown = internalStats.NextKbpsDown;
			stats.PacketsSentClientToServer = internalStats.PacketsSentClientToServer;
			stats.PacketsSentServerToClient = internalStats.PacketsSentServerToClient;
			stats.PacketsLostClientToServer = internalStats.PacketsLostClientToServer;
			stats.PacketsLostServerToClient = internalStats.PacketsLostServerToClient;
			stats.PacketsOutOfOrderClientToServer = internalStats.PacketsOutOfOrderClientToServer;
			stats.PacketsOutOfOrderServerToClient = internalStats.PacketsOutOfOrderServerToClient;
			stats.JitterClientToServer = internalStats.JitterClientToServer;
			stats.JitterServerToClient = internalStats.JitterServerToClient;

			return stats;
		}

		#endregion // #region Utility functions

		// ----------------------------------------------------------

		// Network Next Config
		
		#region NextConfig definition

		/**
		* <summary>
		*	Configuration struct for the Network Next SDK.
		* </summary>
		* <value <see cref="ServerBackendHostname"/> - The hostname for the backend the Network Next SDK is talking to. Set to "prod.spacecats.net" by default.</value>
		* <value <see cref="CustomerPublicKey"/>- The customer public key as a base64 encoded string.</value>
		* <value <see cref="CustomerPrivateKey"/>- The customer private key as a base64 encoded string.</value>
		* <value <see cref="SocketSendBufferSize"/> - The size of the socket send buffer in bytes.</value>
		* <value <see cref="SocketReceiveBufferSize"/> - The size of the socket receive buffer in bytes.</value>
		* <value <see cref="DisableNetworkNext"/> - Set this to true to disable Network Next entirely and always send packets across the public internet.</value>
		*/
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

		/**
		* <summary>
		*	Internal version of <see cref="NextConfig"/> to handle <see cref="NEXT_BOOL"/>.
		* </summary>
		*/
		[StructLayout (LayoutKind.Sequential, CharSet=CharSet.Ansi)]
		private struct NextConfigInternal
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
			public NEXT_BOOL DisableNetworkNext;
		}

		#endregion // #region NextConfig definition

		// Network Next global functions

		#region Global functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_default_config")]
		private static extern void next_default_config(out NextConfigInternal config);

		/**
		* <summary>
		*	Sets the given config to a default configuration.
		*	<remarks>
		*		Use this to set default values for config variables, then make only the changes you want on top.
		*	</remarks>
		* </summary>
		* <list type="bullet">
		* 	<item>
		*		<description>
		*			<para>
						<em>ServerBackendHostname - "prod.spacecats.net"</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>CustomerPublicKey - ""</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>CustomerPrivateKey - ""</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>SocketSendBufferSize - 1000000</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>SocketReceiveBufferSize - 1000000</em>
		*			</para>
		*		</description>
		*	</item>
		*	<item>
		*		<description>
		*			<para>
		*				<em>DisableNetworkNext - false</em>
		*			</para>
		*		</description>
		*	</item>
		* </list>
		* <example>
		*	This is how to get the default config.
		* <code>
		*	NextConfig config;
		*	Next.NextDefaultConfig(out config);
		*	config.CustomerPublicKey = "my_public_key"; // replace this with your public key
		*	Debug.Log(String.Format("default hostname is {0}", config.ServerBackendHostname));
		* </code>
		* </example>
		*/
		public static void NextDefaultConfig(out NextConfig config)
		{
			NextConfigInternal internalConfig;
			next_default_config(out internalConfig);

			config = new NextConfig();
			config.ServerBackendHostname = internalConfig.ServerBackendHostname;
			config.CustomerPublicKey = internalConfig.CustomerPublicKey;
			config.CustomerPrivateKey = internalConfig.CustomerPrivateKey;
			config.SocketSendBufferSize = internalConfig.SocketSendBufferSize;
			config.SocketReceiveBufferSize = internalConfig.SocketReceiveBufferSize;
			config.DisableNetworkNext = internalConfig.DisableNetworkNext == NEXT_BOOL.NEXT_TRUE;
		}
		
		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_init", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_init(IntPtr ctxPtr, ref NextConfigInternal config);

		/**
		* <summary>
		*	Initializes the Network Next SDK.
		*	<remarks>
		*		Call this before creating a client or server.
		*		Note that context is an IntPtr because it is user defined, and thus needs to be marshaled and unmarshaled by the user.
		*		Reference: https://docs.microsoft.com/en-us/dotnet/api/system.runtime.interopservices.marshal.ptrtostructure?view=net-5.0#System_Runtime_InteropServices_Marshal_PtrToStructure_System_IntPtr_System_Type_
		*	</remarks>
		* </summary>
		* <param name="ctxPtr">an optional pointer to context passed to overriden malloc and free functions for global allocations. Use <see cref="IntPtr.Zero"/> to pass in empty context</param>
		* <param name="config">a <see cref="NextConfig"/> object. Pass in an empty <see cref="NextConfig"/> to use the default configuration</param>
		* <returns>NEXT_OK if the Network Next SDK initialized successfully, NEXT_ERROR otherwise.</returns>
		* <example>
		* 	This is how you can pass your custom context struct and <see cref="NextConfig"/> to <see cref="NextInit"/>.
		* <code>
		*	// Create a context
		* 	Context ctx = new Context();
		*
		* 	// Marshal the context into a pointer
		*	ctxPtr = Marshal.AllocHGlobal(Marshal.SizeOf(ctx));
		*	Marshal.StructureToPtr(ctx, ctxPtr, false);
		*
		*	// Get the default configuration
		*	Next.NextConfig config;
		*	Next.NextDefaultConfig(out config);
		*
		*	// Assign our public key to the configuration
		*	config.CustomerPublicKey = customerPublicKey;
		*
		*	// Initialize Network Next
		*	if (Next.NextInit(ctxPtr, ref config) != Next.NEXT_OK)
		*	{
		*		Debug.LogError("error: could not initialize network next");
		*	}
		*
		*	// Free the context once we are no longer using it
		*	Marshal.FreeHGlobal(ctxPtr);  
		* </code>
		* </example>
		* <example>
		*	This is how you pass in an empty context and <see cref="NextConfig"/>.
		* <code>
		*	// Get the default configuration
		*	Next.NextConfig config;
		*	Next.NextDefaultConfig(out config);
		*
		*	// Assign our public key to the configuration
		*	config.CustomerPublicKey = customerPublicKey;
		*
		*	if (Next.NextInit(IntPtr.Zero, ref config) != Next.NEXT_OK)
		*	{
		*		Debug.LogError("error: could not initialize network next");
		*	}	
		* </code>
		* </example>
		*/
		public static int NextInit(IntPtr ctxPtr, ref NextConfig config)
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

			NextConfigInternal internalConfig = new NextConfigInternal();
			internalConfig.ServerBackendHostname = config.ServerBackendHostname;
			internalConfig.CustomerPublicKey = config.CustomerPublicKey;
			internalConfig.CustomerPrivateKey = config.CustomerPrivateKey;
			internalConfig.SocketSendBufferSize = config.SocketSendBufferSize;
			internalConfig.SocketReceiveBufferSize = config.SocketReceiveBufferSize;
			if (config.DisableNetworkNext)
			{
				internalConfig.DisableNetworkNext = NEXT_BOOL.NEXT_TRUE;
			}

			return next_init(ctxPtr, ref internalConfig);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_term")]
		private static extern void next_term();

		/**
		* <summary>
		*	Shuts down the Network Next SDK.
		*	<remarks>
		*		Call this before you shut down your application.
		*	</remarks>
		* </summary>
		* <example>
		* <code>
		*	Next.NextTerm();
		* </code>
		* </example>
		*/
		public static void NextTerm()
		{
			next_term();
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_time", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern double next_time();

		/**
		* <summary>
		*	Gets the current time in seconds.
		*	<remarks>
		*		IMPORTANT: Only defined when called after <see cref="NextInit"/>.
		*	</remarks>
		* </summary>
		* <example>
		* <code>
		*	Next.NextConfig config = new Next.NextConfig();
		*	Next.NextInit(IntPtr.Zero, ref config);
		*	// ... do stuff ...
		*	Debug.Log(String.Format("{0} seconds since NextInit", Next.NextTime().ToString("F2")));
		* </code>
		* </example>
		*/
		public static double NextTime() {
			return next_time();
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_sleep", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_sleep(double timeSeconds);

		/**
		* <summary>
		*	Sleep for some amount of time.
		* </summary>
		* <param name="timeSeconds">the length of time to sleep in seconds</param>
		* <example>
		* <code>
		*	Next.NextConfig config = new Next.NextConfig();
		*	Next.NextInit(IntPtr.Zero, ref config);
		*	double startTime = Next.NextTime();
		*	Next.NextSleep(10.0);
		*	double finishTime = Next.NextTime();
		*	double delta = finishTime - startTime;
		*	Debug.Log(String.Format("slept for {0} seconds", delta.ToString("F2")));
		* </code>
		* </example>
		*/
		public static void NextSleep(double timeSeconds) {
			next_sleep(timeSeconds);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_printf", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_printf(int level, [MarshalAs(UnmanagedType.LPStr)] string format);

		/**
		* <summary>
		*	Log level aware printf.
		*	<remarks>
		*		To output logs to the Unity console, define your custom <see cref="NextLogFunction"/> callback and assign it using <see cref="NextLogFunction"/>.
		*	</remarks>
		* </summary>
		* <param name="level">the log level. Only logs <= the current log level are printed</param>
		* <list type="bullet">
		* 	<item>
		*		<description>
		*			<para
		*				><em>NEXT_LOG_LEVEL_NONE (0)
		*				</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_LOG_LEVEL_ERROR (1)</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_LOG_LEVEL_INFO (2)</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_LOG_LEVEL_WARN (3)</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_LOG_LEVEL_DEBUG (4)</em>
		*			</para>
		*		</description>
		*	</item>
		* </list>
		* <param name="format">the formatted string to print</param>
		* <param name="args">additional variadic args that will be appended to the <paramref name="format"></param>
		*/
		public static void NextPrintf(int level, string format, params Object[] args) {
			// Create the string to pass
			StringBuilder sb = new StringBuilder(format);

			foreach (Object o in args)
			{
				sb.AppendFormat("{0} ", o);
			}

			next_printf(level, sb.ToString());
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "default_assert_function", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void default_assert_function([MarshalAs(UnmanagedType.LPStr)] string condition, [MarshalAs(UnmanagedType.LPStr)] string function, [MarshalAs(UnmanagedType.LPStr)] string file, int line);

		/**
		* <summary>
		*	Assert a condition.
		*	<remarks>
		*		This is implemented in C# because we cannot export macro functions from C++.
		*		Asserts are only ran if the UNITY_ASSERTIONS macro is true.
		*	</remarks>
		* </summary>
		* <param name="condiiton">the condition to assert</param>
		* <param name="assertFunction">optional custom assert function to use instead of default. Useful to gracefully prevent editor crashes</param>
		* <param name="function">the function name where the assertion failed. Provided by System.Runtime.CompilerServices</param>
		* <param name="file">the file name where the assertion failed. Provided by System.Runtime.CompilerServices</param>
		* <param name="line">the line number where the assertion failed. Provided by System.Runtime.CompilerServices<param>
		* <example>
		* <code>
		*	Next.NextAssert(true != false);
		* </code>
		* </example>
		*/
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
					default_assert_function(condition.ToString(), function, file, line);
				}
			#endif // #if UNITY_ASSERTIONS
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_quiet", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_quiet(NEXT_BOOL flag);

		/**
		* <summary>
		*	Enable/disable network next logs entirely.
		* </summary>
		* <example>
		* <code>
		*	// Shut up network next!
		*	Next.NextQuiet(true);
		* </code>
		* </example>
		*/
		public static void NextQuiet(bool flag) {
			if (flag)
			{
				next_quiet(NEXT_BOOL.NEXT_TRUE);
			}
			else
			{
				next_quiet(NEXT_BOOL.NEXT_FALSE);
			}
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_log_level", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_log_level(int level);

		/**
		* <summary>
		*	Sets the Network Next log level.
		*	<remarks>
		*		The default log level is info. This includes both info messages and errors, which are both infrequent.
		*	</remarks>
		* </summary>
		* <param name="level">the log level. Only logs <= the current log level are printed</param>
		* <list type="bullet">
		* 	<item>
		*		<description>
		*			<para
		*				><em>NEXT_LOG_LEVEL_NONE (0)
		*				</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_LOG_LEVEL_ERROR (1)</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_LOG_LEVEL_INFO (2)</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_LOG_LEVEL_WARN (3)</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_LOG_LEVEL_DEBUG (4)</em>
		*			</para>
		*		</description>
		*	</item>
		* </list>
		* <example>
		* <code>
		*	// Unleash the kraken!
		*	Next.NextLogLevel(Next.NEXT_LOG_LEVEL_DEBUG);
		* </code>
		* </example>
		*/
		public static void NextLogLevel(int level) {
			next_log_level(level);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_log_function", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_log_function(NextLogFunction function);

		/**
		* <summary>
		* 	Sets a custom log function.
		* </summary>
		* <param name="function">the custom log function</param>
		* <example>
		* <code>
		*	// Determines the log type from the level 
		*	static string LogLevelString(int level)
		*	{
		*		if (level == Next.NEXT_LOG_LEVEL_ERROR) {
		*		    return "error";
		*		} else if (level == Next.NEXT_LOG_LEVEL_INFO) {
		*		    return "info";
		*		} else if (level == Next.NEXT_LOG_LEVEL_WARN) {
		*		    return "warn";
		*		} else if (level == Next.NEXT_LOG_LEVEL_DEBUG) {
		*		    return "debug";
		*		} else {
		*		    return "???";
		*		}
		*	}
		*	
		*	enum Color { red, green, blue, black, white, yellow, orange };
		*	
		*	// Define custom logging function to output to Unity console
		*	[MonoPInvokeCallback(typeof(NextLogFunction))]
		*	static void UnityLogger(int level, IntPtr formatPtr, IntPtr argsPtr)
		*	{
		*		// Unmarshal the log message into a string
		*	    string argsStr = Marshal.PtrToStringAnsi(argsPtr);
		*	
		*	    // Choose a colour for the log depending on the log level
		*	    Color c;
		*	    if (level == Next.NEXT_LOG_LEVEL_ERROR) {
		*	        c = Color.red;
		*	    } else if (level == Next.NEXT_LOG_LEVEL_INFO) {
		*	        c = Color.green;
		*	    } else if (level == Next.NEXT_LOG_LEVEL_WARN) {
		*	        c = Color.yellow;
		*	    } else if (level == Next.NEXT_LOG_LEVEL_DEBUG) {
		*	        c = Color.orange;
		*	    } else {
		*	        c = Color.white;
		*	    }
		*	
		*	    if (level != Next.NEXT_LOG_LEVEL_NONE)
		*	    {
		*	    	// Log to Unity console
		*	    	Debug.Log(String.Format("<color={0}>{1}: {2}: {3}</color>", 
		*	    		c.ToString(), Next.NextTime().ToString("F2"), LogLevelString(level), argsStr)
		*	    	);
		*	    }
		*	}
		*	
		*	void Start()
		*	{
		*		Next.NextLogFunction(UnityLogger);
		*	
		*		Next.NextConfig config = new Next.NextConfig();
		*		Next.NextInit(IntPtr.Zero, ref config);
		*	
		*		Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, "Hi, Mum!");
		*	
		*		Next.NextTerm();
		*	}
		* </code>
		* </example>
		*/
		public static void NextLogFunction(NextLogFunction function) {
			next_log_function(function);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_assert_function", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_assert_function(NextAssertFunction function);

		/**
		* <summary>
		*	Set a custom assert handler.
		* </summary>
		* <param name="function">the custom callback for the assert handler</param>
		* <example>
		* <code>
		*	// Define custom assert function
		*	[MonoPInvokeCallback(typeof(NextAssertFunction))]
		*	static void AssertFunction(bool condition, string function, string file, int line)
		*	{
		*		#if UNITY_EDITOR
		*			// Stops the editor cleanly
		*			Debug.LogError(String.Format("assert failed: ({0}), function {1}, file {2}, line {3}", condition, function, file, line));
		*			Assert.IsFalse(condition, String.Format("assert failed: ({0}), function {1}, file {2}, line {3}", condition, function, file, line));
		*	
		*			UnityEditor.EditorApplication.isPlaying = false;
		*		#else
		*			Application.Quit();
		*		#endif // #if UNITY_EDITOR 
		*	}
		*	
		*	void Start()
		*	{
		*		// Assign our custom assert function
	    *		Next.NextAssertFunction(AssertFunction);
		*	}
		* </code>
		* </example>
		*/
		public static void NextAssertFunction(NextAssertFunction function) {
			next_assert_function(function);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_allocator", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_allocator(NextMallocFunction mallocFunction, NextFreeFunction freeFunction);

		/**
		* <summary>
		*	Set a custom allocator.
		*	<remarks>
		*		Only set a custom allocator if context is not empty (<see cref="IntPtr.Zero"/>).
		*	</remarks>
		* </summary>
		* <example>
		* <code>
		*	// Define custom malloc function
		*	[MonoPInvokeCallback(typeof(NextMallocFunction))]
		*	static IntPtr MallocFunction(IntPtr ctxPtr, ulong bytes)
		*	{    
		*		// Obtain the malloc function from the custom allocator class stored in context
		*		Context ctx = (Context)Marshal.PtrToStructure(ctxPtr, typeof(Context));
		*	
		*		Next.NextAssert(!ctx.Equals(default(Context)));
		*	
		*		GCHandle allocatorGCH = GCHandle.FromIntPtr(ctx.AllocatorGCH);
		*		Allocator allocator = (Allocator)allocatorGCH.Target; 
		*	
		*		Next.NextAssert(allocator != null);
		*	
		*		return allocator.Alloc((int)bytes);
		*	}
		*	
		*	// Define custom free function
		*	[MonoPInvokeCallback(typeof(NextFreeFunction))]
		*	static void FreeFunction(IntPtr ctxPtr, IntPtr p)
		*	{
		*		// Obtain the free function from the custom allocator class stored in context
		*		Context ctx = (Context)Marshal.PtrToStructure(ctxPtr, typeof(Context));
		*	
		*		Next.NextAssert(!ctx.Equals(default(Context)));
		*	
		*		GCHandle allocatorGCH = GCHandle.FromIntPtr(ctx.AllocatorGCH);
		*		Allocator allocator = (Allocator)allocatorGCH.Target; 
		*	
		*		Next.NextAssert(allocator != null);
		*	
		*		allocator.Free(p);	
		*	}
		*	
		*	void Start() {
		*		Next.NextAllocator(MallocFunction, FreeFunction);
		*	}
		* </code>
		* </example>
		*/
		public static void NextAllocator(NextMallocFunction mallocFunction, NextFreeFunction freeFunction) {
			next_allocator(mallocFunction, freeFunction);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_user_id_string", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_user_id_string(ulong userID, StringBuilder buffer);

		/**
		* <summary>
		*	Converts a legacy <see cref="ulong"/> user ID to a string.
		*	<remarks>
		*		This is the equivalent of <c>String.Format("{0:x}", userID)</c>.
		*		Make sure the capacity is at least 16, otherwise the string will be shortened.
		*	</remarks>
		* </summary>
		* <param name="userID">the user ID to convert to a string</param>
		* <param name="capacity">the capacity of the string. Recommended to at least be 16</param>
		* <example>
		* <code>
		*	Next.NextServerUpgradeSession(serverPtr, fromPtr, Next.NextUserIDString(userID, 16));
		* </code>
		* </example>
		*/
		public static string NextUserIDString(ulong userID, int capacity = 16) {
			StringBuilder buffer = new StringBuilder(capacity);
			next_user_id_string(userID, buffer);
			return buffer.ToString();
		}

		#endregion // #region Global functions
	
		// ----------------------------------------------------------

		// Network Next address

		#region NextAddress definition

		/**
		* <summary>
		*	Next Address struct for the Network Next SDK.
		*	This is a struct that can represent any IPv4 or IPv6 address and port.
		*	<remarks>
		*		It’s used when sending and receiving packets. For example, in the server packet received callback, the address of the client is passed to you via this structure.
		*		The IPv4 and IPv6 addresses share memory to save space.
		*		Use the <see cref="GetNextAddressIPV4"/> and <see cref="GetNextAddressIPV6"/> utility functions to get the IP.
		*	</remarks>
		* </summary>
		* <value <see cref="IPV4_0"/> - The first component of the IPv4 address.</value>
		* <value <see cref="IPV4_1"/> - The second component of the IPv4 address.</value>
		* <value <see cref="IPV4_2"/> - The third component of the IPv4 address.</value>
		* <value <see cref="IPV4_3"/> - The fourth component of the IPv4 address.</value>
		* <value <see cref="IPV6_0"/> - The first component of the IPv6 address.</value>
		* <value <see cref="IPV6_1"/> - The second component of the IPv6 address.</value>
		* <value <see cref="IPV6_2"/> - The third component of the IPv6 address.</value>
		* <value <see cref="IPV6_3"/> - The fourth component of the IPv6 address.</value>
		* <value <see cref="IPV6_4"/> - The fifth component of the IPv6 address.</value>
		* <value <see cref="IPV6_5"/> - The sixth component of the IPv6 address.</value>
		* <value <see cref="IPV6_6"/> - The seventh component of the IPv6 address.</value>
		* <value <see cref="IPV6_7"/> - The eighth component of the IPv6 address.</value>
		* <value <see cref="Port"/> - The port number of the address.</value>
		* <value <see cref="Type"/> - 0 for IPv4, 1 for IPv6.</value>
		*/
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

		#endregion // #region NextAddress definition

		// Network Next address functions

		#region NextAddress functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_address_parse", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_address_parse(out NextAddress address, [MarshalAs(UnmanagedType.LPStr)] string addressString);

		/**
		* <summary>
		*	Parses a <see cref="NextAddress"/> from a string.
		* </summary>
		* <param name="address">the <see cref="NextAddress"/> to be initialized</param>
		* <param name="addressString">an address string in IPv4 or IPv6 format</param>
		* <returns>NEXT_OK if the address was parsed successfully, NEXT_ERROR otherwise.</returns>
		* <example>
		* <code>
		*	Next.NextAddress address;
		*	Next.NextAddressParse(out address, "127.0.0.1");
		*	Next.NextAddressParse(out address, "127.0.0.1:50000");
		*	Next.NextAddressParse(out address, "::1");
		*	Next.NextAddressParse(out address, "[::1]:50000");
		* </code>
		* </example>
		*/
		public static int NextAddressParse(out NextAddress address, string addressString) {
			return next_address_parse(out address, addressString);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_address_to_string", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_address_to_string(ref NextAddress address, StringBuilder buffer);

		/**
		* <summary>
		*	Converts a <see cref="NextAddress"/> to a string.
		* </summary>
		* <param name="address">the address to convert to a string</param>
		* <param name="capacity">the buffer size to store the string representation of the address. Must be at least NEXT_MAX_ADDRESS_STRING_LENGTH</param>
		* <returns>a string of the <see cref="NextAddress"/>. Makes it easy to print.</returns>
		* <example>
		* <code>
		*	Next.NextAddress address;
		*	Next.NextAddressParse(out address, "[::1]:50000");
		*	Debug.Log(String.Format("address string = {0}", Next.NextAddressString(ref address, Next.NEXT_MAX_ADDRESS_STRING_LENGTH)));
		* </code>
		* </example>
		*/
		public static string NextAddressToString(ref NextAddress address, int capacity = NEXT_MAX_ADDRESS_STRING_LENGTH) {
			if (capacity > NEXT_MAX_ADDRESS_STRING_LENGTH)
			{
				capacity = NEXT_MAX_ADDRESS_STRING_LENGTH;
			}
			StringBuilder buffer = new StringBuilder(capacity);
			next_address_to_string(ref address, buffer);
			return buffer.ToString();
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_address_equal", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern NEXT_BOOL next_address_equal(ref NextAddress a, ref NextAddress b);

		/**
		* <summary>
		*	Checks if two <see cref="NextAddress"/> structs are equal.
		* </summary>
		* <param name="a">the first address to compare</param>
		* <param name="b">the second address to compare</param>
		* <returns><see langword="true"/> if the addresses are equal, <see langword="false"/> otherwise.</returns>
		* <example>
		* <code>
		*	Next.NextAddress a;
		* 	Next.NextAddress b;
		* 	Next.NextAddressParse(out a, "127.0.0.1");
		*	Next.NextAddressParse(out b, "127.0.0.1:0");
		*	bool addressesAreEqual = Next.NextAddressEqual(ref a, ref b);
		*	Debug.Log(String.Format("addresses are equal = {0}", addressesAreEqual ? "yes" : "no"));
		* </code>
		* </example>
		*/
		public static bool NextAddressEqual(ref NextAddress a, ref NextAddress b) {
			NEXT_BOOL equal = next_address_equal(ref a, ref b);
			return equal == NEXT_BOOL.NEXT_TRUE;
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_address_anonymize", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_address_anonymize(ref NextAddress address);

		/**
		* <summary>
		*	Anonymizes a <see cref="NextAddress"/> by zeroing the last tuple and port.
		* </summary>
		* <param name="address">the address to anonymize</param>
		* <example>
		* <code>
		*	Next.NextAddress address;
		*	Next.NextAddressParse(out address, "127.0.0.1:50000");
		*	Next.NextAddressAnonymize(ref address);
		*	Debug.Log(String.Format("anonymized address string = {0}", Next.NextAddressString(ref address, Next.NEXT_MAX_ADDRESS_STRING_LENGTH)));
		* </code>
		* </example>
		*/
		public static void NextAddressAnonymize(ref NextAddress address) {
			next_address_anonymize(ref address);
		}

		#endregion // #region NextAddress functions

		// ----------------------------------------------------------

		// Network Next client stats

		#region ClientStats definition

		/**
		* <summary>
		*	Client Stats struct for the Network Next SDK.
		*	This is a struct that holds client statistics.
		*	<remarks>
		*		Due to name conflicts, this name of this struct was changed from NextClientStats to ClientStats.
		*	</remarks>
		* </summary>
		* <value <see cref="PlatformID"/> - The platform this client is using (see constants for a list of values).</value>
		* <value <see cref="ConnectionType"/> - The connection type of the client (see constants for a list of values).</value>
		* <value <see cref="Next"/> - If the client is accelerated by Network Next.</value>
		* <value <see cref="Upgraded"/> - If the client is upgraded by the server for monitoring and acceleration on Network Next.</value>
		* <value <see cref="Committed"/> - If the client is ready to take a Network Next route.</value>
		* <value <see cref="Multipath"/> - If the client is enabled to send packets across the public internet (direct) and Network Next.</value>
		* <value <see cref="Reported"/> - If the client has been reported the session as problematic.</value>
		* <value <see cref="FallbackToDirect"/> - If the client has defaulted to using the public internet (direct) in the event of an error on Network Next.</value>
		* <value <see cref="HighFrequencyPings"/> - If the client has high frequency pings enabled.</value>
		* <value <see cref="DirectMinRTT"/> - The smallest RTT between the client and server on the public interent (direct) route in the last 10 seconds.</value>
		* <value <see cref="DirectMaxRTT"/> - The largest RTT between the client and server on the public interent (direct) route in the last 10 seconds.</value>
		* <value <see cref="DirectPrimeRTT"/> - The second largest RTT between the client and server on the public interent (direct) route in the last 10 seconds. Used for approximating P99 etc.</value>
		* <value <see cref="DirectJitter"/> - The jitter between the client and server on the public interent (direct) route.</value>
		* <value <see cref="DirectPacketLoss"/> - The packet loss between the client and server on the public interent (direct) route.</value>
		* <value <see cref="NextRTT"/> - The RTT between the client and server on the Network Next route.</value>
		* <value <see cref="NextJitter"/> - The jitter between the client and server on the Network Next route.</value>
		* <value <see cref="NextPacketLoss"/> - The packet loss between the client and server on the Network Next route.</value>
		* <value <see cref="NextKbpsUp"/> - The bandwidth up in kbps on Network Next.</value>
		* <value <see cref="NextKbpsDown"/> - The bandwidth down in kbps on Network Next.</value>
		* <value <see cref="PacketsSentClientToServer"/> - The number of packets the client sent to the server.</value>
		* <value <see cref="PacketsSentServerToClient"/> - The number of packets the server sent to the client.</value>
		* <value <see cref="PacketsLostClientToServer"/> - The number of packets the client lost sending to the server.</value>
		* <value <see cref="PacketsLostServerToClient"/> - The number of packets the server lost sending to the client.</value>
		* <value <see cref="PacketsOutOfOrderClientToServer"/> - The number of packets the client received out of from the server.</value>
		* <value <see cref="PacketsOutOfOrderServerToClient"/> - The number of packets the server received out of from the client.</value>
		* <value <see cref="JitterClientToServer"/> - The jitter between the client and server.</value>
		* <value <see cref="JitterServerToClient"/> - The jitter between the server and client.</value>
		*/
		[StructLayout (LayoutKind.Sequential)]
		public struct ClientStats
		{
			public int PlatformID;
			public int ConnectionType;
			public bool Next;
			public bool Upgraded;
			public bool Committed;
			public bool Multipath;
			public bool Reported;
			public bool FallbackToDirect;
			public bool HighFrequencyPings;
			public float DirectMinRTT;
			public float DirectMaxRTT;
			public float DirectPrimeRTT;
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

		/**
		* <summary>
		*	Internal version of <see cref="ClientStats"/> to handle <see cref="NEXT_BOOL"/>.
		* </summary>
		*/
		[StructLayout (LayoutKind.Sequential)]
		private struct ClientStatsInternal
		{
			public int PlatformID;
			public int ConnectionType;
			public NEXT_BOOL Next;
			public NEXT_BOOL Upgraded;
			public NEXT_BOOL Committed;
			public NEXT_BOOL Multipath;
			public NEXT_BOOL Reported;
			public NEXT_BOOL FallbackToDirect;
			public NEXT_BOOL HighFrequencyPings;
			public float DirectMinRTT;
			public float DirectMaxRTT;
			public float DirectPrimeRTT;
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

		#endregion // #region ClientStats definition

		// Network Next client functions

		#region NextClient functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_create", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_client_create(IntPtr ctxPtr, [MarshalAs(UnmanagedType.LPStr)] string bindAddress, NextClientPacketReceivedCallback packetReceivedCallback, NextWakeupCallback wakeupCallback);

		/**
		* <summary>
		*	Creates an instance of a client, binding a socket to the specified address and port.
		*	<remarks>
		*		The client is kept as a pointer because it is a complex struct to use with P/Invoke.
		*		Functions exist to get and set any client-related properties.
		*	</remarks>
		* </summary>
		* <param name="ctxPtr">an optional pointer to context passed to any callbacks made from the client. Pass in <see cref="IntPtr.Zero"/> if not used</param>
		* <param name="bindAddress>an address string describing the bind address and port to bind to. Typically "0.0.0.0:0" is passed in, which binds to any IPv4 interface and lets the system pick a port. Alternatively, you can bind to a specific port, for example: "0.0.0.0:50000"</param>
		* <param name="packetReceivedCallback">called from the same thread that calls <see cref="NextClientUpdate"/>, whenever a packet is received from the server. Required</param>
		* <param name="wakeupCallback">optional callback. Pass <see langword="null"/> if not used. Sets a callback function to be called from an internal network next thread when a packet is ready to be received for this client. Intended to let you set an event of your own creation when a packet is ready to receive, making it possible to use Network Next with applications built around traditional select or wait for multiple event style blocking socket loops. Call <see cref="NextClientUpdate"/> to pump received packets to the <paramref name="packetReceivedCallback"> when you wake up on your main thread from your event</param>
		* <returns>a pointer to the client, or <see cref="IntPtr.Zero"/> if the client could not be created. Typically, <see cref="IntPtr.Zero"/> is only returned when another socket is already bound on the same port, or if an invalid bind address is passed in.</returns>
		* <example>
		* <code>
		*	// First define a callback for received packets
		*	[MonoPInvokeCallback(typeof(NextClientPacketReceivedCallback))]
		*	static void ClientPacketReceived(IntPtr clientPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
		*	{
		*		Debug.Log(String.Format("client received packet from server ({0} bytes)", packetBytes));
		*	}
		*	
		*	void Start()
		*	{
		*		Next.NextConfig config = new Next.NextConfig();
		*		Next.NextInit(IntPtr.Zero, config);
		*	
		*		// Create a client
		*		IntPtr client = Next.NextClientCreate(IntPtr.Zero, "0.0.0.0:0", ClientPacketReceived, null);
		*		if (client.Equals(IntPtr.Zero))
		*		{
		*			Debug.LogError("error: failed to create client");
		*			return;
		*		}
		*	}
		* </code>
		* </example>
		*/
		public static IntPtr NextClientCreate(IntPtr ctxPtr, string bindAddress, NextClientPacketReceivedCallback packetReceivedCallback, NextWakeupCallback wakeupCallback)
		{
			return next_client_create(ctxPtr, bindAddress, packetReceivedCallback, wakeupCallback);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_destroy", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_destroy(IntPtr client);

		/**
		* <summary>
		*	Destorys a client instance, and the socket it manages internally.
		* </summary>
		* <param name="client">the client to destroy. Must be a valid client pointer created by <see cref="NextClientCreate"/>. Do NOT pass in <see cref="IntPtr.Zero"/> or <see langword="null"/>.</param>
		* <example>
		* <code>
		*	Next.NextClientDestroy(client);
		* </code>
		* </example>
		*/
		public static void NextClientDestroy(IntPtr client)
		{
			next_client_destroy(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_port", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern ushort next_client_port(IntPtr client);

		/**
		* <summary>
		*	Gets the port the client socket is bound to.
		*	<remarks>
		*		This makes it possible to look up what specific port the client is bound to when you bind to port zero and the system chooses a port.
		*	</remarks>
		* </summary>
		* <param name="client">the client to get the port of</param>
		* <returns>the port number the client socket is bound to</param>
		* <example>
		* <code>
		*	// Create a client
		*	IntPtr client = Next.NextClientCreate(IntPtr.Zero, "0.0.0.0:0", ClientPacketReceived, null);
		*	if (client.Equals(IntPtr.Zero))
		*	{
		*		Debug.LogError("error: failed to create client");
		*		return;
		*	}
		*	
		*	ushort clientPort = Next.NextClientPort(client);
		*	Debug.Log(String.Format("the client was bound to port {0}", clientPort));
		* </code>
		* </example>
		*/
		public static ushort NextClientPort(IntPtr client)
		{
			return next_client_port(client);
		}
		
		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_open_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_open_session(IntPtr client, [MarshalAs(UnmanagedType.LPStr)] string serverAddress);

		/**
		* <summary>
		*	Opens a session between the client and server.
		* </summary>
		* <param name="client">the client pointer</param>
		* <param name="serverAddress">the address of the server that the client wants to connect to</param>
		* <example>
		* <code>
		*	Next.NextClientOpenSession(client, "127.0.0.1:50000");
		* </code>
		* </example>
		*/
		public static void NextClientOpenSession(IntPtr client, string serverAddress)
		{
			next_client_open_session(client, serverAddress);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_close_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_close_session(IntPtr client);

		/**
		* <summary>
		*	Closes the session between the client and server.
		* </summary>
		* <param name="client">the client pointer</param>
		* <example>
		* <code>
		*	Next.NextClientCloseSession(client);
		* </code>
		* </example>
		*/
		public static void NextClientCloseSession(IntPtr client)
		{
			next_client_close_session(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_is_session_open", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern NEXT_BOOL next_client_is_session_open(IntPtr client);

		/**
		* <summary>
		* 	Check if the client has a session open.
		* </summary>
		* <param name="client">the client pointer</param>
		* <returns><see langword="true"/> if the client has an open session with a server, <see langword="false"/> otherwise.</returns>
		* <example>
		* <code>
		*	bool sessionOpen = Next.NextClientSessionOpen(client);
		*	Debug.Log(String.Format("session open = {0}", sessionOpen ? "yes" : "no"));
		* </code>
		* </example>
		*/
		public static bool NextClientIsSessionOpen(IntPtr client)
		{
			NEXT_BOOL open = next_client_is_session_open(client);
			return open == NEXT_BOOL.NEXT_TRUE;
		}
		
		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_state", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_client_state(IntPtr client);

		/**
		* <summary>
		*	Gets the state the client is in.
		* 	<remarks>
		*		The client is initially in closed state. 
		*		After <see cref="NextClientOpenSession"/> the client is immediately in open state on success, or error state if something went wrong while opening the session,
		*		for example, an invalid server address was passed in.
		*	</remarks>
		* </summary>
		* <param name="client">the client pointer</param>
		* <returns>the client state is either:
		* <list type="bullet">
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_CLIENT_STATE_CLOSED</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_CLIENT_STATE_OPEN</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_CLIENT_STATE_ERROR</em>
		*			</para>
		*		</description>
		*	</item>
		* </list>
		* </returns>
		* <example>
		* <code>
		*	string stateStr = "???";
		*	int state = Next.NextClientState(client);
		*	switch (state)
		*	{
		*		case Next.NEXT_CLIENT_STATE_CLOSED:
		*		    stateStr = "closed";
		*		    break;
		*	
		*		case Next.NEXT_CLIENT_STATE_OPEN:
		*		    stateStr = "open";
		*		    break;
		*	
		*		case Next.NEXT_CLIENT_STATE_ERROR:
		*		    stateStr = "error";
		*		    break;
		*	
		*		default:
		*		    break;
		*	}
		*	
		*	Debug.Log(String.Format("client state = {0} ({1})", stateStr, state));
		* </code>
		* </example>
		*/
		public static int NextClientState(IntPtr client)
		{
			return next_client_state(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_update", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_update(IntPtr client);

		/**
		* <summary>
		*	Updates the client.
		*	<remarks>
		*		Please call this every frame as it drives the packet received callback.
		*	</remarks>
		* </summary>
		* <param name="client">the client pointer</param>
		* <example>
		* <code>
		*	void Update()
		*	{
		*		Next.NextClientUpdate(client);
		*		// ... do stuff ...
		*	}
		* </code>
		* </example>
		*/
		public static void NextClientUpdate(IntPtr client)
		{
			next_client_update(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_send_packet", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_send_packet(IntPtr client, byte[] packetData, int packetBytes);

		/**
		* <summary>
		*	Sends a packet to the server.
		*	<remarks>
		*		Depending on whether this player is accelerated or not, this packet will be sent direct across the public internet, or through Network Next’s network of private networks.
		*	</remarks>
		* </summary>
		* <param name="client">the client pointer</param>
		* <param name="packetData">the packet data to send to the server</param>
		* <param name="packetBytes"> the size of the packet in bytes. Must be in range 1 to NEXT_MTU (1300)</param>
		* <example>
		* <code>
		*	void Update()
		*	{
		*		Next.NextClientUpdate(client);
		*		
		*		// Create the packet
		*		int packetBytes;
		*		byte[] packetData = generatePacket(out packetBytes);
		*		
		*		Next.NextClientSendPacket(client, packetData, packetBytes); 
		*	}
		* </code>
		* </example>
		*/
		public static void NextClientSendPacket(IntPtr client, byte[] packetData, int packetBytes)
		{
			next_client_send_packet(client, packetData, packetBytes);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_send_packet_direct", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_send_packet_direct(IntPtr client, byte[] packetData, int packetBytes);

		/**
		* <summary>
		*	Sends a packet to the server, forcing the packet to be sent across the public internet.
		*	<remarks>
		*		The packet will be sent unaccelerated across the public internet and will not count towards your Network Next bandwidth envelope.
		*		This can be very useful when you need to send a burst of non-latency sensitive packets, for example, in a load screen.
		*	</remarks>
		* </summary>
		* <param name="client">the client pointer</param>
		* <param name="packetData">the packet data to send to the server</param>
		* <param name="packetBytes"> the size of the packet in bytes. Must be in range 1 to NEXT_MTU (1300)</param>
		* <example>
		* <code>
		*	void Update()
		*	{
		*		Next.NextClientUpdate(client);
		*		
		*		// Create the packet
		*		int packetBytes;
		*		byte[] packetData = generatePacket(out packetBytes);
		*		
		*		Next.NextClientSendPacketDirect(client, packetData, packetBytes); 
		*	}
		* </code>
		* </example>
		*/
		public static void NextClientSendPacketDirect(IntPtr client, byte[] packetData, int packetBytes)
		{
			next_client_send_packet_direct(client, packetData, packetBytes);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_report_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_client_report_session(IntPtr client);

		/**
		* <summary>
		*	Report the session as problematic.
		*	<remarks>
		*		This feature was added to support our customers who let players flag bad play sessions in their game UI.
		*		Call this function when your players complain, and it’s sent to our backend so we can help you track down why!
		*	</remarks>
		* </summary>
		* <param name="client">the client pointer</param>
		* <example>
		* <code>
		*	Next.NextClientReportSession(client);
		* </code>
		* </example>
		*/
		public static void NextClientReportSession(IntPtr client)
		{
			next_client_report_session(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_session_id", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern ulong next_client_session_id(IntPtr client);

		/**
		* <summary>
		*	Gets the client session ID.
		*	<remarks>
		*		A session ID uniquely identifies each session on Network Next.
		*		Session IDs are distinct from user IDs. User IDs are unique on a per-user basis, while session IDs are unique for each call to <see cref="NextClientOpenSession"/>.
		*		A session ID is assigned when the server upgrades the session via <see cref="NextServerUpgradeSession"/>. Until that point the session ID is 0.
		*	</remarks>
		* </summary>
		* <param name="client">the client pointer</param>
		* <returns>the session id, if the client has been upgraded, otherwise 0.</returns>
		* <example>
		* <code>
		*	ulong sessionID = Next.NextClientSessionID(client);
		*	Debug.Log(String.Format("session id = {0}", sessionID));
		* </code>
		* </example>
		*/
		public static ulong NextClientSessionID(IntPtr client)
		{
			return next_client_session_id(client);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_stats", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_client_stats(IntPtr client);

		/**
		* <summary>
		*	Gets client statistics.
		* </summary>
		* <param name="client">the client pointer</param>
		* <returns>a <see cref="ClientStats"/> struct.</returns>
		* <example>
		* 	Here is how to query it, and print out various interesting values.
		* <code>
		*	StringBuilder sb = new StringBuilder("================================================================\n");
		*	
		*	Next.ClientStats stats = Next.NextClientStats(client);
		*	string platform = "unknown";
		*	
		*	switch (stats.PlatformID)
		*	{
		*		case Next.NEXT_PLATFORM_WINDOWS:
		*	        platform = "windows";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_MAC:
		*	        platform = "mac";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_LINUX:
		*	        platform = "linux";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_SWITCH:
		*	        platform = "nintendo switch";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_PS4:
		*	        platform = "ps4";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_PS5:
		*	        platform = "ps5";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_IOS:
		*	        platform = "ios";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_XBOX_ONE:
		*	        platform = "xbox one";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_XBOX_SERIES_X:
		*	        platform = "xbox series x";
		*	        break;
		*	
		*	    default:
		*	        break;
		*	}
		*	
		*	string stateStr = "???";
		*	int state = Next.NextClientState(client);
		*	switch (state)
		*	{
		*		case Next.NEXT_CLIENT_STATE_CLOSED:
		*		    stateStr = "closed";
		*		    break;
		*	
		*		case Next.NEXT_CLIENT_STATE_OPEN:
		*		    stateStr = "open";
		*		    break;
		*	
		*		case Next.NEXT_CLIENT_STATE_ERROR:
		*		    stateStr = "error";
		*		    break;
		*	
		*		default:
		*		    break;
		*	}
		*	
		*	sb.AppendFormat("state = {0} ({1})\n", stateStr, state);
		*	sb.AppendFormat("session id = {0}\n", Next.NextClientSessionID(client));
		*	sb.AppendFormat("platform id = {0} ({1})\n", platform, (int)stats.PlatformID);
		*	
		*	string connection = "unknown";
		*	
		*	switch (stats.ConnectionType)
		*	{
		*		case Next.NEXT_CONNECTION_TYPE_WIRED:
		*		    connection = "wired";
		*		    break;
		*	
		*		case Next.NEXT_CONNECTION_TYPE_WIFI:
		*		    connection = "wifi";
		*		    break;
		*	
		*		case Next.NEXT_CONNECTION_TYPE_CELLULAR:
		*		    connection = "cellular";
		*		    break;
		*	
		*		default:
		*		    break;
		*	}
		*	
		*	sb.AppendFormat("connection type = {0} ({1})\n", connection, stats.ConnectionType);
		*	
		*	if (!stats.FallbackToDirect)
		*	{
		*		sb.AppendFormat("upgraded = {0}\n", stats.Upgraded.ToString());
		*		sb.AppendFormat("committed = {0}\n", stats.Committed.ToString());
		*		sb.AppendFormat("multipath = {0}\n", stats.Multipath.ToString());
		*		sb.AppendFormat("reported = {0}\n", stats.Reported.ToString());
		*	}
		*	
		*	sb.AppendFormat("fallback to direct = {0}\n", stats.FallbackToDirect.ToString());
		*	
		*	sb.AppendFormat("high frequency pings = {0}\n", stats.HighFrequencyPings.ToString());
		*	
		*	sb.AppendFormat("direct min rtt = {0}ms\n", stats.DirectMinRTT.ToString("F"));
		*	sb.AppendFormat("direct max rtt = {0}ms\n", stats.DirectMaxRTT.ToString("F"));
		*	sb.AppendFormat("direct prime rtt = {0}ms\n", stats.DirectPrimeRTT.ToString("F"));
		*	sb.AppendFormat("direct jitter = {0}ms\n", stats.DirectJitter.ToString("F"));
		*	sb.AppendFormat("direct packet loss = {0}%\n", stats.DirectPacketLoss.ToString("F1"));
		*	
		*	if (stats.Next)
		*	{
		*		sb.AppendFormat("next rtt = {0}ms\n", stats.NextRTT.ToString("F"));
		*		sb.AppendFormat("next jitter = {0}ms\n", stats.NextJitter.ToString("F"));
		*		sb.AppendFormat("next packet loss = {0}%\n", stats.NextPacketLoss.ToString("F1"));
		*		sb.AppendFormat("next bandwidth up = {0}kbps\n", stats.NextKbpsUp.ToString("F1"));
		*		sb.AppendFormat("next bandwidth down = {0}kbps\n", stats.NextKbpsDown.ToString("F1"));
		*	}
		*	
		*	if (stats.Upgraded && !stats.FallbackToDirect)
		*	{
		*		sb.AppendFormat("packets sent client to server = {0}\n", stats.PacketsSentClientToServer);
		*		sb.AppendFormat("packets sent server to client = {0}\n", stats.PacketsSentServerToClient);
		*		sb.AppendFormat("packets lost client to server = {0}\n", stats.PacketsLostClientToServer);
		*		sb.AppendFormat("packets lost server to client = {0}\n", stats.PacketsLostServerToClient);
		*		sb.AppendFormat("packets out of order client to server = {0}\n", stats.PacketsOutOfOrderClientToServer);
		*		sb.AppendFormat("packets out of order server to client = {0}\n", stats.PacketsOutOfOrderServerToClient);
		*		sb.AppendFormat("jitter client to server = {0}\n", stats.JitterClientToServer.ToString("F"));
		*		sb.AppendFormat("jitter server to client = {0}\n", stats.JitterServerToClient.ToString("F"));
		*	}
		*	
		*	sb.AppendFormat("================================================================\n");
		*	
		*	Debug.Log(sb.ToString());
		* </code>
		* </example>
		*/
		public static ClientStats NextClientStats(IntPtr client)
		{	
			IntPtr clientStatsPtr = next_client_stats(client);
			return GetNextClientStatsFromPointer(clientStatsPtr);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_client_server_address", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_client_server_address(IntPtr client);

		/**
		* <summary>
		* 	Gets the server address that a client is connected to.
		* </summary>
		* <name param="client">the client pointer</param>
		* <returns>a <see cref="NextAddress"/> struct with the server's address.</returns>
		*/
		public static NextAddress NextClientServerAddress(IntPtr client)
		{	
			IntPtr serverAddrPtr = next_client_server_address(client);
			return GetNextAddressFromPointer(serverAddrPtr);
		}

		#endregion #region NextClient functions

		// ----------------------------------------------------------
		
		// Network Next server stats

		#region ServerStats definition

		/**
		* <summary>
		*	Server Stats struct for the Network Next SDK.
		*	This is a struct that holds client statistics with a server.
		*	<remarks>
		*		Due to name conflicts, this name of this struct was changed from NextServerStats to ServerStats.
		*	</remarks>
		* </summary>
		* <value <see cref="Address"/> - The <see cref="NextAddress"/> of the client.</value>
		* <value <see cref="SessionID"/> - The session ID of the client.</value>
		* <value <see cref="UserHash"/> - The hash of the client's user ID.</value>
		* <value <see cref="PlatformID"/> - The platform this client is using (see constants for a list of values).</value>
		* <value <see cref="ConnectionType"/> - The connection type of the client (see constants for a list of values).</value>
		* <value <see cref="Next"/> - If the client is accelerated by Network Next.</value>
		* <value <see cref="Upgraded"/> - If the client is upgraded by the server for monitoring and acceleration on Network Next.</value>
		* <value <see cref="Committed"/> - If the client is ready to take a Network Next route.</value>
		* <value <see cref="Multipath"/> - If the client is enabled to send packets across the public internet (direct) and Network Next.</value>
		* <value <see cref="Reported"/> - If the client has been reported the session as problematic.</value>
		* <value <see cref="FallbackToDirect"/> - If the client has defaulted to using the public internet (direct) in the event of an error on Network Next.</value>
		* <value <see cref="HighFrequencyPings"/> - If the client has high frequency pings enabled.</value>
		* <value <see cref="HighFrequencyPings"/> - If the client has high frequency pings enabled.</value>
		* <value <see cref="DirectMinRTT"/> - The smallest RTT between the client and server on the public interent (direct) route in the last 10 seconds.</value>
		* <value <see cref="DirectMaxRTT"/> - The largest RTT between the client and server on the public interent (direct) route in the last 10 seconds.</value>
		* <value <see cref="DirectPrimeRTT"/> - The second largest RTT between the client and server on the public interent (direct) route in the last 10 seconds. Used for approximating P99 etc.</value>
		* <value <see cref="DirectJitter"/> - The jitter between the client and server on the public interent (direct) route.</value>
		* <value <see cref="DirectPacketLoss"/> - The packet loss between the client and server on the public interent (direct) route.</value>
		* <value <see cref="NextRTT"/> - The RTT between the client and server on the Network Next route.</value>
		* <value <see cref="NextJitter"/> - The jitter between the client and server on the Network Next route.</value>
		* <value <see cref="NextPacketLoss"/> - The packet loss between the client and server on the Network Next route.</value>
		* <value <see cref="NextKbpsUp"/> - The bandwidth up in kbps on Network Next.</value>
		* <value <see cref="NextKbpsDown"/> - The bandwidth down in kbps on Network Next.</value>
		* <value <see cref="PacketsSentClientToServer"/> - The number of packets the client sent to the server.</value>
		* <value <see cref="PacketsSentServerToClient"/> - The number of packets the server sent to the client.</value>
		* <value <see cref="PacketsLostClientToServer"/> - The number of packets the client lost sending to the server.</value>
		* <value <see cref="PacketsLostServerToClient"/> - The number of packets the server lost sending to the client.</value>
		* <value <see cref="PacketsOutOfOrderClientToServer"/> - The number of packets the client received out of from the server.</value>
		* <value <see cref="PacketsOutOfOrderServerToClient"/> - The number of packets the server received out of from the client.</value>
		* <value <see cref="JitterClientToServer"/> - The jitter between the client and server.</value>
		* <value <see cref="JitterServerToClient"/> - The jitter between the server and client.</value>
		* <value <see cref="NumTags"/> - The number of tags assigned to the client (max of NEXT_MAX_TAGS).</value>
		* <value <see cref="Tags"/> - The tags assigned to the client.</value>
		*/
		[StructLayout (LayoutKind.Sequential)]
		public struct ServerStats
		{
			public NextAddress Address;
			public ulong SessionID;
			public ulong UserHash;
			public int PlatformID;
			public int ConnectionType;
			public bool Next;
			public bool Committed;
			public bool Multipath;
			public bool Reported;
			public bool FallbackToDirect;
			public float DirectMinRTT;
			public float DirectMaxRTT;
			public float DirectPrimeRTT;
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

		/**
		* <summary>
		*	Internal version of <see cref="ServerStats"/> to handle <see cref="NEXT_BOOL"/>.
		* </summary>
		*/
		[StructLayout (LayoutKind.Sequential)]
		private struct ServerStatsInternal
		{
			public NextAddress Address;
			public ulong SessionID;
			public ulong UserHash;
			public int PlatformID;
			public int ConnectionType;
			public NEXT_BOOL Next;
			public NEXT_BOOL Committed;
			public NEXT_BOOL Multipath;
			public NEXT_BOOL Reported;
			public NEXT_BOOL FallbackToDirect;
			public float DirectMinRTT;
			public float DirectMaxRTT;
			public float DirectPrimeRTT;
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

		#endregion // #region ServerStats definition

		// Network Next server functions

		#region NextServer functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_create", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_server_create(IntPtr ctxPtr, [MarshalAs(UnmanagedType.LPStr)] string serverAddress, [MarshalAs(UnmanagedType.LPStr)] string bindAddress, [MarshalAs(UnmanagedType.LPStr)] string datacenter, NextServerPacketReceivedCallback packetReceivedCallback, NextWakeupCallback wakeupCallback);

		/**
		* <summary>
		*	Creates an instance of a server, binding a socket to the specified address and port.
		*	<remarks>
		*		The server is kept as a pointer because it is a complex struct to use with P/Invoke.
		*		Functions exist to get and set any server-related properties.
		*	</remarks>
		* </summary>
		* <param name="ctxPtr">an optional pointer to context passed to any callbacks made from the server. Pass in <see cref="IntPtr.Zero"/> if not used</param>
		* <param name="serverAddress>the public IP address and port that clients will connect to</param>
		* <param name="bindAddress>an address string describing the bind address and port to bind to. Typically "0.0.0.0:[portnum]" is passed in, binding the server socket to any IPv4 interface on a specific port, for example:  "0.0.0.0:50000"</param>
		* <param name="datacenter">the name of the datacenter that the game server is running in. Please pass in “local” until we work with you to determine the set of datacenters you host servers in</param>
		* <param name="packetReceivedCallback">called from the same thread that calls <see cref="NextServerUpdate"/>, whenever a packet is received from the client. Required</param>
		* <param name="wakeupCallback">optional callback. Pass <see langword="null"/> if not used. Sets a callback function to be called from an internal network next thread when a packet is ready to be received for this server. Intended to let you set an event of your own creation when a packet is ready to receive, making it possible to use Network Next with applications built around traditional select or wait for multiple event style blocking socket loops. Call <see cref="NextServerUpdate"/> to pump received packets to the <paramref name="packetReceivedCallback"> when you wake up on your main thread from your event</param>
		* <returns>a pointer to the server, or <see cref="IntPtr.Zero"/> if the server could not be created. Typically, <see cref="IntPtr.Zero"/> is only returned when another socket is already bound on the same port, or if an invalid server or bind address is passed in.</returns>
		* <example>
		* <code>
		*	// First define a callback for received packets
		*	[MonoPInvokeCallback(typeof(NextServerPacketReceivedCallback))]
		*	static void ServerPacketReceived(IntPtr serverPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
		*	{
		*		Next.NextAddress clientAddr = Next.GetNextAddressFromPointer(fromPtr);
		*		Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server received packet from client {0} ({0} bytes)", Next.NextAddressToString(ref clientAddr, Next.NEXT_MAX_ADDRESS_STRING_LENGTH), packetBytes));		
		*	}
		*	
		*	void Start()
		*	{
		*		Next.NextConfig config = new Next.NextConfig();
		*		Next.NextInit(IntPtr.Zero, config);
		*	
		*		// Create a server
		*		IntPtr server = Next.NextServerCreate(IntPtr.Zero, "127.0.0.1", "0.0.0.0:50000", "local", ServerPacketReceived, null);
		*		if (server.Equals(IntPtr.Zero))
		*		{
		*			Debug.LogError("error: failed to create server");
		*			return;
		*		}
		*	}
		* </code>
		* </example>
		*/
		public static IntPtr NextServerCreate(IntPtr ctxPtr, string serverAddress, string bindAddress, string datacenter, NextServerPacketReceivedCallback packetReceivedCallback, NextWakeupCallback wakeupCallback)
		{
			return next_server_create(ctxPtr, serverAddress, bindAddress, datacenter, packetReceivedCallback, wakeupCallback);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_destroy", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_destroy(IntPtr server);

		/**
		* <summary>
		*	Destroys a server instance, and the socket it manages internally.
		* </summary>
		* <param name="server">the server to destroy. Must be a valid server pointer created by <see cref="NextServerCreate"/>. Do NOT pass in <see cref="IntPtr.Zero"/> or <see langword="null"/>.</param>
		* <example>
		* <code>
		*	Next.NextServerDestroy(server);
		* </code>
		* </example>
		*/
		public static void NextServerDestroy(IntPtr server)
		{
			next_server_destroy(server);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_port", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern ushort next_server_port(IntPtr server);

		/**
		* <summary>
		*	Gets the port the server socket is bound to.
		* </summary>
		* <param name="server">the server to get the port of</param>
		* <returns>the port number the server socket is bound to</param>
		* <example>
		* <code>
		*	// Create a server
		*	IntPtr server = Next.NextServerCreate(IntPtr.Zero, "127.0.0.1", "0.0.0.0:50000", "local" serverPacketReceived, null);
		*	if (server.Equals(IntPtr.Zero))
		*	{
		*		Debug.LogError("error: failed to create server");
		*		return;
		*	}
		*	
		*	ushort serverPort = Next.NextServerPort(server);
		*	Debug.Log(String.Format("the server was bound to port {0}", serverPort));
		* </code>
		* </example>
		*/
		public static ushort NextServerPort(IntPtr server)
		{
			return next_server_port(server);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_address", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern NextAddress next_server_address(IntPtr server);

		/**
		* <summary>
		*	Gets the address of the server.
		* </summary>
		* <param name="server"> the server to get the address of</param>
		* <returns>the address of the server</param>
		* <example>
		* <code>
		*	// Create a server
		*	IntPtr server = Next.NextServerCreate(IntPtr.Zero, "127.0.0.1", "0.0.0.0:0", "local" serverPacketReceived, null);
		*	if (server.Equals(IntPtr.Zero))
		*	{
		*		Debug.LogError("error: failed to create server");
		*		return;
		*	}
		*	
		*	Next.NextAddress serverAddr = Next.NextServerAddress(server);
		*	Debug.Log(String.Format("the server address is {0}", Next.NextAddressToString(ref serverAddr, Next.NEXT_MAX_ADDRESS_STRING_LENGTH)));
		* </code>
		* </example>
		*/
		public static NextAddress NextServerAddress(IntPtr server)
		{
			return next_server_address(server);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_state", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_server_state(IntPtr server);

		/**
		* <summary>
		*	Gets the state the server is in.
		* 	<remarks>
		*		The server is initially in the direct only state.
		*		If a valid customer private key is setup, the server will first try to resolve the backend hostname, which is "prod.spacecats.net" by default.
		*		Once the backend hostname is resolved, the server initializes with the backend. When everything works, the server lands in the initialized state and is ready to accelerate players.
		*		If anything fails, the server falls back to the direct only state, and only serves up direct routes over the public internet.
		*	</remarks>
		* </summary>
		* <param name="server">the server pointer</param>
		* <returns>the server state, which is one of the following:
		* <list type="bullet">
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_SERVER_STATE_DIRECT_ONLY</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_SERVER_STATE_RESOLVING_HOSTNAME</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_SERVER_STATE_INITIALIZING</em>
		*			</para>
		*		</description>
		*	</item>
		* 	<item>
		*		<description>
		*			<para>
		*				<em>NEXT_SERVER_STATE_INITIALIZED</em>
		*			</para>
		*		</description>
		*	</item>
		* </list>
		* </returns>
		* <example>
		* <code>
		*	string stateStr = "???";
		*	int state = Next.NextServerState(server);
		*	switch (state)
		*	{
		*		case Next.NEXT_SERVER_STATE_DIRECT_ONLY:
		*		    stateStr = "direct only";
		*		    break;
		*	
		*		case Next.NEXT_SERVER_STATE_RESOLVING_HOSTNAME:
		*		    stateStr = "resolving hostname";
		*		    break;
		*	
		*		case Next.NEXT_SERVER_STATE_INITIALIZING:
		*		    stateStr = "initializing";
		*		    break;
		*		
		*		case Next.NEXT_SERVER_STATE_INITIALIZED:
		*		    stateStr = "intialized";
		*		    break;
		*
		*		default:
		*		    break;
		*	}
		*	
		*	Debug.Log(String.Format("server state = {0} ({1})", stateStr, state));
		* </code>
		* </example>
		*/
		public static int NextServerState(IntPtr server)
		{
			return next_server_state(server);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_update", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_update(IntPtr server);

		/**
		* <summary>
		*	Updates the server.
		*	<remarks>
		*		Please call this every frame as it drives the packet received callback.
		*	</remarks>
		* </summary>
		* <param name="server">the server pointer</param>
		* <example>
		* <code>
		*	void Update()
		*	{
		*		Next.NextServerUpdate(server);
		*		// ... do stuff ...
		*	}
		* </code>
		* </example>
		*/
		public static void NextServerUpdate(IntPtr server)
		{
			next_server_update(server);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_upgrade_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern ulong next_server_upgrade_session(IntPtr server, IntPtr addressPtr, [MarshalAs(UnmanagedType.LPStr)] string userID);

		/**
		* <summary>
		*	Upgrades a session for monitoring and potential acceleration by Network Next.
		* 	<remarks>
		*		IMPORTANT: Make sure you only call this function when you are 100% sure this is a real player in your game.
		*	</remarks>
		* </summary>
		* <param name="server">the server pointer</param>
		* <param name="addressPtr">the pointer to the <see cref="NextAddress"/> of the client to be upgraded</param>
		* <param name="userID">the userID for the session. Pass in any unique per-user identifier you have</param>
		* <returns>the session ID assigned to the session that was upgraded</param>
		* <example>
		*	This example shows how to upgrade a session in the packet received callback.
		* <code>
		*	[MonoPInvokeCallback(typeof(NextServerPacketReceivedCallback))]
		*	static void ServerPacketReceived(IntPtr serverPtr, IntPtr ctxPtr, IntPtr NextAddress fromPtr, IntPtr packetDataPtr, int packetBytes)
		*	{
		*		Next.NextAddress clientAddr = Next.GetNextAddressFromPointer(fromPtr);
		*		Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server received packet from client {0} ({0} bytes)", Next.NextAddressToString(ref clientAddr, Next.NEXT_MAX_ADDRESS_STRING_LENGTH), packetBytes));
		*		// ... verify the player is real ...
		*		
		*		// Upgrade the session
		*			ulong sessionID = Next.NextServerUpgradeSession(serverPtr, fromPtr, userID);
		*			Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, Strint.Format("server upgraded session {0}", sessionID));
		*	}
		* </code>
		* </example>
		*/
		public static ulong NextServerUpgradeSession(IntPtr server, IntPtr addressPtr, string userID)
		{
			return next_server_upgrade_session(server, addressPtr, userID);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_tag_session", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_tag_session(IntPtr server, IntPtr addressPtr, [MarshalAs(UnmanagedType.LPStr)] string tag);

		/**
		* <summary>
		*	Tags a session for potentially different network optimization parameters.
		* </summary>
		* <param name="server">the server pointer</param>
		* <param name="addressPtr">the pointer to the <see cref="NextAddress"/> of the client to be tagged</param>
		* <param name="tag">the tag to be applied to the client. Some ideas: "pro", "streamer" or "dev"</param>
		* <example>
		* <code>
		*	Next.NextServerTagSession(server, clientAddrPtr, "pro");
		* </code>
		* </example>
		*/
		public static void NextServerTagSession(IntPtr server, IntPtr addressPtr, string tag)
		{
			next_server_tag_session(server, addressPtr, tag);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_tag_session_multiple", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_tag_session_multiple(IntPtr server, IntPtr addressPtr, string[] tags, int numTags);

		/**
		* <summary>
		*	Tags a session with multiple tags for potentially different network optimization parameters.
		* </summary>
		* <param name="server">the server pointer</param>
		* <param name="addressPtr">the pointer to the <see cref="NextAddress"/> of the client to be tagged</param>
		* <param name="tags">the tags to be applied to the client. Some ideas: "pro", "streamer" or "dev"</param>
		* <param name="numTags">the number of tags to be applied to the client. Must be less less than or equal to NEXT_MAX_TAGS</param>
		* <example>
		* <code>
		*	string[] tags = new string[]{"pro", "streamer"};
		*	Next.NextServerTagSessionMultiple(server, clientAddrPtr, tags, tags.Length);
		* </code>
		* </example>
		*/
		public static void NextServerTagSessionMultiple(IntPtr server, IntPtr addressPtr, string[] tags, int numTags)
		{
			if (tags.Length > NEXT_MAX_TAGS) {
				string[] maxTags = new string[NEXT_MAX_TAGS];
				for (int i = 0; i < NEXT_MAX_TAGS; i++) {
					maxTags[i] = tags[i];
				}
				next_server_tag_session_multiple(server, addressPtr, maxTags, NEXT_MAX_TAGS);
				return;
			}

			next_server_tag_session_multiple(server, addressPtr, tags, numTags);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_session_upgraded", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern NEXT_BOOL next_server_session_upgraded(IntPtr server, IntPtr addressPtr);

		/**
		* <summary>
		* 	Determines if a session has been upgraded.
		* </summary>
		* <param name="server">the server pointer</param>
		* <param name="addressPtr">the pointer to the <see cref="NextAddress"/> of the client to be checked</param>
		* <returns><see langword="true"/> if the session has been upgraded, otherwise <see langword="false"/>.</returns>
		* <example>
		* <code>
		*	bool upgraded = Next.NextServerSessionUpgraded(server, clientAddrPtr);
		*	Debug.Log(String.Format("session upgraded = {0}", upgraded ? "true" : "false"));
		* </code>
		* </example>
		*/
		public static bool NextServerSessionUpgraded(IntPtr server, IntPtr addressPtr)
		{
			NEXT_BOOL upgraded = next_server_session_upgraded(server, addressPtr);
			return upgraded == NEXT_BOOL.NEXT_TRUE;
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_send_packet", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_send_packet(IntPtr server, IntPtr toAddress, byte[] packetData, int packetBytes);

		/**
		* <summary>
		*	Send a packet to a client.
		* 	<remarks>
		*		Sends a packet to a client. If the client is upgraded and accelerated by network next, the packet will be sent across our private network of networks.
		*		Otherwise, the packet will be sent across the public internet.
		*	</remarks>
		* </summary>
		* <param name="server">the server pointer</param>
		* <param name="toAddress">the pointer to the <see cref="NextAddress"/> of the cleint to send the packet to</param>
		* <param name="packetData">the pakcet data to send</param>
		* <param name="packetBytes">the size of the packet. Must be in the range 1 to NEXT_MTU (1300)</param>
		* <example>
		* <code>
		*	[MonoPInvokeCallback(typeof(NextServerPacketReceivedCallback))]
		*	static void ServerPacketReceived(IntPtr serverPtr, IntPtr ctxPtr, IntPtr NextAddress fromPtr, IntPtr packetDataPtr, int packetBytes)
		*	{
		*		Next.NextAddress clientAddr = Next.GetNextAddressFromPointer(fromPtr);
		*		Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server received packet from client {0} ({0} bytes)", Next.NextAddressToString(ref clientAddr, Next.NEXT_MAX_ADDRESS_STRING_LENGTH), packetBytes));
		*		// ... verify the player is real ...
		*
		*		// Unmarshal the packet data
		*       byte[] packetData = new byte[packetBytes];
		*       Marshal.Copy(packetDataPtr, packetData, 0, packetBytes);
		*
		*		// Reflect the packet data back to the client
		*		Next.NextServerSendPacket(serverPtr, fromPtr, packetData, packetBytes);		
		*	}
		* </code>
		* </example
		*/
		public static void NextServerSendPacket(IntPtr server, IntPtr toAddress, byte[] packetData, int packetBytes)
		{
			next_server_send_packet(server, toAddress, packetData, packetBytes);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_send_packet_direct", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_send_packet_direct(IntPtr server, IntPtr toAddress, byte[] packetData, int packetBytes);

		/**
		* <summary>
		*	Send a packet to a client, forcing the packet to be sent over the public internet.
		* 	<remarks>
		*		This function is useful when you need to send non-latency sensitive packets to the client, for example, during a load screen.
		*		Packets sent via this function do not apply to your network next bandwidth envelope.
		*	</remarks>
		* </summary>
		* <param name="server">the server pointer</param>
		* <param name="toAddress">the pointer to the <see cref="NextAddress"/> of the cleint to send the packet to</param>
		* <param name="packetData">the pakcet data to send</param>
		* <param name="packetBytes">the size of the packet. Must be in the range 1 to NEXT_MTU (1300)</param>
		* <example>
		* <code>
		*	[MonoPInvokeCallback(typeof(NextServerPacketReceivedCallback))]
		*	static void ServerPacketReceived(IntPtr serverPtr, IntPtr ctxPtr, IntPtr NextAddress fromPtr, IntPtr packetDataPtr, int packetBytes)
		*	{
		*		Next.NextAddress clientAddr = Next.GetNextAddressFromPointer(fromPtr);
		*		Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server received packet from client {0} ({0} bytes)", Next.NextAddressToString(ref clientAddr, Next.NEXT_MAX_ADDRESS_STRING_LENGTH), packetBytes));
		*		// ... verify the player is real ...
		*
		*		// Unmarshal the packet data
		*       byte[] packetData = new byte[packetBytes];
		*       Marshal.Copy(packetDataPtr, packetData, 0, packetBytes);
		*
		*		// Reflect the packet data back to the client over the direct route
		*		Next.NextServerSendPacketDirect(serverPtr, fromPtr, packetData, packetBytes);		
		*	}
		* </code>
		* </example
		*/
		public static void NextServerSendPacketDirect(IntPtr server, IntPtr toAddress, byte[] packetData, int packetBytes)
		{
			next_server_send_packet_direct(server, toAddress, packetData, packetBytes);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_stats", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern NEXT_BOOL next_server_stats(IntPtr server, IntPtr addressPtr, out ServerStatsInternal stats);

		/**
		* <summary>
		*	Gets server statistics for a client.
		* </summary>
		* <param name="server">the server pointer</param>
		* <param name="addressPtr">the pointer to the <see cref="NextAddress"/> of the client to get statistics of</param>
		* <param name="stats">the <see cref="ServerStats"/> struct to fill with statistics</param>
		* <returns>a <see cref="ClientStats"/> struct.</returns>
		* <example>
		* 	Here is how to query it, and print out various interesting values.
		* <code>
		*	StringBuilder sb = new StringBuilder();
		*	
		*	// Create IntPtr for the address
		*	IntPtr addrPtr = Marshal.AllocHGlobal(Marshal.SizeOf(address));
		*	Marshal.StructureToPtr(address, addrPtr, false);
		*	
		*	// Get the stats for this address
		*	Next.ServerStats stats;
		*	if (!Next.NextServerStats(serverPtr, addrPtr, out stats))
		*	{
		*		Next.NextPrintf(Next.NEXT_LOG_LEVEL_DEBUG, "server does not contain a session for the provided address");
		*		return;
		*	}
		*	
		*	// Release memory for the address pointer
		*	Marshal.FreeHGlobal(addrPtr);
		*	
		*	sb.Append("================================================================\n");
		*	sb.AppendFormat("address = {0}\n", Next.NextAddressToString(ref address));
		*	
		*	string platform = "unknown";
		*	switch (stats.PlatformID)
		*	{
		*		case Next.NEXT_PLATFORM_WINDOWS:
		*	        platform = "windows";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_MAC:
		*	        platform = "mac";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_LINUX:
		*	        platform = "linux";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_SWITCH:
		*	        platform = "nintendo switch";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_PS4:
		*	        platform = "ps4";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_PS5:
		*	        platform = "ps5";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_IOS:
		*	        platform = "ios";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_XBOX_ONE:
		*	        platform = "xbox one";
		*	        break;
		*	
		*	    case Next.NEXT_PLATFORM_XBOX_SERIES_X:
		*	        platform = "xbox series x";
		*	        break;
		*	
		*	    default:
		*	        break;
		*	}
		*	
		*	sb.AppendFormat("session id = {0}\n", stats.SessionID.ToString());
		*	sb.AppendFormat("platform id = {0} ({1})\n", platform, stats.PlatformID);
		*	
		*	string connection = "unknown";
		*	switch (stats.ConnectionType)
		*	{
		*		case Next.NEXT_CONNECTION_TYPE_WIRED:
		*		    connection = "wired";
		*		    break;
		*	
		*		case Next.NEXT_CONNECTION_TYPE_WIFI:
		*		    connection = "wifi";
		*		    break;
		*	
		*		case Next.NEXT_CONNECTION_TYPE_CELLULAR:
		*		    connection = "cellular";
		*		    break;
		*	
		*		default:
		*		    break;
		*	}
		*	
		*	sb.AppendFormat("connection type = {0} ({1})\n", connection, stats.ConnectionType);
		*	
		*	if (!stats.FallbackToDirect)
		*	{
		*		sb.AppendFormat("committed = {0}\n", stats.Committed.ToString());
		*		sb.AppendFormat("multipath = {0}\n", stats.Multipath.ToString());
		*		sb.AppendFormat("reported = {0}\n", stats.Reported.ToString());
		*	}
		*	
		*	sb.AppendFormat("fallback to direct = {0}\n", stats.FallbackToDirect.ToString());
		*	
		*	sb.AppendFormat("direct min rtt = {0}ms\n", stats.DirectMinRTT.ToString("F"));
		*	sb.AppendFormat("direct max rtt = {0}ms\n", stats.DirectMaxRTT.ToString("F"));
		*	sb.AppendFormat("direct prime rtt = {0}ms\n", stats.DirectPrimeRTT.ToString("F"));
		*	sb.AppendFormat("direct jitter = {0}ms\n", stats.DirectJitter.ToString("F"));
		*	sb.AppendFormat("direct packet loss = {0}%\n", stats.DirectPacketLoss.ToString("F1"));
		*	
		*	if (stats.Next)
		*	{
		*		sb.AppendFormat("next rtt = {0}ms\n", stats.NextRTT.ToString("F"));
		*		sb.AppendFormat("next jitter = {0}ms\n", stats.NextJitter.ToString("F"));
		*		sb.AppendFormat("next packet loss = {0}%\n", stats.NextPacketLoss.ToString("F1"));
		*		sb.AppendFormat("next bandwidth up = {0}kbps\n", stats.NextKbpsUp.ToString("F1"));
		*		sb.AppendFormat("next bandwidth down = {0}kbps\n", stats.NextKbpsDown.ToString("F1"));
		*	}
		*	
		*	if (!stats.FallbackToDirect)
		*	{
		*		sb.AppendFormat("packets sent client to server = {0}\n", stats.PacketsSentClientToServer);
		*		sb.AppendFormat("packets sent server to client = {0}\n", stats.PacketsSentServerToClient);
		*		sb.AppendFormat("packets lost client to server = {0}\n", stats.PacketsLostClientToServer);
		*		sb.AppendFormat("packets lost server to client = {0}\n", stats.PacketsLostServerToClient);
		*		sb.AppendFormat("packets out of order client to server = {0}\n", stats.PacketsOutOfOrderClientToServer);
		*		sb.AppendFormat("packets out of order server to client = {0}\n", stats.PacketsOutOfOrderServerToClient);
		*		sb.AppendFormat("jitter client to server = {0}\n", stats.JitterClientToServer.ToString("F"));
		*		sb.AppendFormat("jitter server to client = {0}\n", stats.JitterServerToClient.ToString("F"));
		*	}
		*	
		*	if (stats.NumTags > 0)
		*	{
		*		sb.Append("tags = [");
		*		for (int j = 0; j < stats.NumTags; j++)
		*		{
		*			if (j < stats.NumTags - 1)
		*			{
		*				sb.AppendFormat("{0},", stats.Tags[j].ToString());
		*			}
		*			else
		*			{
		*				sb.AppendFormat("{0}", stats.Tags[j].ToString());
		*			}
		*		}
		*		sb.Append("]\n");
		*	}
		*	
		*	sb.Append("================================================================\n");
		*	
		*	Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, sb.ToString());
		* </code>
		* </example
		*/
		public static bool NextServerStats(IntPtr server, IntPtr addressPtr, out ServerStats stats)
		{
			stats = new ServerStats();
			ServerStatsInternal internalStats;
			NEXT_BOOL sessionExists = next_server_stats(server, addressPtr, out internalStats);
			if (sessionExists == NEXT_BOOL.NEXT_FALSE)
			{
				return false;
			}

			stats.Address = internalStats.Address;
			stats.SessionID = internalStats.SessionID;
			stats.UserHash = internalStats.UserHash;
			stats.PlatformID = internalStats.PlatformID;
			stats.ConnectionType = internalStats.ConnectionType;
			stats.Next = internalStats.Next == NEXT_BOOL.NEXT_TRUE;
			stats.Committed = internalStats.Committed == NEXT_BOOL.NEXT_TRUE;
			stats.Multipath = internalStats.Multipath == NEXT_BOOL.NEXT_TRUE;
			stats.Reported = internalStats.Reported == NEXT_BOOL.NEXT_TRUE;
			stats.FallbackToDirect = internalStats.FallbackToDirect == NEXT_BOOL.NEXT_TRUE;
			stats.DirectMinRTT = internalStats.DirectMinRTT;
			stats.DirectMinRTT = internalStats.DirectMaxRTT;
			stats.DirectPrimeRTT = internalStats.DirectPrimeRTT;
			stats.DirectJitter = internalStats.DirectJitter;
			stats.DirectPacketLoss = internalStats.DirectPacketLoss;
			stats.NextRTT = internalStats.NextRTT;
			stats.NextJitter = internalStats.NextJitter;
			stats.NextPacketLoss = internalStats.NextPacketLoss;
			stats.NextKbpsUp = internalStats.NextKbpsUp;
			stats.NextKbpsDown = internalStats.NextKbpsDown;
			stats.PacketsSentClientToServer = internalStats.PacketsSentClientToServer;
			stats.PacketsSentServerToClient = internalStats.PacketsSentServerToClient;
			stats.PacketsLostClientToServer = internalStats.PacketsLostClientToServer;
			stats.PacketsLostServerToClient = internalStats.PacketsLostServerToClient;
			stats.PacketsOutOfOrderClientToServer = internalStats.PacketsOutOfOrderClientToServer;
			stats.PacketsOutOfOrderServerToClient = internalStats.PacketsOutOfOrderServerToClient;
			stats.JitterClientToServer = internalStats.JitterClientToServer;
			stats.JitterServerToClient = internalStats.JitterServerToClient;
			stats.NumTags = internalStats.NumTags;
			stats.Tags = internalStats.Tags;

			return true;			
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_autodetect_finished", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern NEXT_BOOL next_server_autodetect_finished(IntPtr server);

		/**
		* <summary>
		*	Determines if the server has finished autodetecting the datacenter name after calling <see cref="NextServerCreate"/>.
		* </summary>
		* <param name="server">the server pointer</param>
		* <returns><see langword="true"/> if the server has finished autodetecting its datacenter, otherwise <see langword="false"/>.</returns>
		* <example>
		* <code>
		*	bool autodetectFinished = Next.NextServerAutodetectFinished(server);
		*	Debug.Log(String.Format("server autodetect finished = {0}", autodetectFinished ? "true" : "false"));
		* </code>
		* </example>
		*/
		public static bool NextServerAutodetectFinished(IntPtr server)
		{
			NEXT_BOOL autodetectFinished = next_server_autodetect_finished(server);
			return autodetectFinished == NEXT_BOOL.NEXT_TRUE;
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_autodetected_datacenter", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern IntPtr next_server_autodetected_datacenter(IntPtr server);

		/**
		* <summary>
		*	Gets the autodetected datacenter name.
		* </summary>
		* <param name="server">the server pointer</param>
		* <returns>the name of the autodetected datacenter.</returns>
		* <example>
		* <code>
		*	bool autodetectFinished = Next.NextServerAutodetectFinished(server);
		*	if (autodetectFinished)
		*	{
		*		string autodetectedDatacenter = Next.NextServerAutodetectedDatacenter(server);
		*		Debug.Log(String.Format("server autodetected datacenter = {0}", autodetectedDatacenter));
		*	}
		* </code>
		* </example>
		*/
		public static string NextServerAutodetectedDatacenter(IntPtr server)
		{
			return Marshal.PtrToStringAnsi(next_server_autodetected_datacenter(server));
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_event", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_event(IntPtr server, IntPtr addressPtr, ulong serverEvents);

		/**
		* <summary>
		*	Triggers a user-defined event on a session. This event is stored alongside network performance data once every 10 seconds.
		*	<remarks>
		*		You can define up to 64 event flags for your game, one event per bit in the <paramref name="serverEvents"/> bitfield.
		*		Use this function to input in-game events that may be relevant to analytics.
		*	</remarks>
		* </summary>
		* <param name="server">the server pointer</param>
		* <param name="addressPtr">the pointer to the address of the client that triggered the event</param>
		* <param name="serverEvents">the bitfield of events that just triggered for the session</param>
		* <example>
		* <code>
		*	public enum GameEvents : ulong
		*	{
		*		Respawned = (1<<0),
		*		Catch = (1<<1),
		*		Throw = (1<<2),
		*		KnockedOut = (1<<3),
		*		WonMatch = (1<<4),
		*		LostMatch = (1<<5),
		*	}
		*	
		*	Next.NextServerEvent(server, clientAddrPtr, (ulong)(GameEvents.KnockedOut) | (ulong)(GameEvents.LostMatch));
		* </code>
		* </example>
		*/
		public static void NextServerEvent(IntPtr server, IntPtr addressPtr, ulong serverEvents)
		{
			next_server_event(server, addressPtr, serverEvents);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_match", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_match(IntPtr server, IntPtr addressPtr, [MarshalAs(UnmanagedType.LPStr)] string matchID, double[] matchValues, int numMatchValues);

		/**
		* <summary>
		*	Associates a session with a match id and set of match values for that session.
		*	<remarks>
		*		Match id can be any unique match id you have.
		*		Match values can include any information that you want to feed into analytics.
		*		For example: win/loss ratio, skill, kill/death ratio, skill, time spent in matchmaker, load time in seconds.
		*		Call this function once per-session at the beginning of each match on the server.
		*	</remarks>
		* </summary>
		* <param name="server">the server pointer</param>
		* <param name="addressPtr">the pointer to the <see cref="NextAddress"/> of the client to assign match data</param>
		* <param name="matchID">the match id to assign to the session. Pass in any unique per-match identifier you have</param>
		* <param name="matchValues">the array of match values for the session</param>
		* <param name="numMatchValues">the number of match values in the array. Must be less less than or equal to NEXT_MAX_MATCH_VALUES</param>
		* <example>
		* <code>
		*	string matchID = "this is a unique match id";
		*	double[] matchValues = new double[]{10.0, 20.0, 30.0};
		*	Next.NextServerMatch(server, clientAddrPtr, matchID, matchValues, matchValues.Length);
		* </code>
		* </example>
		*/
		public static void NextServerMatch(IntPtr server, IntPtr addressPtr, string matchID, double[] matchValues, int numMatchValues)
		{
			if (matchValues.Length > NEXT_MAX_MATCH_VALUES) {
				double[] maxMatchValues = new double[NEXT_MAX_MATCH_VALUES];
				for (int i = 0; i < NEXT_MAX_MATCH_VALUES; i++) {
					maxMatchValues[i] = matchValues[i];
				}
				next_server_match(server, addressPtr, matchID, maxMatchValues, NEXT_MAX_MATCH_VALUES);
				return;
			}

			next_server_match(server, addressPtr, matchID, matchValues, numMatchValues);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_server_flush", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_server_flush(IntPtr server);

		/**
		* <summary>
		*	Call this to flush all server data before shutting a server down.
		*	<remarks>
		*		This function blocks for up to 10 seconds to ensure that all session data, server events and match data are recorded.
		*		After calling this function, destroy the server via <see cref="NextServerDestroy"/>.
		*	</remarks>
		* </summary>
		* <param name="server">the server pointer</param>
		* <example>
		* <code>
		*	Next.NextServerFlush(server);
		*	Next.NextServerDestroy(server);
		* </code>
		* </example>
		*/
		public static void NextServerFlush(IntPtr server)
		{
			next_server_flush(server);
		}

		#endregion // #region NextServer functions

		// ----------------------------------------------------------

		// Network Next mutex

		#region NextMutex definition
		/**
		* <summary>
		*	Next Mutex struct for the Network Next SDK.
		*	This is a struct that represents a platform independent mutex.
		*	<remarks>
			*	Use this when you have custom classes that can be called from multiple threads.
		*	</remarks>
		* </summary>
		* <value <see cref="Dummy"/> - The <see cref="byte"/> array representing the mutex. No need to access this field.</value>
		*/
		[StructLayout (LayoutKind.Sequential)]
		public struct NextMutex
		{
			[MarshalAs(UnmanagedType.ByValArray, SizeConst = NEXT_MUTEX_BYTES)]
			public byte[] Dummy;
		}

		#endregion // #region NextMutex definition

		// Network Next mutex functions

		#region NextMutex functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_mutex_create", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern int next_mutex_create(out NextMutex mutex);

		/**
		* <summary>
		* 	Creates a <see cref="NextMutex"/>.
		* <summary>
		* <param name="mutex">the mutex object to assign and create</param>
		* <example>
		*	This example shows how to create a mutex.
		* <code>
		*	Next.NextMutex mutex;
		* 	Next.NextMutexCreate(out mutex);
		*	// Mutex can now be used for acquire, release, and destroy operations
		* </code>
		* </example>
		*/
		public static int NextMutexCreate(out NextMutex mutex)
		{
			return next_mutex_create(out mutex);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_mutex_destroy", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_mutex_destroy(ref NextMutex mutex);

		/**
		* <summary>
		*	Destroys a <see cref="NextMutex"/>.
		* </summary>
		* <param name="mutex">the mutex to destroy</param>
		* <example>
		*	// Create a mutex
		*	Next.NextMutex mutex;
		* 	Next.NextMutexCreate(out mutex);
		*
		*	// Destroy a mutex
		*	Next.NextMutexDestroy(ref mutex);
		* </code>
		* </example>
		*/
		public static void NextMutexDestroy(ref NextMutex mutex)
		{
			next_mutex_destroy(ref mutex);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_mutex_acquire", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_mutex_acquire(ref NextMutex mutex);

		/**
		* <summary>
		*	Acquires a lock from a <see cref="NextMutex"/>.
		* </summary>
		* <param name="mutex">the mutex to lock</param>
		* <example>
		*	// Create a mutex
		*	Next.NextMutex mutex;
		* 	Next.NextMutexCreate(out mutex);
		*
		*	// Lock the mutex
		*	Next.NextMutexAcquire(ref mutex);
		*	
		*	// ... do something thread safe ...
		* </code>
		* </example>
		*/
		public static void NextMutexAcquire(ref NextMutex mutex)
		{
			next_mutex_acquire(ref mutex);
		}

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_mutex_release", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_mutex_release(ref NextMutex mutex);

		/**
		* <summary>
		*	Release a lock from a <see cref="NextMutex"/>.
		* </summary>
		* <param name="mutex">the mutex to unlock</param>
		* <example>
		*	// Create a mutex
		*	Next.NextMutex mutex;
		* 	Next.NextMutexCreate(out mutex);
		*
		*	// Lock the mutex
		*	Next.NextMutexAcquire(ref mutex);
		*	
		*	// ... do something thread safe ...
		*	
		*	// Unlock the mutex
		*	Next.NextMutexRelease(ref mutex);
		*
		*	// Destroy the mutex
		*	Next.NextMutexDestroy(ref mutex);
		* </code>
		* </example>
		*/
		public static void NextMutexRelease(ref NextMutex mutex)
		{
			next_mutex_release(ref mutex);
		}

		#endregion // #region NextMutex functions

		// ----------------------------------------------------------

		// Network Next unit test functions

		#region UnitTest functions

		[DllImport(dll, CallingConvention = CallingConvention.Cdecl, EntryPoint = "next_test", CharSet = CharSet.Ansi, ExactSpelling = true)]
		private static extern void next_test();

		/**
		* <summary>
		* 	Run the Network Next SDK unit tests.
		* </summary>
		*/
		public static void NextTest()
		{
			next_test();
		}

		#endregion // #region UnitTest functions
	}

	#endregion // #region Next definition
}
