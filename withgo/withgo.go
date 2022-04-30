package withgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	wsc "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/utain/poc/go-realtime/pubsub"
	"github.com/utain/poc/go-realtime/ws"
)

type withgo struct {
	conns        []*wsc.Conn
	wsm          *ws.WsManager
	internalPort uint
	addrs        []string
}

type InternalMessage struct {
	Topic string `json:"t"`
	Key   string `json:"k"`
	Value []byte `json:"v"`
}

func (s *InternalMessage) Marshal() []byte {
	b, _ := json.Marshal(s)
	return b
}

type Options struct {
	pubsub.DefaultOptions
	InternalPort uint
}

func Open(opts Options) pubsub.Pubsub {
	engine := &withgo{
		conns:        []*wsc.Conn{},
		wsm:          opts.WsManager,
		internalPort: opts.InternalPort,
		addrs:        opts.Addrs,
	}
	go engine.openConnection()
	return engine
}

func (w *withgo) openConnection() {
	log.Println("Connecting...")
	// TODO implement retry
	time.Sleep(time.Second * 2)

	for _, v := range w.addrs {
		location, err := url.Parse(v)
		if err != nil {
			log.Fatal(err)
			return
		}
		header := http.Header{}
		conn, res, err := wsc.DefaultDialer.Dial(location.String(), header)
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Println("StatusCode:", res.StatusCode)
		w.conns = append(w.conns, conn)
	}
	log.Println("Internal broker connected")
}

func (w *withgo) Close() (err error) {
	for _, conn := range w.conns {
		conn.Close()
	}
	return
}
func (w *withgo) Subscriber(ctx context.Context, topic string) (err error) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/", websocket.New(func(c *websocket.Conn) {
		var (
			mt  int
			msg []byte
			err error
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Fatal("BorkerReadErr:", err)
				break
			}
			log.Printf("BorkerRead: %d, %s\r\n", mt, msg)
			internalMsg := InternalMessage{}
			if err := json.Unmarshal(msg, &internalMsg); err != nil {
				log.Fatal("WrongTopicFormat:", err)
				break
			}
			bmsg, err := ws.ParseMessage(internalMsg.Value)
			if err != nil {
				return
			}
			w.wsm.Boardcast <- bmsg
		}
	}))

	log.Fatal(app.Listen(fmt.Sprintf(":%d", w.internalPort)))
	return
}
func (w *withgo) WriteMessage(ctx context.Context, topic string, chanId string, payload []byte) (err error) {
	// publisher
	for _, conn := range w.conns {
		msg := InternalMessage{
			Topic: topic,
			Key:   chanId,
			Value: payload,
		}
		if err := conn.WriteMessage(websocket.TextMessage, msg.Marshal()); err != nil {
			log.Println("BorkerWriteErr:", err)
		}
	}
	return
}
