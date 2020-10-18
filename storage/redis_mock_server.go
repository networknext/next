package storage

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/alicebob/miniredis"
)

type ErrMockRedisListen struct {
	address string
	err     error
}

func (e *ErrMockRedisListen) Error() string {
	return fmt.Sprintf("could not listen on %s: %v", e.address, e.err)
}

type ErrMockRedisAcceptConn struct {
	err error
}

func (e *ErrMockRedisAcceptConn) Error() string {
	return fmt.Sprintf("failed to accept connection: %v", e.err)
}

type ErrMockRedisRequest struct {
	request string
	err     error
}

func (e *ErrMockRedisRequest) Error() string {
	return fmt.Sprintf("bad request %s: %v", e.request, e.err)
}

type ErrMockRedisReply struct {
	reply string
	err   error
}

func (e *ErrMockRedisReply) Error() string {
	return fmt.Sprintf("failed to reply with %s: %v", e.reply, e.err)
}

type ErrMockRedisCommand struct {
	commandName string
	args        []string
	err         error
}

func (e *ErrMockRedisCommand) Error() string {
	return fmt.Sprintf("failed to execute command %s with args %s: %v", e.commandName, e.args, e.err)
}

// MockRedisServer is a mock redis server intended to be used to test the RawRedisClient
// It opens a tcp connection with the RawRedisClient and only responds to ping commands.
type MockRedisServer struct {
	ctx      context.Context
	listener net.Listener
	Storage  *miniredis.Miniredis
	mutex    sync.Mutex
}

func NewMockRedisServer(ctx context.Context, address string) (*MockRedisServer, error) {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return nil, &ErrMockRedisListen{address: address, err: err}
	}

	return &MockRedisServer{
		ctx:      ctx,
		listener: listener,
		Storage:  miniredis.NewMiniRedis(),
	}, nil
}

func (mock *MockRedisServer) Start(errCallback func(err error)) {
	go func() {
		for {
			select {
			case <-mock.ctx.Done():
				return
			default:
				conn, err := mock.listener.Accept()
				if err != nil {
					if errCallback != nil {
						errCallback(&ErrMockRedisAcceptConn{err: err})
					}
					continue
				}

				for {
					redisReplyReader := bufio.NewReader(conn)
					request, err := redisReplyReader.ReadString('\n')
					if err != nil {
						break
					}

					request = strings.Trim(request, "\r\n")
					split := strings.Split(request, " ")

					command := split[0]
					args := []string{}
					if len(split) > 1 {
						args = split[1:]
					}

					switch command {
					case "PING":
						reply := "PONG\r\n"
						if _, err := fmt.Fprint(conn, reply); err != nil {
							if errCallback != nil {
								errCallback(&ErrMockRedisReply{reply: reply, err: err})
							}
						}
						break

					case "SET":
						if len(args) < 2 {
							if errCallback != nil {
								errCallback(&ErrMockRedisRequest{request: request, err: err})
							}
							continue
						}

						key := args[0]
						value := args[1]

						fmt.Println("inserting into storage")

						mock.mutex.Lock()
						if err := mock.Storage.Set(key, value); err != nil {
							if errCallback != nil {
								errCallback(&ErrMockRedisCommand{commandName: "SET", args: args, err: err})
							}
						}
						mock.mutex.Unlock()

						fmt.Println("inserted into storage")
						break

					default:
						break

					}
				}

				conn.Close()
			}
		}
	}()
}

func (mock *MockRedisServer) Addr() net.Addr {
	return mock.listener.Addr()
}

func (mock *MockRedisServer) Dump() string {
	mock.mutex.Lock()
	defer mock.mutex.Unlock()

	return mock.Storage.Dump()
}
