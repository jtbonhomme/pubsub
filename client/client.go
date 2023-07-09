package client

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

// Client is a pubsub service client.
type Client struct {
	Name       string          // Client nice name.
	ID         string          // Client unique ID.
	Connection *websocket.Conn // Client websocket connection.
}

// New create a pubsub client. It generates an ID and assign it a name.
func New(name string, conn *websocket.Conn) *Client {
	return &Client{
		Name:       name,
		ID:         uuid.NewString(),
		Connection: conn,
	}
}

// Send sends a message to a client.
func (client *Client) Send(message []byte) error {
	return websocket.Message.Send(client.Connection, message)
}

// String is the Client structure stringer function.
func (client *Client) String() string {
	return fmt.Sprintf("pubsub client %s (%s)", client.Name, client.ID)
}
