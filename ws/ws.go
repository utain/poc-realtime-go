package ws

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/websocket/v2"
)

type Client struct {
	ChID     string          `json:"chid"`
	ClientID string          `json:"clientID"`
	Conn     *websocket.Conn `json:"-"`
}

type Message struct {
	ChID string  `json:"chid"`
	Type int     `json:"mst"`
	Body []byte  `json:"body"`
	From *Client `json:"from"`
}

func ParseMessage(data []byte) (out Message, err error) {
	err = json.Unmarshal(data, &out)
	return
}

func (s *Message) JSON() []byte {
	b, _ := json.Marshal(s)
	return b
}

type WsManager struct {
	Clients   map[*Client]bool
	Regiser   chan *Client
	Boardcast chan Message
}

func (s *WsManager) Listen() {
	for {
		select {
		case c := <-s.Regiser:
			fmt.Println("Client:", c.ChID, c.ClientID)
			s.Clients[c] = true
		case c := <-s.Boardcast:
			// listen from redis/kafka
			fmt.Println("Boardcast", c.From.ClientID, c.From.ChID)
			for cl, v := range s.Clients {
				if !v {
					continue
				}
				if err := cl.Conn.WriteMessage(c.Type, c.Body); err != nil {
					log.Println("write:", err)
				}
			}
		}
	}
}
