package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func New(url string) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Connection{
		Conn:    conn,
		Channel: ch,
	}, nil
}

func (c *Connection) Close() {
	if err := c.Channel.Close(); err != nil {
		log.Println("channel close error:", err)
	}

	if err := c.Conn.Close(); err != nil {
		log.Println("connection close error:", err)
	}
}
