package pubsub

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/pebbe/zmq4"
)

type GenericPublisher struct {
	socket *zmq4.Socket
	mutex  sync.Mutex
}

func NewMultiPublisher(hosts []string, sendBufferSize int) ([]Publisher, error) {
	var publishers []Publisher
	for _, host := range hosts {
		publisher, err := NewGenericPublisher(host, sendBufferSize)
		if err != nil {
			return nil, err
		}

		publishers = append(publishers, publisher)
	}
	return publishers, nil
}

func NewGenericPublisher(host string, sendBufferSize int) (*GenericPublisher, error) {
	socket, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		return nil, fmt.Errorf("create socket error %v", err)
	}

	if err := socket.SetXpubNodrop(true); err != nil {
		return nil, fmt.Errorf("Xpubnodrop error %v", err)
	}

	if err := socket.SetSndhwm(sendBufferSize); err != nil {
		return nil, fmt.Errorf("sndhwm error %v", err)
	}

	if err = socket.Connect(host); err != nil {
		return nil, fmt.Errorf("connection error %v", err)
	}

	return &GenericPublisher{
		socket: socket,
	}, nil
}

func (p *GenericPublisher) Publish(ctx context.Context, topic Topic, message []byte) (int, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	bytes, err := p.socket.SendMessageDontwait([]byte{byte(topic)}, message)
	errno := zmq4.AsErrno(err)
	switch errno {
	case zmq4.AsErrno(syscall.EAGAIN):
		err = &ErrRetry{}
	}

	return bytes, err
}

func (p *GenericPublisher) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.socket.Close()
}
