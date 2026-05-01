package handlers

import (
	"fmt"
	"net/http"
	"strings"

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
	orders.GET("", authMiddleware.RequireRole(sharedauth.RoleBarista, sharedauth.RoleAdmin), h.list)
	orders.GET("/mine", authMiddleware.RequireRole(sharedauth.RoleGuest, sharedauth.RoleUser, sharedauth.RoleAdmin), h.listMine)
	orders.GET("/:id", authMiddleware.RequireRole(sharedauth.RoleBarista, sharedauth.RoleAdmin), h.get)
	orders.POST("", authMiddleware.RequireRole(sharedauth.RoleGuest, sharedauth.RoleUser, sharedauth.RoleAdmin), h.create)
	orders.POST("/:id/ready", authMiddleware.RequireRole(sharedauth.RoleBarista, sharedauth.RoleAdmin), h.ready)
	orders.POST("/:id/complete", authMiddleware.RequireRole(sharedauth.RoleBarista, sharedauth.RoleAdmin), h.complete)
	orders.POST("/:id/cancel", authMiddleware.RequireRole(sharedauth.RoleBarista, sharedauth.RoleAdmin), h.cancel)
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

func (h *OrderHandler) listMine(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	if claims, ok := sharedauth.CurrentClaims(c); ok && claims.Email != "" {
		email = claims.Email
	}
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_email is required"})
		return
	}

	orders, err := h.orderSvc.ListOrdersByEmail(c.Request.Context(), email)
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

	order := &models.Order{CustomerEmail: strings.TrimSpace(req.CustomerEmail)}
	if claims, ok := sharedauth.CurrentClaims(c); ok && claims.Email != "" && order.CustomerEmail == "" {
		order.CustomerEmail = claims.Email
	}
	if order.CustomerEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_email is required"})
		return
	}

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
			ProductName:  product.Name,
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

func (h *OrderHandler) ready(c *gin.Context) {
	h.updateStatus(c, models.StatusReady)
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
		respondError(c, err, http.StatusBadRequest, services.ErrOrderNotFound, services.ErrInvalidStatusTransition)
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
