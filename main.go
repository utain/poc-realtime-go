package main

import (
	"context"
	"flag"
	"log"

	"github.com/utain/poc/go-realtime/config"
	"github.com/utain/poc/go-realtime/pubsub"
	"github.com/utain/poc/go-realtime/server"
	"github.com/utain/poc/go-realtime/withgo"
	"github.com/utain/poc/go-realtime/withkafka"
	"github.com/utain/poc/go-realtime/withpulsar"
	"github.com/utain/poc/go-realtime/withredis"
	"github.com/utain/poc/go-realtime/ws"
)

func init() {
	config.Parse()
}

func main() {
	var port uint
	var internalPort uint
	var bus uint
	flag.UintVar(&port, "p", 6000, "Server port default 6000")
	flag.UintVar(&internalPort, "b", 16000, "Server port default 16000")
	flag.UintVar(&bus, "t", 0, "Pub/sub types 0 (redis), 1 (kafka), 2 (pulsar), 3 (internal)")
	flag.Parse()

	manager := ws.WsManager{
		Clients:   make(map[*ws.Client]bool),
		Regiser:   make(chan *ws.Client),
		Boardcast: make(chan ws.Message),
	}
	go manager.Listen()
	conf := config.Get()
	// recieve from redis
	var ibus pubsub.Pubsub
	switch bus {
	case 0:
		log.Println("Using redis")
		ibus = withredis.Open(pubsub.DefaultOptions{
			WsManager: &manager,
			Addrs:     conf.RedisAddrs,
		})
	case 1:
		log.Println("Using kafka")
		ibus = withkafka.Open(pubsub.DefaultOptions{
			WsManager: &manager,
			Addrs:     conf.KafkaAddrs,
		})
	case 2:
		log.Println("Using pulsar")
		ibus = withpulsar.Open(pubsub.DefaultOptions{
			WsManager: &manager,
			Addrs:     conf.PulsarAddrs,
		})
	case 3:
		log.Println("Using puregolang")
		ibus = withgo.Open(withgo.Options{
			DefaultOptions: pubsub.DefaultOptions{
				WsManager: &manager,
				Addrs:     conf.InternalAddrs,
			},
			InternalPort: internalPort,
		})
	}

	defer ibus.Close()

	go ibus.Subscriber(context.Background(), conf.Topic)

	server.Start(server.ServerOpts{
		Port: port,
		WriteMessage: func(ctx context.Context, chanId string, message []byte) error {
			return ibus.WriteMessage(ctx, conf.Topic, chanId, message)
		},
		Manager: &manager,
	})
}
