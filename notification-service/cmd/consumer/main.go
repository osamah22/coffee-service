package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/osamah22/coffee-service/shared/events"
	"github.com/osamah22/coffee-service/shared/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

const orderQueue = "notification-service.orders"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	consumer := newConsumer(envOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"))
	consumer.Run(ctx)
}

type consumer struct {
	amqpURL   string
	processed map[string]struct{}
	mu        sync.Mutex
}

func newConsumer(amqpURL string) *consumer {
	return &consumer{
		amqpURL:   amqpURL,
		processed: map[string]struct{}{},
	}
}

func (c *consumer) Run(ctx context.Context) {
	for {
		if err := c.consume(ctx); err != nil {
			log.Println("consumer error:", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *consumer) consume(ctx context.Context) error {
	conn, err := rabbitmq.New(c.amqpURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := rabbitmq.DeclareTopicExchange(conn.Channel, events.OrderExchange); err != nil {
		return err
	}

	queue, err := conn.Channel.QueueDeclare(orderQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	for _, key := range []string{events.OrderCreatedType, events.OrderStatusUpdatedType} {
		if err := conn.Channel.QueueBind(queue.Name, key, events.OrderExchange, false, nil); err != nil {
			return err
		}
	}

	if err := conn.Channel.Qos(10, 0, false); err != nil {
		return err
	}

	deliveries, err := conn.Channel.Consume(queue.Name, "notification-service", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Println("notification consumer waiting for order events")
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}
			c.handle(delivery)
		}
	}
}

func (c *consumer) handle(delivery amqp.Delivery) {
	if c.seen(delivery.MessageId) {
		_ = delivery.Ack(false)
		return
	}

	switch delivery.Type {
	case events.OrderCreatedType:
		var event events.OrderCreated
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			_ = delivery.Nack(false, false)
			return
		}
		log.Printf("notify: order %s created with status %s total=%d", event.OrderID, event.Status, event.Total)
	case events.OrderStatusUpdatedType:
		var event events.OrderStatusUpdated
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			_ = delivery.Nack(false, false)
			return
		}
		log.Printf("notify: order %s status changed from %s to %s", event.OrderID, event.PreviousStatus, event.Status)
	default:
		log.Printf("notify: ignored unknown event type %q", delivery.Type)
	}

	c.markSeen(delivery.MessageId)
	_ = delivery.Ack(false)
}

func (c *consumer) seen(eventID string) bool {
	if eventID == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.processed[eventID]
	return ok
}

func (c *consumer) markSeen(eventID string) {
	if eventID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.processed[eventID] = struct{}{}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
