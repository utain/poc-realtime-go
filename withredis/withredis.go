package withredis

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/utain/poc/go-realtime/pubsub"
	"github.com/utain/poc/go-realtime/ws"
)

type WithRedis struct {
	db  redis.UniversalClient
	wsm *ws.WsManager
}

func Open(opt pubsub.DefaultOptions) pubsub.Pubsub {
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: opt.Addrs,
	})
	return &WithRedis{rdb, opt.WsManager}
}

func (w *WithRedis) Close() error {
	return w.db.Close()
}

func (w *WithRedis) Subscriber(ctx context.Context, topic string) error {
	s := w.db.Subscribe(context.Background(), topic)
	for {
		// recieved data from redis pub/sub
		msg, err := s.ReceiveMessage(context.Background())
		if err != nil {
			log.Println("BUG:", err)
			continue
		}
		cmsg, err := ws.ParseMessage([]byte(msg.Payload))
		if err != nil {
			log.Println("BUG:", err)
			continue
		}
		// send to ws manager
		fmt.Println("Redis send to wsmanager:", msg.Payload)
		w.wsm.Boardcast <- cmsg
	}
}

func (w *WithRedis) WriteMessage(ctx context.Context, topic string, key string, payload []byte) error {
	if e := w.db.Publish(ctx, topic, payload); e.Err() != nil {
		return e.Err()
	}
	return nil
}
