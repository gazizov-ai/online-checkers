package websocket

import (
	"sync"

	gorilla "github.com/gorilla/websocket"
)

type Client struct {
	UserID string
	Conn   *gorilla.Conn

	writeMu sync.Mutex
}

func NewClient(userID string, conn *gorilla.Conn) *Client {
	return &Client{
		UserID: userID,
		Conn:   conn,
	}
}

func (c *Client) Send(message OutgoingMessage) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.Conn.WriteJSON(message)
}

func (c *Client) Close() error {
	return c.Conn.Close()
}
