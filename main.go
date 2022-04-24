package main

import (
	"context"
	"flag"

	"github.com/utain/poc/go-realtime/pubsub"
	"github.com/utain/poc/go-realtime/server"
	"github.com/utain/poc/go-realtime/withkafka"
	"github.com/utain/poc/go-realtime/withpulsar"
	"github.com/utain/poc/go-realtime/withredis"
	"github.com/utain/poc/go-realtime/ws"
)

func main() {
	var port uint
	var bus uint
	flag.UintVar(&port, "p", 6000, "Server port default 6000")
	flag.UintVar(&bus, "t", 0, "Pub/sub types 0 (redis), 1 (kafka), 2 (pulsar)")
	flag.Parse()

	var topic = "realtime.go"

	manager := ws.WsManager{
		Clients:   make(map[*ws.Client]bool),
		Regiser:   make(chan *ws.Client),
		Boardcast: make(chan ws.Message),
	}
	go manager.Listen()
	// recieve from redis
	var ibus pubsub.Pubsub
	switch bus {
	case 0:
		ibus = withredis.Open(&manager)
	case 1:
		ibus = withkafka.Open(&manager)
	case 2:
		ibus = withpulsar.Open(&manager)
	}
	defer ibus.Close()

	go ibus.Subscriber(context.Background(), topic)

	server.Start(server.ServerOpts{
		Port: port,
		WriteMessage: func(ctx context.Context, chanId string, message []byte) error {
			return ibus.WriteMessage(ctx, topic, chanId, message)
		},
		Manager: &manager,
	})
}
