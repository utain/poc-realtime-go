package withpulsar

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/google/uuid"
	"github.com/utain/poc/go-realtime/pubsub"
	"github.com/utain/poc/go-realtime/ws"
)

type WithPulsar struct {
	client pulsar.Client
	wsm    *ws.WsManager
}

func Open(opts pubsub.DefaultOptions) pubsub.Pubsub {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               strings.Join(opts.Addrs, ","),
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Could not instantiate Pulsar client: %v", err)
	}
	return &WithPulsar{client, opts.WsManager}
}

func (w *WithPulsar) Close() (err error) {
	w.client.Close()
	return
}

func (w *WithPulsar) Subscriber(ctx context.Context, topic string) error {
	consumer, err := w.client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       topic,
		SubscriptionName:            uuid.NewString(),
		Type:                        pulsar.Shared,
		SubscriptionInitialPosition: pulsar.SubscriptionPositionLatest,
	})
	if err != nil {
		return err
	}
	defer func() {
		fmt.Println("Close consumer")
		consumer.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context:Done")
			return nil
		case msg := <-consumer.Chan():
			fmt.Println("Consume:", msg)
			json, err := ws.ParseMessage(msg.Payload())
			if err != nil {
				continue
			}
			defer consumer.AckID(msg.ID())
			w.wsm.Boardcast <- json
			// case <-time.After(time.Minute * 30):
			// 	fmt.Println("Timer")
		}
	}
}

func (w *WithPulsar) WriteMessage(ctx context.Context, topic string, key string, payload []byte) error {
	producer, err := w.client.CreateProducer(pulsar.ProducerOptions{Topic: topic})
	if err != nil {
		return err
	}
	defer producer.Close()

	_, err = producer.Send(ctx, &pulsar.ProducerMessage{
		Key:     key,
		Payload: payload,
	})
	return err
}
