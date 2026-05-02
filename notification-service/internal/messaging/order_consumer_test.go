package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	notificationservices "github.com/osamah22/coffee-service/notification-service/internal/services"
	"github.com/osamah22/coffee-service/shared/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type fakeAcknowledger struct {
	acks    []uint64
	nacks   []nackCall
	rejects []uint64
}

type nackCall struct {
	tag     uint64
	requeue bool
}

func (a *fakeAcknowledger) Ack(tag uint64, _ bool) error {
	a.acks = append(a.acks, tag)
	return nil
}

func (a *fakeAcknowledger) Nack(tag uint64, _ bool, requeue bool) error {
	a.nacks = append(a.nacks, nackCall{tag: tag, requeue: requeue})
	return nil
}

func (a *fakeAcknowledger) Reject(tag uint64, _ bool) error {
	a.rejects = append(a.rejects, tag)
	return nil
}

type fakeSender struct {
	emails []notificationservices.Email
	err    error
}

func (s *fakeSender) Send(_ context.Context, email notificationservices.Email) error {
	if s.err != nil {
		return s.err
	}
	s.emails = append(s.emails, email)
	return nil
}

func TestHandleOrderCreatedSendsEmailAndAcks(t *testing.T) {
	sender := &fakeSender{}
	consumer := NewOrderConsumer(
		"amqp://example/",
		notificationservices.NewOrderNotificationService(sender, "fallback@example.test"),
	)
	acker := &fakeAcknowledger{}
	body := mustJSON(t, events.OrderCreated{
		EventID:       "event-1",
		OrderID:       "order-1",
		CustomerEmail: "customer@example.test",
		Status:        "preparing",
	})

	consumer.handle(context.Background(), amqp.Delivery{
		Acknowledger: acker,
		DeliveryTag:  42,
		Type:         events.OrderCreatedType,
		Body:         body,
	})

	if len(acker.acks) != 1 || acker.acks[0] != 42 {
		t.Fatalf("expected ack for tag 42, got %#v", acker.acks)
	}
	if len(acker.nacks) != 0 {
		t.Fatalf("expected no nacks, got %#v", acker.nacks)
	}
	if len(sender.emails) != 1 {
		t.Fatalf("expected one email, got %d", len(sender.emails))
	}
	if !consumer.seen("event-1") {
		t.Fatal("expected event to be marked processed")
	}
}

func TestHandleDuplicateEventAcksWithoutSending(t *testing.T) {
	sender := &fakeSender{}
	consumer := NewOrderConsumer(
		"amqp://example/",
		notificationservices.NewOrderNotificationService(sender, "fallback@example.test"),
	)
	consumer.markSeen("event-1")
	acker := &fakeAcknowledger{}

	consumer.handle(context.Background(), amqp.Delivery{
		Acknowledger: acker,
		DeliveryTag:  9,
		MessageId:    "event-1",
		Type:         events.OrderCreatedType,
		Body:         []byte(`{}`),
	})

	if len(acker.acks) != 1 || acker.acks[0] != 9 {
		t.Fatalf("expected duplicate ack for tag 9, got %#v", acker.acks)
	}
	if len(sender.emails) != 0 {
		t.Fatalf("expected no email for duplicate, got %d", len(sender.emails))
	}
}

func TestHandleBadPayloadNacksWithRequeue(t *testing.T) {
	consumer := NewOrderConsumer(
		"amqp://example/",
		notificationservices.NewOrderNotificationService(&fakeSender{}, "fallback@example.test"),
	)
	acker := &fakeAcknowledger{}

	consumer.handle(context.Background(), amqp.Delivery{
		Acknowledger: acker,
		DeliveryTag:  7,
		Type:         events.OrderCreatedType,
		Body:         []byte(`{`),
	})

	if len(acker.nacks) != 1 {
		t.Fatalf("expected one nack, got %#v", acker.nacks)
	}
	if acker.nacks[0].tag != 7 || !acker.nacks[0].requeue {
		t.Fatalf("expected nack tag 7 with requeue, got %#v", acker.nacks[0])
	}
	if len(acker.acks) != 0 {
		t.Fatalf("expected no ack, got %#v", acker.acks)
	}
}

func TestHandleSenderErrorNacksWithRequeue(t *testing.T) {
	consumer := NewOrderConsumer(
		"amqp://example/",
		notificationservices.NewOrderNotificationService(&fakeSender{err: errors.New("smtp failed")}, "fallback@example.test"),
	)
	acker := &fakeAcknowledger{}
	body := mustJSON(t, events.OrderStatusUpdated{
		EventID:        "event-2",
		OrderID:        "order-1",
		PreviousStatus: "preparing",
		Status:         "ready",
	})

	consumer.handle(context.Background(), amqp.Delivery{
		Acknowledger: acker,
		DeliveryTag:  8,
		Type:         events.OrderStatusUpdatedType,
		Body:         body,
	})

	if len(acker.nacks) != 1 {
		t.Fatalf("expected one nack, got %#v", acker.nacks)
	}
	if acker.nacks[0].tag != 8 || !acker.nacks[0].requeue {
		t.Fatalf("expected nack tag 8 with requeue, got %#v", acker.nacks[0])
	}
	if consumer.seen("event-2") {
		t.Fatal("failed event should not be marked processed")
	}
}

func TestHandleUnknownEventAcksAndIgnores(t *testing.T) {
	sender := &fakeSender{}
	consumer := NewOrderConsumer(
		"amqp://example/",
		notificationservices.NewOrderNotificationService(sender, "fallback@example.test"),
	)
	acker := &fakeAcknowledger{}

	consumer.handle(context.Background(), amqp.Delivery{
		Acknowledger: acker,
		DeliveryTag:  3,
		Type:         "unknown.event",
		Body:         []byte(`{}`),
	})

	if len(acker.acks) != 1 || acker.acks[0] != 3 {
		t.Fatalf("expected ack for unknown event, got %#v", acker.acks)
	}
	if len(sender.emails) != 0 {
		t.Fatalf("expected no emails for unknown event, got %d", len(sender.emails))
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "a", "b"); got != "a" {
		t.Fatalf("expected first non-empty value, got %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Fatalf("expected empty value, got %q", got)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()

	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal test payload: %v", err)
	}
	return payload
}
