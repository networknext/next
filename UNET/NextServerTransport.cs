using System;
using System.Net;
using UnityEngine.Networking;
using UnityEngine.Networking.Types;
using NetworkNext;

public class NextServerTransport : INetworkTransport
{
    public IntPtr server;
    public IntPtr serverCtxPtr;
    public Next.NextConfig nextConfig;
    public string serverAddress;
    public string serverBindAddress;
    public string serverDatacenter;
    public NextServerPacketReceivedCallback serverPacketReceivedCallback;
    public NextWakeupCallback serverWakeupCallback;

    public NextServerTransport(IntPtr ctxPtr, ref Next.NextConfig config, string serverIP, int serverPort, string bindIP, int bindPort, string datacenter, NextServerPacketReceivedCallback packetRecvCallback, NextWakeupCallback wakeupCallback = null)
    {
        serverCtxPtr = ctxPtr;
        nextConfig = config;
        serverAddress = String.Format("{0}:{1}", serverIP, serverPort);
        serverBindAddress = String.Format("{0}:{1}", bindIP, bindPort);
        serverDatacenter = datacenter;
        serverPacketReceivedCallback = packetRecvCallback;
        serverWakeupCallback = wakeupCallback;

        if (!IsServerReady())
        {
            if (Next.NextInit(serverCtxPtr, ref nextConfig) != Next.NEXT_OK)
            {
                throw new InvalidOperationException("could not initialize Network Next");
            }

            server = Next.NextServerCreate(serverCtxPtr, serverAddress, serverBindAddress, serverDatacenter, serverPacketReceivedCallback, serverWakeupCallback);
            if (server == null || server == IntPtr.Zero)
            {
                throw new InvalidOperationException("could not create server");
            }
        }
    }

    #region Utility Functions

    public bool IsServerReady()
    {
        return server != null && server != IntPtr.Zero;
    }

    #endregion // #region Utility Functions

    #region Next Server Functions

    public void NextServerDestroy()
    {
        if (IsServerReady())
        {
            Next.NextServerDestroy(server);
            server = IntPtr.Zero;
        }
    }

    public ushort NextServerPort()
    {
        if (IsServerReady())
        {
            return Next.NextServerPort(server);
        }
        return 0;
    }

    public Next.NextAddress NextServerAddress()
    {
        if (IsServerReady())
        {
            return Next.NextServerAddress(server);
        }
        return new Next.NextAddress();
    }

    public int NextServerState()
    {
        if (IsServerReady())
        {
            return Next.NextServerState(server);
        }
        return Next.NEXT_ERROR;
    }

    public void NextServerUpdate()
    {
        if (IsServerReady())
        {
            Next.NextServerUpdate(server);
        }
    }

    public ulong NextServerUpgradeSession(IntPtr addressPtr, string userID)
    {
        if (IsServerReady())
        {
            return Next.NextServerUpgradeSession(server, addressPtr, userID);
        }
        return 0;
    }

    public void NextServerTagSession(IntPtr addressPtr, string tag)
    {
        if (IsServerReady())
        {
            Next.NextServerTagSession(server, addressPtr, tag);
        }
    }

    public void NextServerTagSessionMultiple(IntPtr addressPtr, string[] tags, int numTags)
    {
        if (IsServerReady())
        {
            Next.NextServerTagSessionMultiple(server, addressPtr, tags, numTags);
        }
    }

    public bool NextServerSessionUpgraded(IntPtr addressPtr)
    {
        return IsServerReady() && Next.NextServerSessionUpgraded(server, addressPtr);
    }

    public void NextServerSendPacket(IntPtr clientAddr, byte[] buffer, int size)
    {
        if (IsServerReady())
        {
            Next.NextServerSendPacket(server, clientAddr, buffer, size);
        }
    }

    public void NextServerSendPacketDirect(IntPtr clientAddr, byte[] buffer, int size)
    {
        if (IsServerReady())
        {
            Next.NextServerSendPacketDirect(server, clientAddr, buffer, size);
        }
    }

    public bool NextServerStats(IntPtr addressPtr, out Next.ServerStats stats)
    {
        if (IsServerReady())
        {
            return Next.NextServerStats(server, addressPtr, out stats);
        }
        stats = new Next.ServerStats();
        return false;
    }

    public bool NextServerAutodetectFinished()
    {
        return IsServerReady() && Next.NextServerAutodetectFinished(server);
    }

    public string NextServerAutodetectedDatacenter()
    {
        if (IsServerReady())
        {
            return Next.NextServerAutodetectedDatacenter(server);
        }
        return serverDatacenter;
    }

    public void NextServerEvent(IntPtr addressPtr, ulong serverEvents)
    {
        if (IsServerReady())
        {
            Next.NextServerEvent(server, addressPtr, serverEvents);
        }
    }

    public void NextServerMatch(IntPtr addressPtr, string matchID, double[] matchValues, int numMatchValues)
    {
        if (IsServerReady())
        {
            Next.NextServerMatch(server, addressPtr, matchID, matchValues, numMatchValues);
        }
    }

    public void NextServerFlush()
    {
        if (IsServerReady())
        {
            Next.NextServerFlush(server);
        }
    }

    #endregion // #region Next Server Functions

    #region INetworkTransport Functions

    public bool IsStarted { get; }

