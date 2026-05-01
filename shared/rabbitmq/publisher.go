package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

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
			Timestamp:   time.Now().UTC(),
			Body:        body,
		},
	)
}

func (p *Publisher) PublishRaw(ctx context.Context, exchange, key string, msg amqp.Publishing) error {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}
	if msg.DeliveryMode == 0 {
		msg.DeliveryMode = amqp.Persistent
	}

	return p.channel.PublishWithContext(ctx, exchange, key, false, false, msg)
}

func DeclareTopicExchange(ch *amqp.Channel, name string) error {
	return ch.ExchangeDeclare(
		name,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	)
}
