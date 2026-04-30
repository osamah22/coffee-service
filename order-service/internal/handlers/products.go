package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/order-service/internal/dtos"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"github.com/osamah22/coffee-service/order-service/internal/services"
)

type ProductHandler struct {
	svc *services.ProductService
}

func NewProductHandler(svc *services.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) Register(router gin.IRouter) {
	products := router.Group("/products")
	products.GET("", h.list)
	products.GET("/:id", h.get)
	products.POST("", h.create)
	products.PUT("/:id", h.update)
	products.DELETE("/:id", h.delete)
}

func (h *ProductHandler) list(c *gin.Context) {
	products, err := h.svc.ListAll(c.Request.Context())
	if err != nil {
		respondError(c, err, http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) get(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	product, err := h.svc.Find(c.Request.Context(), id)
	if err != nil {
		respondError(c, err, http.StatusInternalServerError, services.ErrProductNotFound)
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) create(c *gin.Context) {
	var req dtos.CreateProductRequest
	if !bind(c, &req) {
		return
	}

	available := true
	if req.Available != nil {
		available = *req.Available
	}

	product := &models.Product{
		Name:         req.Name,
		Category:     models.Category(req.Category),
		PriceInKurus: req.PriceInKurus,
		Available:    available,
	}

	product, err := h.svc.AddProduct(c.Request.Context(), product)
	if err != nil {
		respondError(c, err, http.StatusBadRequest)
		return
	}

	c.Header("Location", fmt.Sprintf("/products/%s", product.ID))
	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) update(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var req dtos.UpdateProductRequest
	if !bind(c, &req) {
		return
	}

	available := true
	if req.Available != nil {
		available = *req.Available
	}

	product := &models.Product{
		ID:           id,
		Name:         req.Name,
		Category:     models.Category(req.Category),
		PriceInKurus: req.PriceInKurus,
		Available:    available,
	}

	product, err := h.svc.UpdateProduct(c.Request.Context(), product)
	if err != nil {
		respondError(c, err, http.StatusBadRequest, services.ErrProductNotFound)
		return
	}

	c.Header("Location", fmt.Sprintf("/products/%s", product.ID))
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) delete(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	if err := h.svc.DeleteProduct(c.Request.Context(), id); err != nil {
		respondError(c, err, http.StatusInternalServerError, services.ErrProductNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}
