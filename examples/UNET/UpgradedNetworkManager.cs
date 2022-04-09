using System;
using System.Runtime.InteropServices;
using UnityEngine;
using UnityEngine.Networking;
using AOT;
using NetworkNext;

public class UpgradedNetworkManager : NetworkManager
{
    // Constants
    const string bindIP = "0.0.0.0";
    const int bindPort = 50000;
    const string serverIP = "127.0.0.1";
    const int serverPort = 50000;
    const int unetPort = 7777;
    const string serverDatacenter = "local";
    const string serverBackendHostname = "prod.spacecats.net";
    const string customerPrivateKey = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"; // Replace with the private key from your account

    enum Color { red, green, blue, black, white, yellow, orange };

    // Global variables
    bool serverReady;
    NextServerTransport serverTransport;

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

        serverTransport.NextServerSendPacket(fromPtr, packetData, packetBytes);
        Next.NextPrintf(Next.NEXT_LOG_LEVEL_INFO, String.Format("server received packet from client ({0} bytes)", packetBytes));

        if (!serverTransport.NextServerSessionUpgraded(fromPtr))
        {
            string userIDString = "12345";
            serverTransport.NextServerUpgradeSession(fromPtr, userIDString);
        }
    }

    // Start is called before the first frame update
    void Start()
    {
        // Assign our custom logging function to view Network Next SDK logs
        Next.NextLogFunction(UnityLogger);

        #if UNITY_SERVER
        // Set the UNET IP, port, and bind address
        this.networkAddress = serverIP;
        this.networkPort = unetPort;
        this.serverBindAddress = String.Format("{0}:{1}", bindIP, unetPort);

        // Bypass the Network Manager HUD and start the server
        this.StartServer();
        #endif // #if UNITY_SERVER
    }

    // Update is called once per frame
    void Update()
    {
        serverTransport?.NextServerUpdate();
    }

    // Override this method and do not call the base method
    // because the Network Manager creates the player in OnServerAddPlayer
    // and we do not want the server to spawn in the player twice
    // (avoids "A connection has already been set as ready." error).
    public override void OnClientConnect(NetworkConnection conn) {}

    // Adds the player to the scene upon scene change
    public override void OnClientSceneChanged(NetworkConnection conn)
    {
        ClientScene.AddPlayer(conn, 0);
    }

    // Sets up the game server
    public override void OnStartServer()
    {
        base.OnStartServer();

        // Configure the Network Next server
        if (!serverReady)
        {
            // Get the default configuration
            Next.NextConfig config;
            Next.NextDefaultConfig(out config);

            // Assign our private key and the server backend hostname to the configuration
            config.CustomerPrivateKey = customerPrivateKey;
            config.ServerBackendHostname = serverBackendHostname;

            // Create the packet received callback
            NextServerPacketReceivedCallback recvCallBack = new NextServerPacketReceivedCallback(ServerPacketReceived);

            // Create the NextServerTransport, which sets up the Network Next server on its own socket independent of UNET
            NextServerTransport nextTransport = new NextServerTransport(IntPtr.Zero, ref config, serverIP, serverPort, bindIP, bindPort, serverDatacenter, recvCallBack, null);

            // Create the NextServerTransport, which sets up the Network Next server on its own socket independent of UNET,
            // and save it as an instance var for use in other callbacks
            serverTransport = new NextServerTransport(IntPtr.Zero, ref config, serverIP, serverPort, bindIP, bindPort, serverDatacenter, recvCallBack, null);

            // Set the NextServerTransport as the active transport
            activeTransport = serverTransport;
            serverTransport.Init();

            // Set the server as ready
            serverReady = true;
        }
    }

    // Adds the player prefab object to the server
    public override void OnServerAddPlayer(NetworkConnection conn, short playerControllerId)
    {
        GameObject player;

        player = (GameObject)UnityEngine.Object.Instantiate(this.playerPrefab, new Vector3(0, 2, 0), Quaternion.identity);
        NetworkServer.AddPlayerForConnection(conn, player, playerControllerId);
    }

    // Cleans up the Network Next Server
    public override void OnStopServer()
    {
        base.OnStopServer();

        // Set the server as not ready
        serverReady = false;

        // Flush the server
        serverTransport.NextServerFlush();

        // Destroy the server
        serverTransport.NextServerDestroy();
    }

    // OnApplicationQuit is called when the application quits or when playmode is stopped in the editor
    void OnApplicationQuit()
    {
        // Set the server as not ready
        serverReady = false;

        // Shutdown the transport and Network Next SDK
        serverTransport?.Shutdown();
    }
}
