package storage

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
)

func NewRedisPool(hostname string, maxIdleConnections int, maxActiveConnections int) *redis.Pool {
	pool := redis.Pool{
		MaxIdle:     maxIdleConnections,
		MaxActive:   maxActiveConnections,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", hostname)
		},
	}

	return &pool
}

func ValidateRedisPool(pool *redis.Pool) error {
	redisConn := pool.Get()
	defer redisConn.Close()

	redisConn.Send("PING")
	redisConn.Flush()
	pong, err := redisConn.Receive()
	if err != nil || pong != "PONG" {
		return fmt.Errorf("could not ping: %v", err)
	}

	return nil
}

type RawRedisClient struct {
	conn net.Conn
}

func NewRawRedisClient(hostname string) (*RawRedisClient, error) {
	conn, err := net.Dial("tcp", hostname)
	if err != nil {
		return nil, fmt.Errorf("could not dial: %v", err)
	}

	client := RawRedisClient{
		conn: conn,
	}

	return &client, nil
}

func (r *RawRedisClient) Ping() error {
	if err := r.Command("PING", ""); err != nil {
		return err
	}

	redisReplyReader := bufio.NewReader(r.conn)
	_, err := redisReplyReader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("could not ping: %v", err)
	}

	return nil
}

func (r *RawRedisClient) Command(command string, format string, args ...interface{}) error {
	if len(args) != 0 {
		commandString := fmt.Sprintf(command+" "+format+"\r\n", args...)
		if _, err := fmt.Fprint(r.conn, commandString); err != nil {
			return fmt.Errorf("failed to write redis command '%s': %v", commandString, err)
		}
	} else {
		commandString := fmt.Sprintf(command+"\r\n", args...)
		if _, err := fmt.Fprint(r.conn, commandString); err != nil {
			return fmt.Errorf("failed to write redis command '%s': %v", commandString, err)
		}
	}

	return nil
}

func (r *RawRedisClient) StartCommand(command string) {
	fmt.Fprintf(r.conn, command+" ")
}

func (r *RawRedisClient) CommandArgs(format string, args ...interface{}) {
	fmt.Fprintf(r.conn, format, args...)
}

func (r *RawRedisClient) EndCommand() {
	fmt.Fprintf(r.conn, "\r\n")
}

func (r *RawRedisClient) Close() error {
	return r.conn.Close()
}
