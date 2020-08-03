package pubsub

import (
	"fmt"
)

type Topic byte

func (topic Topic) String() string {
	return fmt.Sprintf("%d", topic)
}

type Publisher interface {
	Publish(topic Topic, message []byte) (int, error)
}

type Subscriber interface {
	Subscribe(topic Topic) error
	Unsubscribe(topic Topic) error
	ReceiveMessage() (Topic, []byte, error)
}
