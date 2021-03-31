package pubsub

import (
	"context"
	"sync"
	"syscall"

	"github.com/pebbe/zmq4"
)

const (
	TopicVanityMetricData Topic = 1
)

type VanityMetricPublisher struct {
	socket *zmq4.Socket
	mutex  sync.Mutex
}

func NewVanityMetricPublisher(host string, sendBufferSize int) (*VanityMetricPublisher, error) {
	socket, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		return nil, err
	}

	if err := socket.SetXpubNodrop(true); err != nil {
		return nil, err
	}

	if err := socket.SetSndhwm(sendBufferSize); err != nil {
		return nil, err
	}

	if err = socket.Connect(host); err != nil {
		return nil, err
	}

	return &VanityMetricPublisher{
		socket: socket,
	}, nil
}

func (pub *VanityMetricPublisher) Publish(ctx context.Context, topic Topic, message []byte) (int, error) {
	pub.mutex.Lock()
	defer pub.mutex.Unlock()

	bytes, err := pub.socket.SendMessageDontwait([]byte{byte(topic)}, message)
	errno := zmq4.AsErrno(err)
	switch errno {
	case zmq4.AsErrno(syscall.EAGAIN):
		err = &ErrRetry{}
	default:
	}

	return bytes, err
}

func (pub *VanityMetricPublisher) Close() error {
	pub.mutex.Lock()
	defer pub.mutex.Unlock()
	return pub.socket.Close()
}