    public int AddHost(HostTopology topology, int port, string ip)
    {
        return NetworkTransport.AddHost(topology, port, ip);
    }

    public int AddHostWithSimulator(HostTopology topology, int minTimeout, int maxTimeout, int port)
    {
        return NetworkTransport.AddHostWithSimulator(topology, minTimeout, maxTimeout, port);
    }

    public int AddWebsocketHost(HostTopology topology, int port, string ip)
    {
        return NetworkTransport.AddWebsocketHost(topology, port, ip);
    }

    public int Connect(int hostId, string address, int port, int specialConnectionId, out byte error)
    {
        return NetworkTransport.Connect(hostId, address, port, specialConnectionId, out error);
    }

    public void ConnectAsNetworkHost(int hostId, string address, int port, NetworkID network, SourceID source, NodeID node, out byte error)
    {
        NetworkTransport.ConnectAsNetworkHost(hostId, address, port, network, source, node, out error);
    }

    public int ConnectEndPoint(int hostId, EndPoint endPoint, int specialConnectionId, out byte error)
    {
        return NetworkTransport.ConnectEndPoint(hostId, endPoint, specialConnectionId, out error);
    }

    public int ConnectToNetworkPeer(int hostId, string address, int port, int specialConnectionId, int relaySlotId, NetworkID network, SourceID source, NodeID node, out byte error)
    {
        return NetworkTransport.ConnectToNetworkPeer(hostId, address, port, specialConnectionId, relaySlotId, network, source, node, out error);
    }

    public int ConnectWithSimulator(int hostId, string address, int port, int specialConnectionId, out byte error, ConnectionSimulatorConfig conf)
    {
        return NetworkTransport.ConnectWithSimulator(hostId, address, port, specialConnectionId, out error, conf);
    }

    public bool Disconnect(int hostId, int connectionId, out byte error)
    {
        return NetworkTransport.Disconnect(hostId, connectionId, out error);
    }

    public bool DoesEndPointUsePlatformProtocols(EndPoint endPoint)
    {
        throw new NotImplementedException("NextServerTransport does not support DoesEndPointUsePlatformProtocols");
    }

    public void GetBroadcastConnectionInfo(int hostId, out string address, out int port, out byte error)
    {
        NetworkTransport.GetBroadcastConnectionInfo(hostId, out address, out port, out error);
    }

    public void GetBroadcastConnectionMessage(int hostId, byte[] buffer, int bufferSize, out int receivedSize, out byte error)
    {
        NetworkTransport.GetBroadcastConnectionMessage(hostId, buffer, bufferSize, out receivedSize, out error);
    }

    public void GetConnectionInfo(int hostId, int connectionId, out string address, out int port, out NetworkID network, out NodeID dstNode, out byte error)
    {
        NetworkTransport.GetConnectionInfo(hostId, connectionId, out address, out port, out network, out dstNode, out error);
    }

    public int GetCurrentRTT(int hostId, int connectionId, out byte error)
    {
        return NetworkTransport.GetCurrentRTT(hostId, connectionId, out error);
    }

    public void Init() {

        NetworkTransport.Init();
    }

    public void Init(GlobalConfig config)
    {
        NetworkTransport.Init(config);
    }

    public NetworkEventType Receive(out int hostId, out int connectionId, out int channelId, byte[] buffer, int bufferSize, out int receivedSize, out byte error)
    {
        return NetworkTransport.Receive(out hostId, out connectionId, out channelId, buffer, bufferSize, out receivedSize, out error);
    }

    public NetworkEventType ReceiveFromHost(int hostId, out int connectionId, out int channelId, byte[] buffer, int bufferSize, out int receivedSize, out byte error)
    {
        return NetworkTransport.ReceiveFromHost(hostId, out connectionId, out channelId, buffer, bufferSize, out receivedSize, out error);
    }

    public NetworkEventType ReceiveRelayEventFromHost(int hostId, out byte error)
    {
        return NetworkTransport.ReceiveRelayEventFromHost(hostId, out error);
    }

    public bool RemoveHost(int hostId)
    {
        return NetworkTransport.RemoveHost(hostId);
    }

    public bool Send(int hostId, int connectionId, int channelId, byte[] buffer, int size, out byte error)
    {
        return NetworkTransport.Send(hostId, connectionId, channelId, buffer, size, out error);
    }

    public void SetBroadcastCredentials(int hostId, int key, int version, int subversion, out byte error)
    {
        NetworkTransport.SetBroadcastCredentials(hostId, key, version, subversion, out error);
    }

    public void SetPacketStat(int direction, int packetStatId, int numMsgs, int numBytes)
    {
        NetworkTransport.SetPacketStat(direction, packetStatId, numMsgs, numBytes);
    }

    public void Shutdown()
    {
        NetworkTransport.Shutdown();
        NextServerFlush();
        NextServerDestroy();
        Next.NextTerm();
    }

    public bool StartBroadcastDiscovery(int hostId, int broadcastPort, int key, int version, int subversion, byte[] buffer, int size, int timeout, out byte error)
    {
        return NetworkTransport.StartBroadcastDiscovery(hostId, broadcastPort, key, version, subversion, buffer, size, timeout, out error);
    }

    public void StopBroadcastDiscovery()
    {
        NetworkTransport.StopBroadcastDiscovery();
    }

    #endregion // #region INetworkTransport Functions
}
