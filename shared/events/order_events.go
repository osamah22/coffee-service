package events

import "time"

const (
	OrderExchange          = "coffee.orders"
	OrderCreatedType       = "order.created"
	OrderStatusUpdatedType = "order.status_updated"
)

type OrderCreated struct {
	EventID       string      `json:"event_id"`
	OrderID       string      `json:"order_id"`
	CustomerEmail string      `json:"customer_email"`
	Status        string      `json:"status"`
	Items         []OrderItem `json:"items"`
	Total         int64       `json:"total"`
	OccurredAt    time.Time   `json:"occurred_at"`
}

type OrderStatusUpdated struct {
	EventID        string    `json:"event_id"`
	OrderID        string    `json:"order_id"`
	CustomerEmail  string    `json:"customer_email"`
	PreviousStatus string    `json:"previous_status"`
	Status         string    `json:"status"`
	OccurredAt     time.Time `json:"occurred_at"`
}

type OrderItem struct {
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	Quantity     int    `json:"quantity"`
	PriceInKurus int64  `json:"price_in_kurus"`
}
