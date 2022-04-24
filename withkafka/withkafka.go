package withkafka

import (
	"context"
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/utain/poc/go-realtime/pubsub"
	"github.com/utain/poc/go-realtime/ws"
)

type WithKafka struct {
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
	wsm      *ws.WsManager
}

func Open(opts pubsub.DefaultOptions) pubsub.Pubsub {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Return.Successes = true
	fmt.Println("opts.Addrs:", opts.Addrs)
	producer, err := sarama.NewSyncProducer(opts.Addrs, config)
	if err != nil {
		log.Fatalln("SyncProducer:", err)
	}
	consumer, err := sarama.NewConsumerGroup(opts.Addrs, uuid.NewString(), config)
	if err != nil {
		log.Fatalln("NewConsumer:", err)
	}
	return &WithKafka{
		producer: producer,
		consumer: consumer,
		wsm:      opts.WsManager,
	}
}

func (w *WithKafka) Close() (err error) {
	w.producer.Close()
	w.consumer.Close()
	return
}

func (w *WithKafka) Subscriber(ctx context.Context, topic string) error {
	handler := NewHandler(w.wsm)
	for {
		w.consumer.Consume(ctx, []string{topic}, handler)
	}
}
func (w *WithKafka) WriteMessage(ctx context.Context, topic string, chanId string, payload []byte) error {
	log.Println("Send to broker:", topic)
	partition, offset, err := w.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(chanId),
		Value: sarama.StringEncoder(payload),
	})
	if err != nil {
		log.Fatalln("Can't send message to the broker", err)
	}
	log.Printf("Sent at partition: %d and offset: %d\r\n", partition, offset)
	return err
}

// consumer handler
func NewHandler(wsm *ws.WsManager) sarama.ConsumerGroupHandler {
	return &wsHandler{wsm}
}

type wsHandler struct {
	wsm *ws.WsManager
}

func (wsHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (wsHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h wsHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		sess.MarkMessage(msg, "")
		json, err := ws.ParseMessage(msg.Value)
		if err != nil {
			return err
		}
		h.wsm.Boardcast <- json
	}
	return nil
}
