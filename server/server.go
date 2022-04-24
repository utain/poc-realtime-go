package server

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/utain/poc/go-realtime/ws"
)

type ServerOpts struct {
	Port         uint
	WriteMessage func(ctx context.Context, chanId string, message []byte) error
	Manager      *ws.WsManager
}

func Start(opts ServerOpts) {
	serv := fiber.New()
	serv.Get("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			c.Locals("hello", "world")
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	serv.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		chanId := c.Params("id")
		client := &ws.Client{
			ChID:     chanId,
			ClientID: uuid.New().String(),
			Conn:     c,
		}
		opts.Manager.Regiser <- client
		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			mt  int
			msg []byte
			err error
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)
			payload := ws.Message{
				ChID: c.Params("id"),
				Type: mt,
				Body: msg,
				From: client,
			}
			// send to redis
			opts.WriteMessage(context.Background(), chanId, payload.JSON())
		}
	}))
	serv.Listen(fmt.Sprintf(":%d", opts.Port))
}
