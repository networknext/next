package pubsub

import (
"context"
"sync"
"syscall"

"github.com/pebbe/zmq4"
)

type GenericPublisher struct {
	socket *zmq4.Socket
	mutex  sync.Mutex
}

func NewMultiPublisher(hosts []string, sendBufferSize int) ([]Publisher,error) {
	var publishers []Publisher
	for _, host := range hosts {
		var publisher Publisher
		gPub, err := NewGenericPublisher(host,sendBufferSize)
		if err != nil {
			return nil, err
		}

		publisher = gPub
		publishers = append(publishers, publisher)
	}
	return publishers, nil
}

func NewGenericPublisher(host string, sendBufferSize int) (*GenericPublisher, error) {
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
