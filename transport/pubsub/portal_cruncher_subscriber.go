package pubsub

import (
	"errors"
	"fmt"
	"sync"

	"github.com/pebbe/zmq4"
)

type PortalCruncherSubscriber struct {
	socket *zmq4.Socket
	mutex  sync.Mutex

	topics []Topic
}

func NewPortalCruncherSubscriber(port string) (*PortalCruncherSubscriber, error) {
	socket, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		return nil, err
	}

	if err := socket.SetRcvhwm(1); err != nil {
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
	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	sub.topics = append(sub.topics, topic)
	return sub.socket.SetSubscribe(string(topic))
}

func (sub *PortalCruncherSubscriber) Unsubscribe(topic Topic) error {
	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	containsTopic, topicIndex := sub.containsTopic(topic)
	if !containsTopic {
		return fmt.Errorf("failed to unsubscribe from topic %s: not subscribed to topic", topic.String())
	}

	sub.topics = append(sub.topics[:topicIndex], sub.topics[topicIndex+1:]...)
	return sub.socket.SetUnsubscribe(string(topic))
}

func (sub *PortalCruncherSubscriber) ReceiveMessage() (Topic, []byte, error) {
	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	message, err := sub.socket.RecvMessageBytes(0)
	if err != nil {
		return 0, nil, err
	}

	if len(message) <= 1 {
		return 0, nil, errors.New("message size is 0")
	}

	if len(message[0]) == 0 {
		return 0, nil, errors.New("topic size is 0")
	}

	topic := Topic(message[0][0])

	if containsTopic, topicIndex := sub.containsTopic(topic); containsTopic {
		if topic.String() != sub.topics[topicIndex].String() {
			return 0, nil, errors.New("subscriber received message from wrong topic")
		}
	}

	return topic, message[1], nil
}

func (sub *PortalCruncherSubscriber) containsTopic(topic Topic) (bool, int) {
	var containsTopic bool
	var topicIndex int
	for i, t := range sub.topics {
		if t == topic {
			containsTopic = true
			topicIndex = i
			break
		}
	}

	return containsTopic, topicIndex
}
