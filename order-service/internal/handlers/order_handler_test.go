package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/osamah22/coffee-service/order-service/internal/dtos"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"github.com/osamah22/coffee-service/order-service/internal/services"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateOrderUsesServerProductData(t *testing.T) {
	router, db := newOrderHandlerTestRouter(t, sharedauth.Claims{
		Subject: "user-1",
		Email:   "customer@example.test",
		Role:    sharedauth.RoleCustomer,
	})
	product := createTestProduct(t, db, "Latte", 4250)

	response := performJSON(t, router, http.MethodPost, "/orders", map[string]any{
		"items": []map[string]any{
			{"product_id": product.ID.String(), "quantity": 2},
		},
	})

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}

	var body dtos.OrderResponse
	decodeResponse(t, response, &body)

	if body.CustomerEmail != "customer@example.test" {
		t.Fatalf("expected auth email, got %q", body.CustomerEmail)
	}
	if body.Total != 8500 {
		t.Fatalf("expected total from stored product price, got %d", body.Total)
	}
	if len(body.Items) != 1 {
		t.Fatalf("expected one line item, got %d", len(body.Items))
	}
	if body.Items[0].ProductName != "Latte" {
		t.Fatalf("expected stored product name, got %q", body.Items[0].ProductName)
	}
	if body.Items[0].PriceInKurus != 4250 {
		t.Fatalf("expected stored product price, got %d", body.Items[0].PriceInKurus)
	}

	var outboxCount int64
	if err := db.Model(&models.OutboxEvent{}).Count(&outboxCount).Error; err != nil {
		t.Fatalf("count outbox events: %v", err)
	}
	if outboxCount != 1 {
		t.Fatalf("expected one outbox event, got %d", outboxCount)
	}
}

func TestListMineUsesAuthenticatedEmailOverQuery(t *testing.T) {
	router, db := newOrderHandlerTestRouter(t, sharedauth.Claims{
		Subject: "alice",
		Email:   "alice@example.test",
		Role:    sharedauth.RoleCustomer,
	})
	orderSvc := services.NewOrderService(db)
	createTestOrder(t, orderSvc, "alice@example.test")
	createTestOrder(t, orderSvc, "bob@example.test")

	response := performJSON(t, router, http.MethodGet, "/orders/mine?email=bob@example.test", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	var body []dtos.OrderResponse
	decodeResponse(t, response, &body)

	if len(body) != 1 {
		t.Fatalf("expected one order, got %d", len(body))
	}
	if body[0].CustomerEmail != "alice@example.test" {
		t.Fatalf("expected authenticated user's order, got %q", body[0].CustomerEmail)
	}
}

func TestStaffCanAdvanceOrderStatus(t *testing.T) {
	router, db := newOrderHandlerTestRouter(t, sharedauth.Claims{
		Subject: "staff",
		Email:   "staff@example.test",
		Role:    sharedauth.RoleStaff,
	})
	order := createTestOrder(t, services.NewOrderService(db), "customer@example.test")

	response := performJSON(t, router, http.MethodPost, "/staff/orders/"+order.ID.String()+"/ready", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	var body dtos.OrderResponse
	decodeResponse(t, response, &body)
	if body.Status != string(models.StatusReady) {
		t.Fatalf("expected ready status, got %q", body.Status)
	}
}

func TestUserCannotListStaffOrderQueue(t *testing.T) {
	router, _ := newOrderHandlerTestRouter(t, sharedauth.Claims{
		Subject: "user-1",
		Email:   "customer@example.test",
		Role:    sharedauth.RoleCustomer,
	})

	response := performJSON(t, router, http.MethodGet, "/staff/orders", nil)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", response.Code, response.Body.String())
	}
}

func newOrderHandlerTestRouter(t *testing.T, claims sharedauth.Claims) (*gin.Engine, *gorm.DB) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	db := newHandlerTestDB(t)
	productSvc := services.NewProductService(db)
	orderSvc := services.NewOrderService(db)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("auth_claims", claims)
		c.Next()
	})

	authMiddleware := sharedauth.NewMiddleware(sharedauth.Config{})
	NewOrderHandler(orderSvc, productSvc).Register(router, authMiddleware)

	return router, db
}

func newHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&models.Product{}, &models.Order{}, &models.LineItem{}, &models.OutboxEvent{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db
}

func createTestProduct(t *testing.T, db *gorm.DB, name string, priceInKurus int64) models.Product {
	t.Helper()

	product := models.Product{
		Name:         name,
		Category:     models.Hot,
		PriceInKurus: priceInKurus,
		ImagePath:    "/test.png",
		Available:    true,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create test product: %v", err)
	}
	return product
}

func createTestOrder(t *testing.T, orderSvc *services.OrderService, email string) *models.Order {
	t.Helper()

	productID := uuid.New()
	order, err := orderSvc.CreateOrder(context.Background(), &models.Order{
		CustomerEmail: email,
		Items: []models.LineItem{
			{
				ProductID:    productID,
				ProductName:  "Latte",
				Quantity:     1,
				PriceInKurus: 4000,
			},
		},
	})
	if err != nil {
		t.Fatalf("create test order: %v", err)
	}
	return order
}

func performJSON(t *testing.T, router http.Handler, method, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, &body)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, dst any) {
	t.Helper()

	if err := json.NewDecoder(response.Body).Decode(dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
