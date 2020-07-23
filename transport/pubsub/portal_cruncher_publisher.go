package pubsub

/*
import (
	"github.com/pebbe/zmq4"
)

const (
	TopicPortalCruncherSessionData Topic = 1
)

type PortalCruncherPublisher struct {
	socket *zmq4.Socket
}

func NewPortalCruncherPublisher(host string) (*PortalCruncherPublisher, error) {
	socket, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
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
	return pub.socket.SendMessageDontwait([]byte{byte(topic)}, message)
}
*/