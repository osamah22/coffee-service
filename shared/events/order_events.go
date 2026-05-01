package events

import "time"

const (
	OrderExchange          = "coffee.orders"
	OrderCreatedType       = "order.created"
	OrderStatusUpdatedType = "order.status_updated"
)

type OrderCreated struct {
	EventID    string    `json:"event_id"`
	OrderID    string    `json:"order_id"`
	Status     string    `json:"status"`
	Total      int64     `json:"total"`
	OccurredAt time.Time `json:"occurred_at"`
}

type OrderStatusUpdated struct {
	EventID        string    `json:"event_id"`
	OrderID        string    `json:"order_id"`
	PreviousStatus string    `json:"previous_status"`
	Status         string    `json:"status"`
	OccurredAt     time.Time `json:"occurred_at"`
}
