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
	producer sarama.AsyncProducer
	consumer sarama.ConsumerGroup
	wsm      *ws.WsManager
}

func Open(wsm *ws.WsManager) pubsub.Pubsub {
	// reader := kafka.NewReader(kafka.ReaderConfig{
	// 	Brokers: []string{"localhost:9092"},
	// 	Topic:   topic,
	// })
	// writer := kafka.Writer{
	// 	Addr: ,
	// 	Brokers: []string{"localhost:9092"},
	// }
	config := sarama.NewConfig()
	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalln("NewAsyncProducer:", err)
	}
	consumer, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, uuid.NewString(), config)
	if err != nil {
		log.Fatalln("NewConsumer:", err)
	}
	return &WithKafka{
		producer: producer,
		consumer: consumer,
		wsm:      wsm,
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
		fmt.Println("Hello")
	}
}
func (w *WithKafka) WriteMessage(ctx context.Context, topic string, chanId string, payload []byte) error {
	w.producer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(chanId),
		Value: sarama.StringEncoder(payload),
	}
	return nil
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
