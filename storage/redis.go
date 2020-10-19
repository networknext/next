package storage

import (
	"bufio"
	"fmt"
	"net"
	"strings"
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
	cmdArgsString := fmt.Sprintf(format, args...)
	var cmdArgs []string
	if len(cmdArgsString) > 0 {
		cmdArgs = strings.Split(cmdArgsString, " ")
	}
	argCount := fmt.Sprintf("%d", 1+len(cmdArgs))

	// Convert the command and arguments to follow the redis RESP specification:
	// https://redis.io/topics/protocol
	commandLength := fmt.Sprintf("%d", len(command))
	commandString := "*" + argCount + "\r\n$" + commandLength + "\r\n" + command + "\r\n"
	for i := range cmdArgs {
		commandString += fmt.Sprintf("$%d\r\n%s\r\n", len(cmdArgs[i]), cmdArgs[i])
	}

	if _, err := fmt.Fprint(r.conn, commandString); err != nil {
		return fmt.Errorf("failed to write redis command '%s': %v", commandString, err)
	}

	return nil
}

func (r *RawRedisClient) Close() error {
	return r.conn.Close()
}
