package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"github.com/osamah22/coffee-service/shared/events"
	"gorm.io/gorm"
)

var ErrOrderNotFound = errors.New("order_not_found")

type OrderService struct {
	DB *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{DB: db}
}

// CreateOrder takes an order and inserts it to the database, in case order is not valid it will returns an error.
func (svc *OrderService) CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	tx := svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if len(order.Items) == 0 {
		tx.Rollback()
		return nil, errors.New("order must have at least one item")
	}

	var total int64
	for i := range order.Items {
		item := &order.Items[i]

		if item.Quantity < 1 {
			tx.Rollback()
			return nil, errors.New("quantity must be at least 1")
		}

		if item.PriceInKurus < 0 {
			tx.Rollback()
			return nil, errors.New("price cannot be negative")
		}

		total += int64(item.Quantity) * item.PriceInKurus
	}

	order.Total = total
	if order.Status == "" {
		order.Status = models.StatusPreparing
	}

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := enqueueOrderCreated(tx, order); err != nil {
		tx.Rollback()
		return nil, err
	}

	return order, tx.Commit().Error
}

// GetOrder retrieves an order by ID with items.
// if not found it returns an error
func (svc *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	tx := svc.DB.WithContext(ctx).Preload("Items").First(&order, "id = ?", id)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, ErrOrderNotFound
		}
		return nil, tx.Error
	}
	return &order, nil
}

// ListOrders returns all orders with items
func (svc *OrderService) ListOrders(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order
	tx := svc.DB.WithContext(ctx).Preload("Items").Find(&orders)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return orders, nil
}

func (svc *OrderService) UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) (*models.Order, error) {
	var order models.Order
	if err := svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.First(&order, "id = ?", id)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		if result.Error != nil {
			return result.Error
		}

		if order.Status == models.StatusCompleted || order.Status == models.StatusCancelled {
			return errors.New("order-already_finalized")
		}

		previousStatus := order.Status
		order.Status = status

		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		return enqueueOrderStatusUpdated(tx, &order, previousStatus)
	}); err != nil {
		return nil, err
	}

	return svc.GetOrder(ctx, id)
}

// DeleteOrder removes an order by ID.
// If not found it returns a order_not_found error.
func (svc *OrderService) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	return svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&models.LineItem{}, "order_id = ?", id)
		if result.Error != nil {
			return result.Error
		}

		result = tx.Delete(&models.Order{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrOrderNotFound
		}
		return nil
	})
}

func enqueueOrderCreated(tx *gorm.DB, order *models.Order) error {
	eventID := uuid.New()
	occurredAt := time.Now().UTC()
	payload, err := json.Marshal(events.OrderCreated{
		EventID:    eventID.String(),
		OrderID:    order.ID.String(),
		Status:     string(order.Status),
		Total:      order.Total,
		OccurredAt: occurredAt,
	})
	if err != nil {
		return err
	}

	return tx.Create(&models.OutboxEvent{
		ID:            eventID,
		EventType:     events.OrderCreatedType,
		AggregateType: "order",
		AggregateID:   order.ID.String(),
		RoutingKey:    events.OrderCreatedType,
		Payload:       string(payload),
		OccurredAt:    occurredAt,
	}).Error
}

func enqueueOrderStatusUpdated(tx *gorm.DB, order *models.Order, previousStatus models.OrderStatus) error {
	eventID := uuid.New()
	occurredAt := time.Now().UTC()
	payload, err := json.Marshal(events.OrderStatusUpdated{
		EventID:        eventID.String(),
		OrderID:        order.ID.String(),
		PreviousStatus: string(previousStatus),
		Status:         string(order.Status),
		OccurredAt:     occurredAt,
	})
	if err != nil {
		return err
	}

	return tx.Create(&models.OutboxEvent{
		ID:            eventID,
		EventType:     events.OrderStatusUpdatedType,
		AggregateType: "order",
		AggregateID:   order.ID.String(),
		RoutingKey:    events.OrderStatusUpdatedType,
		Payload:       string(payload),
		OccurredAt:    occurredAt,
	}).Error
}
