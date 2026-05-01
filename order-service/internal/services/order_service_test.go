package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"github.com/osamah22/coffee-service/shared/events"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newOrderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&models.Order{}, &models.LineItem{}, &models.OutboxEvent{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestCreateOrderWritesOutboxEvent(t *testing.T) {
	db := newOrderTestDB(t)
	svc := NewOrderService(db)

	order, err := svc.CreateOrder(context.Background(), &models.Order{
		Items: []models.LineItem{{
			ProductID:    uuid.New(),
			Quantity:     2,
			PriceInKurus: 4500,
		}},
	})
	if err != nil {
		t.Fatalf("expected create order to succeed, got %v", err)
	}

	var outbox models.OutboxEvent
	if err := db.First(&outbox).Error; err != nil {
		t.Fatalf("expected outbox event, got %v", err)
	}

	if outbox.EventType != events.OrderCreatedType {
		t.Fatalf("expected event type %q, got %q", events.OrderCreatedType, outbox.EventType)
	}
	if outbox.AggregateID != order.ID.String() {
		t.Fatalf("expected aggregate id %s, got %s", order.ID, outbox.AggregateID)
	}

	var payload events.OrderCreated
	if err := json.Unmarshal([]byte(outbox.Payload), &payload); err != nil {
		t.Fatalf("payload should be valid JSON: %v", err)
	}
	if payload.OrderID != order.ID.String() {
		t.Fatalf("expected payload order id %s, got %s", order.ID, payload.OrderID)
	}
	if payload.Total != 9000 {
		t.Fatalf("expected total 9000, got %d", payload.Total)
	}
}

func TestUpdateStatusWritesOutboxEvent(t *testing.T) {
	db := newOrderTestDB(t)
	svc := NewOrderService(db)

	order, err := svc.CreateOrder(context.Background(), &models.Order{
		Items: []models.LineItem{{
			ProductID:    uuid.New(),
			Quantity:     1,
			PriceInKurus: 4500,
		}},
	})
	if err != nil {
		t.Fatalf("expected create order to succeed, got %v", err)
	}

	if _, err := svc.UpdateStatus(context.Background(), order.ID, models.StatusCompleted); err != nil {
		t.Fatalf("expected status update to succeed, got %v", err)
	}

	var outbox models.OutboxEvent
	if err := db.Where("event_type = ?", events.OrderStatusUpdatedType).First(&outbox).Error; err != nil {
		t.Fatalf("expected status update outbox event, got %v", err)
	}

	var payload events.OrderStatusUpdated
	if err := json.Unmarshal([]byte(outbox.Payload), &payload); err != nil {
		t.Fatalf("payload should be valid JSON: %v", err)
	}
	if payload.PreviousStatus != string(models.StatusPreparing) {
		t.Fatalf("expected previous status %q, got %q", models.StatusPreparing, payload.PreviousStatus)
	}
	if payload.Status != string(models.StatusCompleted) {
		t.Fatalf("expected status %q, got %q", models.StatusCompleted, payload.Status)
	}
}
