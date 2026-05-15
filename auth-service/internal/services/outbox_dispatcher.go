package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/osamah22/coffee-service/auth-service/internal/models"
	"github.com/osamah22/coffee-service/shared/events"
	"github.com/osamah22/coffee-service/shared/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type OutboxDispatcher struct {
	db          *gorm.DB
	amqpURL     string
	pollEvery   time.Duration
	batchSize   int
	maxAttempts int
}

func NewOutboxDispatcher(db *gorm.DB, amqpURL string, pollEvery time.Duration, batchSize int) *OutboxDispatcher {
	return &OutboxDispatcher{
		db:          db,
		amqpURL:     amqpURL,
		pollEvery:   pollEvery,
		batchSize:   batchSize,
		maxAttempts: 10,
	}
}

func (d *OutboxDispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(d.pollEvery)
	defer ticker.Stop()

	for {
		if err := d.flush(ctx); err != nil {
			log.Println("auth outbox dispatch error:", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (d *OutboxDispatcher) flush(ctx context.Context) error {
	conn, err := rabbitmq.New(d.amqpURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := rabbitmq.DeclareTopicExchange(conn.Channel, events.UserExchange); err != nil {
		return err
	}
	publisher := rabbitmq.NewPublisher(conn.Channel)

	var batch []models.OutboxEvent
	if err := d.db.WithContext(ctx).
		Where("published_at IS NULL AND attempts < ?", d.maxAttempts).
		Order("occurred_at ASC").
		Limit(d.batchSize).
		Find(&batch).Error; err != nil {
		return err
	}

	for i := range batch {
		event := batch[i]
		if err := d.publishOne(ctx, publisher, &event); err != nil {
			if updateErr := d.recordFailure(ctx, &event, err); updateErr != nil {
				log.Println("auth outbox failure update error:", updateErr)
			}
			continue
		}

		if err := d.recordPublished(ctx, event.ID.String()); err != nil {
			return err
		}
	}

	return nil
}

func (d *OutboxDispatcher) publishOne(ctx context.Context, publisher *rabbitmq.Publisher, event *models.OutboxEvent) error {
	return publisher.PublishRaw(ctx, events.UserExchange, event.RoutingKey, amqp.Publishing{
		ContentType:  "application/json",
		MessageId:    event.ID.String(),
		Type:         event.EventType,
		Timestamp:    event.OccurredAt,
		DeliveryMode: amqp.Persistent,
		Body:         []byte(event.Payload),
	})
}

func (d *OutboxDispatcher) recordPublished(ctx context.Context, eventID string) error {
	publishedAt := time.Now().UTC()
	return d.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ? AND published_at IS NULL", eventID).
		Updates(map[string]any{
			"published_at": publishedAt,
			"last_error":   "",
		}).Error
}

func (d *OutboxDispatcher) recordFailure(ctx context.Context, event *models.OutboxEvent, publishErr error) error {
	return d.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ? AND published_at IS NULL", event.ID).
		Updates(map[string]any{
			"attempts":   gorm.Expr("attempts + 1"),
			"last_error": truncateError(publishErr),
		}).Error
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	msg := fmt.Sprintf("%v", err)
	if len(msg) > 1000 {
		return msg[:1000]
	}
	return msg
}
