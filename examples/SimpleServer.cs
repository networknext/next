using System;
using System.Runtime.InteropServices;
using UnityEngine;
using AOT;
using NetworkNext;

public class SimpleServer : MonoBehaviour
{
	// Constants
	const string bindAddress = "0.0.0.0:0";
	const string serverAddress = "127.0.0.1:50000";
	const string serverDatacenter = "local";

	enum Color { red, green, blue, black, white, yellow, orange };

	// Global variables
	IntPtr server;

	// ----------------------------------------------------------

	// Delegate functions

	// Create custom logging function to output to Unity console
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
			Debug.Log(String.Format("<color={0}>{1}</color>", c.ToString(), argsStr));
		}
	}

	// Define packet receive callback function
	[MonoPInvokeCallback(typeof(NextServerPacketReceivedCallback))]
	public void ServerPacketReceived(IntPtr serverPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
	{
		// Unmarshal the packet data into byte[]
		byte[] packetData = new byte[packetBytes];
		Marshal.Copy(packetDataPtr, packetData, 0, packetBytes);

		Next.NextServerSendPacket(serverPtr, fromPtr, packetData, packetBytes);
		Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server received packet from client ({0} bytes)", packetBytes));
	}

	// ----------------------------------------------------------

	// Start is called before the first frame update
	void Start()
	{
		// Assign our custom logging function
		Next.NextLogFunction(UnityLogger);

		// Create a configuration
		Next.NextConfig config = new Next.NextConfig();

		// Initialize Network Next
		if (Next.NextInit(IntPtr.Zero, ref config) != Next.NEXT_OK)
		{
			Debug.LogError("error: could not initialize network next");
			this.gameObject.SetActive(false);
			return;
		}

		// Create the packet received callback
		NextServerPacketReceivedCallback recvCallBack = new NextServerPacketReceivedCallback(ServerPacketReceived);

		// Create a pointer to the server (store as global var)
		server = Next.NextServerCreate(IntPtr.Zero, serverAddress, bindAddress, serverDatacenter, recvCallBack, null);
		if (server == IntPtr.Zero)
		{
			Debug.LogError("error: failed to create server");
			this.gameObject.SetActive(false);
			return;
		}
	}

	// Update is called once per frame
	void Update()
	{
		Next.NextServerUpdate(server);
	}

	// OnApplicationQuit is called when the application quits or when playmode is stopped in the editor
	// These actions should be done in Destroy() rather than when the application quits
	void OnApplicationQuit()
	{
		// Destroy the server
		Next.NextServerDestroy(server);

		// Shut down the SDK
		Next.NextTerm();
	}
}
