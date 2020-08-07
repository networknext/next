package pubsub

import (
	"sync"

	"github.com/pebbe/zmq4"
)

const (
	TopicPortalCruncherSessionData   Topic = 1
	TopicPortalCruncherSessionCounts Topic = 2
)

type PortalCruncherPublisher struct {
	socket *zmq4.Socket
	mutex  sync.Mutex
}

func NewPortalCruncherPublisher(host string, sendBufferSize int) (*PortalCruncherPublisher, error) {
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

	return &PortalCruncherPublisher{
		socket: socket,
	}, nil
}

func (pub *PortalCruncherPublisher) Publish(topic Topic, message []byte) (int, error) {
	pub.mutex.Lock()
	defer pub.mutex.Unlock()
	return pub.socket.SendMessageDontwait([]byte{byte(topic)}, message)
}
