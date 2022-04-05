using System;
using System.Runtime.InteropServices;
using UnityEngine;
using UnityEngine.Networking;
using AOT;
using NetworkNext;

public class SimpleNetworkManager : NetworkManager
{

    IntPtr server;
    NextServerTransport serverTransport;

    // ----------------------------------------------------------

    // Delegate functions

    enum Color { red, green, blue, black, white, yellow, orange };


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
        Next.NextAddress clientAddr = Next.GetNextAddressFromPointer(fromPtr);
        Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server received packet from client {0} ({1} bytes)", Next.NextAddressToString(ref clientAddr), packetBytes));
    }

    public virtual void Start()
    {
        Next.NextLogFunction(UnityLogger);
        Next.NextLogLevel(Next.NEXT_LOG_LEVEL_DEBUG);
    }

    // Update is called once per frame
    public virtual void Update()
    {
        serverTransport?.NextServerUpdate();
    }

    public override void OnStartHost()
    {
        base.OnStartHost();
        Debug.Log("CALLBACK: STARTED HOSTING");
    }

    public override void OnStartServer()
    {
        base.OnStartServer();
        Debug.Log("CALLBACK: STARTED SERVER");

        if (server == null || server == IntPtr.Zero)
        {
            // Get empty config
            Next.NextConfig config = new Next.NextConfig();

            // Create the packet received callback
            NextServerPacketReceivedCallback recvCallBack = new NextServerPacketReceivedCallback(ServerPacketReceived);

            NextServerTransport nextTransport = new NextServerTransport(IntPtr.Zero, ref config, "127.0.0.1:50000", "0.0.0.0:50000", "local", recvCallBack, null);
            activeTransport = nextTransport;
            serverTransport = nextTransport;
            nextTransport.Init();
            server = nextTransport.server;
            Debug.Log("finished setting up server");
        }
    }

    public override void OnServerAddPlayer(NetworkConnection conn, short playerControllerId)
    {
        GameObject player;

        player = (GameObject)UnityEngine.Object.Instantiate(this.playerPrefab, new Vector3(-3, 2, -3), Quaternion.identity);
        NetworkServer.AddPlayerForConnection(conn, player, playerControllerId);
    }

    public override void OnClientSceneChanged(NetworkConnection conn)
    {
        ClientScene.AddPlayer(conn, 0);
    }

    // We override this method and do not call the base method
    // because the Network Manager auto creates the player
    // and we do not want the server to spawn in the player twice
    // (avoids "A connection has already been set as ready." error).
    public override void OnClientConnect(NetworkConnection conn)
    {
        // base.OnClientConnect(conn);
        // Debug.Log("CALLBACK: CLIENT HAS CONNECTED TO SERVER " + conn);
    }

    public override void OnStopServer()
    {
        base.OnStopServer();
        // Debug.Log("CALLBACK: STOPPING SERVER");
        serverTransport.NextServerFlush();
        serverTransport.NextServerDestroy();
        server = IntPtr.Zero;
    }

    // Temporary for use in editor to ensure clean shutdown of transport
    void OnApplicationQuit()
    {
        serverTransport.Shutdown();
        server = IntPtr.Zero;
    }
}
