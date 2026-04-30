package dtos

import (
	"time"

	"github.com/osamah22/coffee-service/order-service/internal/models"
)

type CreateOrderRequest struct {
	Items []struct {
		ProductID string `json:"product_id" binding:"required,uuid"`
		Quantity  int    `json:"quantity" binding:"required,gt=0"`
	} `json:"items" binding:"required,dive,required"`
}

type UpdateOrderRequest struct {
	Items []struct {
		ProductID string `json:"product_id" binding:"required,uuid"`
		Quantity  int    `json:"quantity" binding:"required,gt=0"`
	} `json:"items" binding:"required,dive,required"`
}

type OrderResponse struct {
	ID        string             `json:"id"`
	Items     []LineItemResponse `json:"items"`
	Total     int64              `json:"total"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}

type LineItemResponse struct {
	ProductID    string `json:"product_id"`
	Quantity     int    `json:"quantity"`
	PriceInKurus int64  `json:"price_in_kurus"`
}

func ToOrderResponse(order *models.Order) OrderResponse {
	items := make([]LineItemResponse, len(order.Items))

	for i, item := range order.Items {
		items[i] = LineItemResponse{
			ProductID:    item.ProductID.String(),
			Quantity:     item.Quantity,
			PriceInKurus: item.PriceInKurus,
		}
	}

	return OrderResponse{
		ID:        order.ID.String(),
		Items:     items,
		Total:     order.Total,
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt,
	}
}
