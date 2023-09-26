package common

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

type UDPServerConfig struct {
	Port              int
	NumThreads        int
	SocketReadBuffer  int
	SocketWriteBuffer int
	MaxPacketSize     int
	BindAddress       net.UDPAddr
}

type UDPServer struct {
	config UDPServerConfig
	conn   []*net.UDPConn
}

func CreateUDPServer(ctx context.Context, config UDPServerConfig, packetHandler func(conn *net.UDPConn, from *net.UDPAddr, packet []byte)) *UDPServer {

	udpServer := UDPServer{}
	udpServer.config = config
	udpServer.conn = make([]*net.UDPConn, config.NumThreads)

	lc := net.ListenConfig{
		Control: func(network string, address string, c syscall.RawConn) error {
			err := c.Control(func(fileDescriptor uintptr) {
				err := unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse address socket option: %v", err))
				}

				err = unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse port socket option: %v", err))
				}
			})

			return err
		},
	}

	for i := 0; i < config.NumThreads; i++ {

		bindAddress := config.BindAddress.String()

		lp, err := lc.ListenPacket(ctx, "udp", bindAddress)
		if err != nil {
			panic(fmt.Sprintf("could not bind socket: %v", err))
		}

		udpServer.conn[i] = lp.(*net.UDPConn)

		if err := udpServer.conn[i].SetReadBuffer(config.SocketReadBuffer); err != nil {
			panic(fmt.Sprintf("could not set socket read buffer size: %v", err))
		}

		if err := udpServer.conn[i].SetWriteBuffer(config.SocketWriteBuffer); err != nil {
			panic(fmt.Sprintf("could not set socket write buffer size: %v", err))
		}

		go func(thread int) {

			for {

				// read packet

				receiveBuffer := make([]byte, config.MaxPacketSize)

				receivePacketBytes, from, err := udpServer.conn[thread].ReadFromUDP(receiveBuffer[:])
				if err != nil {
					fmt.Printf("udp receive error: %v\n", err)
					break
				}

				go packetHandler(udpServer.conn[thread], from, receiveBuffer[:receivePacketBytes])
			}

			udpServer.conn[thread].Close()

		}(i)
	}

	return &udpServer
}
