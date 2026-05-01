package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	channel *amqp.Channel
}

func NewPublisher(ch *amqp.Channel) *Publisher {
	return &Publisher{
		channel: ch,
	}
}

func (p *Publisher) Publish(ctx context.Context, exchange, key string, msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
