package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
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

	if err := db.AutoMigrate(&models.Product{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestProductServiceAddProduct(t *testing.T) {
	db := newTestDB(t)
	svc := NewProductService(db)

	product := &models.Product{
		Name:         "Latte",
		Category:     models.Hot,
		PriceInKurus: 7500,
		Available:    true,
	}

	created, err := svc.AddProduct(context.Background(), product)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.ID == uuid.Nil {
		t.Fatal("expected product ID to be set")
	}
}

func TestProductServiceListAll(t *testing.T) {
	db := newTestDB(t)
	svc := NewProductService(db)

	_, _ = svc.AddProduct(context.Background(), &models.Product{
		Name:         "Latte",
		Category:     models.Hot,
		PriceInKurus: 7500,
		Available:    true,
	})

	_, _ = svc.AddProduct(context.Background(), &models.Product{
		Name:         "Iced Americano",
		Category:     models.Cold,
		PriceInKurus: 6500,
		Available:    true,
	})

	products, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
}

func TestProductServiceFind(t *testing.T) {
	db := newTestDB(t)
	svc := NewProductService(db)

	created, err := svc.AddProduct(context.Background(), &models.Product{
		Name:         "Cappuccino",
		Category:     models.Hot,
		PriceInKurus: 8000,
		Available:    true,
	})
	if err != nil {
		t.Fatalf("failed to add product: %v", err)
	}

	found, err := svc.Find(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, found.ID)
	}
}

func TestProductServiceFindNotFound(t *testing.T) {
	db := newTestDB(t)
	svc := NewProductService(db)

	_, err := svc.Find(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != productNotFound {
		t.Fatalf("expected %q, got %q", productNotFound, err.Error())
	}
}

func TestProductServiceUpdateProduct(t *testing.T) {
	db := newTestDB(t)
	svc := NewProductService(db)

	created, err := svc.AddProduct(context.Background(), &models.Product{
		Name:         "Mocha",
		Category:     models.Hot,
		PriceInKurus: 9000,
		Available:    true,
	})
	if err != nil {
		t.Fatalf("failed to add product: %v", err)
	}

	created.Name = "Updated Mocha"
	created.Available = false
	created.PriceInKurus = 9500

	updated, err := svc.UpdateProduct(context.Background(), created)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Name != "Updated Mocha" {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}

	found, err := svc.Find(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("failed to find updated product: %v", err)
	}

	if found.Available != false {
		t.Fatal("expected available to be false")
	}

	if found.PriceInKurus != 9500 {
		t.Fatalf("expected price 9500, got %d", found.PriceInKurus)
	}
}

func TestProductServiceUpdateProductNotFound(t *testing.T) {
	db := newTestDB(t)
	svc := NewProductService(db)

	product := &models.Product{
		ID:           uuid.New(),
		Name:         "Ghost Coffee",
		Category:     models.Hot,
		PriceInKurus: 5000,
		Available:    true,
	}

	_, err := svc.UpdateProduct(context.Background(), product)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != productNotFound {
		t.Fatalf("expected %q, got %q", productNotFound, err.Error())
	}
}

func TestProductServiceDeleteProduct(t *testing.T) {
	db := newTestDB(t)
	svc := NewProductService(db)

	created, err := svc.AddProduct(context.Background(), &models.Product{
		Name:         "Espresso",
		Category:     models.Hot,
		PriceInKurus: 4000,
		Available:    true,
	})
	if err != nil {
		t.Fatalf("failed to add product: %v", err)
	}

	if err := svc.DeleteProduct(context.Background(), created.ID); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	_, err = svc.Find(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected product to be deleted")
	}
}

func TestProductServiceDeleteProductNotFound(t *testing.T) {
	db := newTestDB(t)
	svc := NewProductService(db)

	err := svc.DeleteProduct(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errors.New(productNotFound)) && err.Error() != productNotFound {
		t.Fatalf("expected %q, got %q", productNotFound, err.Error())
	}
}
