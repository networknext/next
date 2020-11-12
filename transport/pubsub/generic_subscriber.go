package pubsub

import (
	"context"
	"errors"
	"fmt"
	"github.com/pebbe/zmq4"
	"sync"
	"syscall"
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

	if _, ok := sub.topics[topic]; !ok{
		return fmt.Errorf("failed to unsubscribe from topic %s: not subscribed to topic", topic.String())
	}

	delete(sub.topics,topic)
	return sub.socket.SetUnsubscribe(string(topic))
}

func (sub *GenericSubscriber) ReceiveMessage(ctx context.Context) (Topic, <-chan []byte, error) {
	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	for {
		select {
		case <-ctx.Done():
			return 0, nil, ctx.Err()
		default:
			message, err := sub.socket.RecvMessageBytes(zmq4.DONTWAIT)
			if err != nil {
				if zmq4.AsErrno(err) == zmq4.AsErrno(syscall.EAGAIN) {
					continue
				}

				return 0, nil, err
			}

			if len(message) <= 1 {
				return 0, nil, errors.New("message size is 0")
			}

			if len(message[0]) == 0 {
				return 0, nil, errors.New("topic size is 0")
			}

			topic := Topic(message[0][0])

			if _, ok := sub.topics[topic]; !ok {
				return 0, nil, errors.New("subscriber received message from wrong topic")
			}

			out := make(chan []byte)
			out <- message[1]
			return topic, out, nil
		}
	}
}

func (sub *GenericSubscriber)Close(){
	sub.socket.Close()
}