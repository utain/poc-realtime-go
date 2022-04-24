package pubsub

import "context"

type Pubsub interface {
	Close() error
	Subscriber(ctx context.Context, topic string) error
	WriteMessage(ctx context.Context, topic string, chanId string, payload []byte) error
}
