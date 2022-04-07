using System;
using System.Net;
using UnityEngine.Networking;
using UnityEngine.Networking.Types;
using NetworkNext;

public class NextClientTransport : INetworkTransport
{
    public IntPtr client;
    public IntPtr clientCtxPtr;
    public Next.NextConfig nextConfig;
    public string serverAddress;
    public string clientBindAddress;
    public NextClientPacketReceivedCallback clientPacketReceivedCallback;
    public NextWakeupCallback clientWakeupCallback;

    public NextClientTransport(IntPtr ctxPtr, ref Next.NextConfig config, string bindIP, int bindPort, string serverIP, int serverPort, NextClientPacketReceivedCallback packetRecvCallback, NextWakeupCallback wakeupCallback = null)
    {
        clientCtxPtr = ctxPtr;
        nextConfig = config;
        serverAddress = String.Format("{0}:{1}", serverIP, serverPort);
        clientBindAddress = String.Format("{0}:{1}", bindIP, bindPort);
        clientPacketReceivedCallback = packetRecvCallback;
        clientWakeupCallback = wakeupCallback;

        if (!IsClientReady())
        {
            if (Next.NextInit(clientCtxPtr, ref nextConfig) != Next.NEXT_OK)
            {
                throw new InvalidOperationException("could not initialize Network Next");
            }

            client = Next.NextClientCreate(clientCtxPtr, clientBindAddress, clientPacketReceivedCallback, clientWakeupCallback);
            if (!IsClientReady())
            {
                throw new InvalidOperationException("could not create client");
            }
        }
    }

    #region Utility Functions

    public bool IsClientReady()
    {
        return client != null && client != IntPtr.Zero;
    }

    public bool IsClientConnected()
    {
        return IsClientReady() && Next.NextClientState(client) == Next.NEXT_CLIENT_STATE_OPEN;
    }

    #endregion // #region Utility Functions

    #region Next Client Functions

    public void NextClientDestroy()
    {
        if (IsClientReady())
        {
            Next.NextClientDestroy(client);
            client = IntPtr.Zero;
        }
    }

    public ushort NextClientPort()
    {
        if (IsClientReady())
        {
            return Next.NextClientPort(client);
        }
        return 0;
    }

    public void NextClientOpenSession()
    {
        if (IsClientReady())
        {
            Next.NextClientOpenSession(client, serverAddress);
        }
    }

    public void NextClientCloseSession()
    {
        if (IsClientConnected())
        {
            Next.NextClientCloseSession(client);
        }
    }

    public bool NextClientIsSessionOpen()
    {
        return IsClientReady() && Next.NextClientIsSessionOpen(client);
    }

    public int NextClientState()
    {
        if (IsClientReady())
        {
            return Next.NextClientState(client);
        }
        return Next.NEXT_ERROR;
    }

    public void NextClientUpdate()
    {
        if (IsClientConnected())
        {
            Next.NextClientUpdate(client);
        }
    }

    public void NextClientSendPacket(byte[] buffer, int size)
    {
        if (IsClientConnected())
        {
            Next.NextClientSendPacket(client, buffer, size);
        }
    }

    public void NextClientSendPacketDirect(byte[] buffer, int size)
    {
        if (IsClientConnected())
        {
            Next.NextClientSendPacketDirect(client, buffer, size);
        }
    }

    public void NextClientReportSession()
    {
        if (IsClientConnected())
        {
            Next.NextClientReportSession(client);
        }
    }

    public ulong NextClientSessionID()
    {
        if (IsClientConnected())
        {
            return Next.NextClientSessionID(client);
        }
        return 0;
    }

    public Next.ClientStats NextClientStats()
    {
        if (IsClientConnected())
        {
            return Next.NextClientStats(client);
        }
        return new Next.ClientStats();
    }

    public Next.NextAddress NextClientServerAddress()
    {
        if (IsClientConnected())
        {
            return Next.NextClientServerAddress(client);
        }
        return new Next.NextAddress();
    }

    #endregion // #region Next Client Functions

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
        NextClientCloseSession();
        return NetworkTransport.Disconnect(hostId, connectionId, out error);
    }

    public bool DoesEndPointUsePlatformProtocols(EndPoint endPoint)
    {
        throw new NotImplementedException("NextClientTransport does not support DoesEndPointUsePlatformProtocols");
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
        if (IsClientConnected())
        {
            error = (byte)NetworkError.Ok;
            Next.ClientStats clientStats = NextClientStats();
            if (clientStats.Next)
            {
                return (int)clientStats.NextRTT;
            }
            return (int)clientStats.DirectMinRTT;
        }
        return NetworkTransport.GetCurrentRTT(hostId, connectionId, out error);
    }

    public void Init()
    {
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
        NextClientCloseSession();
        NextClientDestroy();
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
