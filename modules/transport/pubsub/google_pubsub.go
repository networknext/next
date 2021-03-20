package pubsub

import (
	"context"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
)

// Entry defines the methods for Google PubSub structs to implement
type Entry interface {
	validate() bool                  // private function within write entry
	checkNaNOrInf() (bool, []string) // private function within write entry
	Save() (map[string]bigquery.Value, string, error)
	WriteEntry() ([]byte, error)
	ReadEntry(data []byte) bool
}

type GooglePubSubPublisher interface {
	Publish(ctx context.Context, entry *Entry) error
}

type GooglePubSubSubscriber interface {
	Receive(ctx context.Context) error
	Submit(ctx context.Context) error
    WriteLoop(ctx context.Context) error
    unbatchMessages(m *pubsub.Message) ([][]byte, error) // private function within receive
}
