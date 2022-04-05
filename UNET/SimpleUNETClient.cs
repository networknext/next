using System;
using System.Runtime.InteropServices;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.Networking;
using AOT;
using NetworkNext;

public class UNETClient : NetworkBehaviour
{
    // Constants
    const string bindAddress = "0.0.0.0:0";
    const string serverIP = "127.0.0.1";
    const int serverPort = 50000;
    const int unetPort = 7777;
    const int hostID = 0;
    const int channelID = 0;
    const string customerPublicKey = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="; // Replace with the public key from your account

    enum Color { red, green, blue, black, white, yellow, orange };

    // Global variables
    IntPtr client;
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

            if (client == null || client == IntPtr.Zero)
            {
                // Get empty config
                Next.NextConfig config = new Next.NextConfig();

                // Create the packet received callback
                NextClientPacketReceivedCallback recvCallBack = new NextClientPacketReceivedCallback(ClientPacketReceived);

                string serverAddress = String.Format("{0}:{1}", serverIP, serverPort);
                clientTransport = new NextClientTransport(IntPtr.Zero, ref config, bindAddress, serverAddress, recvCallBack, null);
                clientTransport.Init();
                client = clientTransport.client;
                Debug.Log("finished setting up client");

                byte error;
                connectionID = clientTransport.Connect(hostID, serverIP, unetPort, 0, out error);
                clientTransport.NextClientOpenSession();

                // Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, "Client connected: ", clientTransport.IsClientConnected());

                // Next.NextClientOpenSession(client, serverAddress);
                // Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, "client state is ", Next.NextClientState(client));
            }
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

            // Send the packet to the server potentially over Network Next
            clientTransport.NextClientSendPacket(packetData, packetBytes);
        }
    }

    void Destroy()
    {
        if (isLocalPlayer)
        {
            byte error;
            clientTransport.Disconnect(hostID, connectionID, out error);
        }
        clientTransport.NextClientDestroy();
    }

    void OnApplicationQuit()
    {
        clientTransport.Shutdown();
    }
}
