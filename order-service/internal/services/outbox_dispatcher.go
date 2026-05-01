package services

import (
	"context"
	"log"
	"time"

	"github.com/osamah22/coffee-service/order-service/internal/models"
	"github.com/osamah22/coffee-service/shared/events"
	"github.com/osamah22/coffee-service/shared/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type OutboxDispatcher struct {
	db       *gorm.DB
	amqpURL  string
	interval time.Duration
	batch    int
}

func NewOutboxDispatcher(db *gorm.DB, amqpURL string, interval time.Duration, batch int) *OutboxDispatcher {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if batch <= 0 {
		batch = 10
	}
	return &OutboxDispatcher{
		db:       db,
		amqpURL:  amqpURL,
		interval: interval,
		batch:    batch,
	}
}

func (d *OutboxDispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		if err := d.dispatch(ctx); err != nil {
			log.Println("outbox dispatch error:", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (d *OutboxDispatcher) dispatch(ctx context.Context) error {
	conn, err := rabbitmq.New(d.amqpURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := rabbitmq.DeclareTopicExchange(conn.Channel, events.OrderExchange); err != nil {
		return err
	}

	publisher := rabbitmq.NewPublisher(conn.Channel)
	var pending []models.OutboxEvent
	if err := d.db.WithContext(ctx).
		Where("published_at IS NULL").
		Order("occurred_at ASC").
		Limit(d.batch).
		Find(&pending).Error; err != nil {
		return err
	}

	for i := range pending {
		if err := d.publishOne(ctx, publisher, &pending[i]); err != nil {
			if updateErr := d.recordFailure(ctx, &pending[i], err); updateErr != nil {
				log.Println("outbox failure update error:", updateErr)
			}
			continue
		}
		if err := d.recordPublished(ctx, pending[i].ID.String()); err != nil {
			return err
		}
	}

	return nil
}

func (d *OutboxDispatcher) publishOne(ctx context.Context, publisher *rabbitmq.Publisher, event *models.OutboxEvent) error {
	return publisher.PublishRaw(ctx, events.OrderExchange, event.RoutingKey, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    event.ID.String(),
		Type:         event.EventType,
		Timestamp:    event.OccurredAt,
		Body:         []byte(event.Payload),
	})
}

func (d *OutboxDispatcher) recordPublished(ctx context.Context, eventID string) error {
	now := time.Now().UTC()
	return d.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ? AND published_at IS NULL", eventID).
		Updates(map[string]any{
			"published_at": now,
			"last_error":   "",
		}).Error
}

func (d *OutboxDispatcher) recordFailure(ctx context.Context, event *models.OutboxEvent, publishErr error) error {
	return d.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ? AND published_at IS NULL", event.ID).
		Updates(map[string]any{
			"attempts":   gorm.Expr("attempts + 1"),
			"last_error": publishErr.Error(),
		}).Error
}
