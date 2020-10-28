package pubsub

import (
	"context"
	"fmt"
)

type Topic byte

func (topic Topic) String() string {
	return fmt.Sprintf("%d", topic)
}

type Publisher interface {
	Publish(ctx context.Context, topic Topic, message []byte) (int, error)
}

type Subscriber interface {
	Subscribe(topic Topic) error
	Unsubscribe(topic Topic) error
	ReceiveMessage(ctx context.Context) (Topic, <-chan []byte, error)
}

type ErrRetry struct{}

func (e *ErrRetry) Error() string {
	return fmt.Sprintf("retry")
}
