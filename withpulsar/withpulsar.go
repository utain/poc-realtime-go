package withpulsar

import (
	"context"
	"fmt"
	"log"
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

func Open(wsm *ws.WsManager) pubsub.Pubsub {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               "pulsar://localhost:6650",
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Could not instantiate Pulsar client: %v", err)
	}
	return &WithPulsar{client, wsm}
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
			fmt.Println("SubscriberDone")
			return nil
		case msg := <-consumer.Chan():
			fmt.Println("SubscriberDone:", msg)
			json, err := ws.ParseMessage(msg.Payload())
			if err != nil {
				continue
			}
			defer consumer.AckID(msg.ID())
			w.wsm.Boardcast <- json
		case <-time.After(time.Minute * 30):
			fmt.Println("SubscriberDone:Loop")
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
