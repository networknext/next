package storage

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
)

func NewRedisPool(hostname string) *redis.Pool {
	pool := redis.Pool{
		MaxIdle:     5,
		MaxActive:   64,
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
	fmt.Fprint(r.conn, "PING\r\n")

	redisReplyReader := bufio.NewReader(r.conn)
	reply, err := redisReplyReader.ReadString('\n')
	if err != nil || reply != "+PONG\r\n" {
		r.conn.Close()
		return fmt.Errorf("could not ping: %v", err)
	}

	return nil
}

func (r *RawRedisClient) Command(command string, format string, args ...interface{}) {
	if len(args) != 0 {
		fmt.Fprintf(r.conn, command+" "+format+"\r\n", args...)
	} else {
		fmt.Fprint(r.conn, command+"\r\n")
	}
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
