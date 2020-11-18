package pubsub

import (
	"context"
	"errors"
	"fmt"
	"github.com/pebbe/zmq4"
	"sync"
)

type GenericSubscriber struct {
	socket *zmq4.Socket
	mutex  sync.Mutex
	topics map[Topic]bool
}

func NewGenericSubscriber(port string, receiveBufferSize int) (*GenericSubscriber, error) {
	socket, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		return nil, err
	}

	if err := socket.SetRcvhwm(receiveBufferSize); err != nil {
		return nil, err
	}

	if err = socket.Bind("tcp://*:" + port); err != nil {
		return nil, err
	}

	return &GenericSubscriber{
		socket: socket,
		topics: make(map[Topic]bool),
	}, nil
}

func (sub *GenericSubscriber) Subscribe(topic Topic) error {
	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	sub.topics[topic] = true
	return sub.socket.SetSubscribe(string(topic))
}

func (sub *GenericSubscriber) Unsubscribe(topic Topic) error {
	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	if _, ok := sub.topics[topic]; !ok {
		return fmt.Errorf("failed to unsubscribe from topic %s: not subscribed to topic", topic.String())
	}

	delete(sub.topics, topic)
	return sub.socket.SetUnsubscribe(string(topic))
}

func (sub *GenericSubscriber) ReceiveMessage(ctx context.Context) <-chan MessageInfo {
	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	receiveChan := make(chan MessageInfo)
	receiveFunc := func(topic Topic, message []byte, err error) {
		receiveChan <- MessageInfo{
			Topic:   topic,
			Message: message,
			Err:     err,
		}
	}

	message, err := sub.socket.RecvMessageBytes(0)
	if err != nil {
		go receiveFunc(0, nil, err)
		return receiveChan
	}

	if len(message) <= 1 {
		go receiveFunc(0, nil, errors.New("message size is 0"))
		return receiveChan
	}

	if len(message[0]) == 0 {
		go receiveFunc(0, nil, errors.New("topic size is 0"))
		return receiveChan
	}

	topic := Topic(message[0][0])

	if _, ok := sub.topics[topic]; !ok {
		go receiveFunc(0, nil, errors.New("subscriber received message from wrong topic"))
		return receiveChan
	}

	go receiveFunc(topic, message[1], nil)
	return receiveChan

}

func (sub *GenericSubscriber) Close() {
	sub.socket.Close()
}
