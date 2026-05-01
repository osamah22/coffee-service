package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/osamah22/coffee-service/order-service/internal/dtos"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"github.com/osamah22/coffee-service/order-service/internal/services"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
)

type OrderHandler struct {
	orderSvc   *services.OrderService
	productSvc *services.ProductService
}

func NewOrderHandler(orderSvc *services.OrderService, productSvc *services.ProductService) *OrderHandler {
	return &OrderHandler{orderSvc: orderSvc, productSvc: productSvc}
}

func (h *OrderHandler) Register(router gin.IRouter, authMiddleware *sharedauth.Middleware) {
	orders := router.Group("/orders")
	orders.GET("", authMiddleware.RequireRole(sharedauth.RoleAdmin), h.list)
	orders.GET("/:id", authMiddleware.RequireRole(sharedauth.RoleAdmin), h.get)
	orders.POST("", authMiddleware.RequireRole(sharedauth.RoleGuest, sharedauth.RoleUser, sharedauth.RoleAdmin), h.create)
	orders.POST("/:id/complete", authMiddleware.RequireRole(sharedauth.RoleAdmin), h.complete)
	orders.POST("/:id/cancel", authMiddleware.RequireRole(sharedauth.RoleAdmin), h.cancel)
	orders.DELETE("/:id", authMiddleware.RequireRole(sharedauth.RoleAdmin), h.delete)
}

func (h *OrderHandler) list(c *gin.Context) {
	orders, err := h.orderSvc.ListOrders(c.Request.Context())
	if err != nil {
		respondError(c, err, http.StatusInternalServerError)
		return
	}

	response := make([]dtos.OrderResponse, len(orders))
	for i := range orders {
		response[i] = dtos.ToOrderResponse(&orders[i])
	}
	c.JSON(http.StatusOK, response)
}

func (h *OrderHandler) get(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	order, err := h.orderSvc.GetOrder(c.Request.Context(), id)
	if err != nil {
		respondError(c, err, http.StatusInternalServerError, services.ErrOrderNotFound)
		return
	}
	c.JSON(http.StatusOK, dtos.ToOrderResponse(order))
}

func (h *OrderHandler) create(c *gin.Context) {
	var req dtos.CreateOrderRequest
	if !bind(c, &req) {
		return
	}

	order := &models.Order{}
	for _, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product_id"})
			return
		}

		product, err := h.productSvc.Find(c.Request.Context(), productID)
		if err != nil {
			respondError(c, err, http.StatusBadRequest, services.ErrProductNotFound)
			return
		}

		order.Items = append(order.Items, models.LineItem{
			ProductID:    productID,
			PriceInKurus: product.PriceInKurus,
			Quantity:     item.Quantity,
		})
	}

	order, err := h.orderSvc.CreateOrder(c.Request.Context(), order)
	if err != nil {
		respondError(c, err, http.StatusBadRequest)
		return
	}

	c.Header("Location", fmt.Sprintf("/orders/%s", order.ID))
	c.JSON(http.StatusCreated, dtos.ToOrderResponse(order))
}

func (h *OrderHandler) complete(c *gin.Context) {
	h.updateStatus(c, models.StatusCompleted)
}

func (h *OrderHandler) cancel(c *gin.Context) {
	h.updateStatus(c, models.StatusCancelled)
}

func (h *OrderHandler) updateStatus(c *gin.Context, status models.OrderStatus) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	order, err := h.orderSvc.UpdateStatus(c.Request.Context(), id, status)
	if err != nil {
		respondError(c, err, http.StatusBadRequest, services.ErrOrderNotFound)
		return
	}

	c.Header("Location", fmt.Sprintf("/orders/%s", order.ID))
	c.JSON(http.StatusOK, dtos.ToOrderResponse(order))
}

func (h *OrderHandler) delete(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	if err := h.orderSvc.DeleteOrder(c.Request.Context(), id); err != nil {
		respondError(c, err, http.StatusInternalServerError, services.ErrOrderNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}
