using System;
using System.Runtime.InteropServices;
using UnityEngine;
using UnityEngine.Networking;
using AOT;
using NetworkNext;

public class UpgradedUNETClient : NetworkBehaviour
{
    // Constants
    const string bindIP = "0.0.0.0";
    const int bindPort = 0;
    const string serverIP = "127.0.0.1";
    const int serverPort = 50000;
    const int unetPort = 7777;
    const int hostID = 0;
    const string customerPublicKey = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="; // Replace with the public key from your account

    enum Color { red, green, blue, black, white, yellow, orange };

    // Global variables
    NextClientTransport clientTransport;
    int connectionID;

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
    static void ClientPacketReceived(IntPtr clientPtr, IntPtr ctxPtr, IntPtr fromPtr, IntPtr packetDataPtr, int packetBytes)
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
        if (isLocalPlayer)
        {
            // Assign our custom logging function
            Next.NextLogFunction(UnityLogger);

            // Get the default configuration
            Next.NextConfig config;
            Next.NextDefaultConfig(out config);

            // Assign our public key to the configuration
            config.CustomerPublicKey = customerPublicKey;

            // Create the packet received callback
            NextClientPacketReceivedCallback recvCallBack = new NextClientPacketReceivedCallback(ClientPacketReceived);

            // Create the NextClientTransport, which sets up the Network Next client on its own socket independent of UNET
            clientTransport = new NextClientTransport(IntPtr.Zero, ref config, bindIP, bindPort, serverIP, serverPort, recvCallBack, null);
            clientTransport.Init();

            // Connect to the server via UNET
            byte error;
            connectionID = clientTransport.Connect(hostID, serverIP, unetPort, 0, out error);

            // Connect to the server via Network Next
            clientTransport.NextClientOpenSession();
        }
    }

    // Update is called once per frame
    void Update()
    {
        if (isLocalPlayer)
        {
            clientTransport.NextClientUpdate();

            // Create a packet to send to the server
            int packetBytes;
            byte[] packetData = GeneratePacket(out packetBytes);

            // Send the packet to the server over Network Next
            clientTransport.NextClientSendPacket(packetData, packetBytes);
        }
    }

    // OnDestroy is called when a Scene or game ends
    void OnDestroy()
    {
        if (isLocalPlayer)
        {
            // Disconnect from the server and close the session
            byte error;
            clientTransport.Disconnect(hostID, connectionID, out error);

            // Destroy the client
            clientTransport.NextClientDestroy();
        }
    }

    // OnApplicationQuit is called when the application quits or when playmode is stopped in the editor
    void OnApplicationQuit()
    {
        // Shutdown the transport and the Network Next SDK
        clientTransport?.Shutdown();
    }
}
