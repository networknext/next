package pubsub

import "context"

type NoOpPublisher struct{}

func (noop *NoOpPublisher) Publish(ctx context.Context, topic Topic, message []byte) (int, error) {
	return 0, nil
}

func (noop *NoOpPublisher) Close() error {
	return nil
}

type NoOpSubscriber struct{}

func (noop *NoOpPublisher) Subscribe(topic Topic) error {
	return nil
}

func (noop *NoOpPublisher) Unsubscribe(topic Topic) error {
	return nil
}

func (noop *NoOpPublisher) ReceiveMessage() (Topic, []byte, error) {
	return 0, nil, nil
}
