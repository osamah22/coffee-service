package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/order-service/internal/services"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
)

type ProductHandler struct {
	svc *services.ProductService
}

func NewProductHandler(svc *services.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) Register(router gin.IRouter, authMiddleware *sharedauth.Middleware) {
	products := router.Group("/products")
	products.GET("", authMiddleware.RequireRole(sharedauth.RoleCustomer, sharedauth.RoleStaff, sharedauth.RoleAdmin), h.list)
}

func (h *ProductHandler) list(c *gin.Context) {
	products, err := h.svc.ListAll(c.Request.Context())
	if err != nil {
		respondError(c, err, http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, products)
}
