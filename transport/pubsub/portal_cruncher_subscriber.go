package pubsub

import (
	"errors"

	"github.com/pebbe/zmq4"
)

type PortalCruncherSubscriber struct {
	socket *zmq4.Socket
	topic  Topic
}

func NewPortalCruncherSubscriber(port string) (*PortalCruncherSubscriber, error) {
	socket, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		return nil, err
	}

	if err = socket.Bind("tcp://*:" + port); err != nil {
		return nil, err
	}

	return &PortalCruncherSubscriber{
		socket: socket,
	}, nil
}

func (sub *PortalCruncherSubscriber) Subscribe(topic Topic) error {
	sub.topic = topic
	return sub.socket.SetSubscribe(string(topic))
}

func (sub *PortalCruncherSubscriber) ReceiveMessage() ([]byte, error) {
	message, err := sub.socket.RecvMessageBytes(0)
	if err != nil {
		return nil, err
	}

	if len(message) <= 1 {
		return nil, errors.New("message size is 0")
	}

	if len(message[0]) == 0 {
		return nil, errors.New("topic size is 0")
	}

	topic := Topic(message[0][0])

	if topic.String() != sub.topic.String() {
		return nil, errors.New("subscriber received message from wrong topic")
	}

	return message[1], nil
}
