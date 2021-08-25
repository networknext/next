using System;
using System.Runtime.InteropServices;
using UnityEngine;
using AOT;
using NetworkNext;

public class SimpleClient : MonoBehaviour
{
	// Constants
	const string bindAddress = "0.0.0.0:0";
	const string serverAddress = "127.0.0.1:50000";

	enum Color { red, green, blue, black, white, yellow, orange };

	// Global variables
	IntPtr client;

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
    [MonoPInvokeCallback(typeof(NextClientPacketReceivedCallback))]
    static void ClientPacketReceived(IntPtr clientPtr, IntPtr ctxPtr, IntPtr packetDataPtr, int packetBytes)
    {
    	Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("client received packet from server ({0} bytes)", packetBytes));
    }

    // ----------------------------------------------------------

    // Utility function to generate a valid Network Next packet
    byte[] GeneratePacket(out int packetBytes) {
    	var rand = new System.Random();

    	packetBytes = 1 + (rand.Next() % Next.NEXT_MTU);

    	byte[] packetData = new byte[packetBytes];

    	int start = packetBytes % 256;
    	for (int i = 0; i < packetBytes; i++) {
    		packetData[i] = (byte)((start + i) % 256);
    	}

    	return packetData;
    }

    // ----------------------------------------------------------

    // Start is called before the first frame update
    void Start()
    {
		// Allow all logging messages to be displayed
	    Next.NextLogLevel(Next.NEXT_LOG_LEVEL_DEBUG);

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
        NextClientPacketReceivedCallback recvCallBack = new NextClientPacketReceivedCallback(ClientPacketReceived);

        // Create a pointer to the client (store as global var)
        client = Next.NextClientCreate(IntPtr.Zero, bindAddress, recvCallBack, null);
        if (client == IntPtr.Zero)
        {
        	Debug.LogError("error: failed to create client");
        	this.gameObject.SetActive(false);
        	return;
        }

        // Open a session to the server
        Next.NextClientOpenSession(client, serverAddress);
    }

    // Update is called once per frame
    void Update()
    {
        Next.NextClientUpdate(client);

        // Create a packet to send to the server
        int packetBytes;
        byte[] packetData = GeneratePacket(out packetBytes);

        // Send the packet to the server
        Next.NextClientSendPacket(client, packetData, packetBytes);
    }

    // OnApplicationQuit is called when the application quits or when playmode is stopped in the editor
	// These actions should be done in Destroy() rather than when the application quits
    void OnApplicationQuit()
    {
    	// Destroy the client
    	Next.NextClientDestroy(client);

    	// Shut down the SDK
    	Next.NextTerm();
    }
}
