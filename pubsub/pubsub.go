package pubsub

import (
	"context"

	"github.com/utain/poc/go-realtime/ws"
)

type Pubsub interface {
	Close() error
	Subscriber(ctx context.Context, topic string) error
	WriteMessage(ctx context.Context, topic string, chanId string, payload []byte) error
}

type DefaultOptions struct {
	WsManager *ws.WsManager
	Addrs     []string
}
