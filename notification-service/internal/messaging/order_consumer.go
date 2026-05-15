package messaging

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/osamah22/coffee-service/notification-service/internal/services"
	"github.com/osamah22/coffee-service/shared/events"
	"github.com/osamah22/coffee-service/shared/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

const orderQueue = "notification-service.orders"

type OrderConsumer struct {
	amqpURL       string
	notifications *services.OrderNotificationService
	processed     map[string]struct{}
	mu            sync.Mutex
}

func NewOrderConsumer(amqpURL string, notifications *services.OrderNotificationService) *OrderConsumer {
	return &OrderConsumer{
		amqpURL:       amqpURL,
		notifications: notifications,
		processed:     map[string]struct{}{},
	}
}

func (c *OrderConsumer) Run(ctx context.Context) {
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

func (c *OrderConsumer) consume(ctx context.Context) error {
	conn, err := rabbitmq.New(c.amqpURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := rabbitmq.DeclareTopicExchange(conn.Channel, events.OrderExchange); err != nil {
		return err
	}
	if err := rabbitmq.DeclareTopicExchange(conn.Channel, events.UserExchange); err != nil {
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
	if err := conn.Channel.QueueBind(queue.Name, events.PasswordResetRequestedType, events.UserExchange, false, nil); err != nil {
		return err
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
			c.handle(ctx, delivery)
		}
	}
}

func (c *OrderConsumer) handle(ctx context.Context, delivery amqp.Delivery) {
	eventID := delivery.MessageId
	if eventID == "" {
		eventID = delivery.CorrelationId
	}
	if c.seen(eventID) {
		_ = delivery.Ack(false)
		return
	}

	var err error
	switch delivery.Type {
	case events.OrderCreatedType:
		var event events.OrderCreated
		if err = json.Unmarshal(delivery.Body, &event); err == nil {
			eventID = firstNonEmpty(eventID, event.EventID)
			err = c.notifications.SendOrderCreated(ctx, event)
		}
	case events.OrderStatusUpdatedType:
		var event events.OrderStatusUpdated
		if err = json.Unmarshal(delivery.Body, &event); err == nil {
			eventID = firstNonEmpty(eventID, event.EventID)
			err = c.notifications.SendOrderStatusUpdated(ctx, event)
		}
	case events.PasswordResetRequestedType:
		var event events.PasswordResetRequested
		if err = json.Unmarshal(delivery.Body, &event); err == nil {
			eventID = firstNonEmpty(eventID, event.EventID)
			err = c.notifications.SendPasswordResetRequested(ctx, event)
		}
	default:
		log.Printf("notify: ignored unknown event type %q", delivery.Type)
		_ = delivery.Ack(false)
		return
	}

	if err != nil {
		log.Printf("notify: failed event type=%s: %v", delivery.Type, err)
		_ = delivery.Nack(false, true)
		return
	}

	c.markSeen(eventID)
	_ = delivery.Ack(false)
}

func (c *OrderConsumer) seen(eventID string) bool {
	if eventID == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.processed[eventID]
	return ok
}

func (c *OrderConsumer) markSeen(eventID string) {
	if eventID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.processed[eventID] = struct{}{}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
