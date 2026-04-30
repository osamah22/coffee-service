package seed

import (
	"context"
	"testing"

	"github.com/osamah22/coffee-service/order-service/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCoffeeMenuSeedsProducts(t *testing.T) {
	db := newTestDB(t)

	if err := CoffeeMenu(context.Background(), db); err != nil {
		t.Fatalf("expected seeding to succeed, got %v", err)
	}

	var products []models.Product
	if err := db.Order("name").Find(&products).Error; err != nil {
		t.Fatalf("expected products query to succeed, got %v", err)
	}

	if len(products) != len(CoffeeProducts) {
		t.Fatalf("expected %d products, got %d", len(CoffeeProducts), len(products))
	}

	for _, product := range products {
		if product.ImagePath == "" {
			t.Fatalf("expected product %q to have an image path", product.Name)
		}
	}
}

func TestCoffeeMenuIsIdempotent(t *testing.T) {
	db := newTestDB(t)

	if err := CoffeeMenu(context.Background(), db); err != nil {
		t.Fatalf("expected first seed to succeed, got %v", err)
	}
	if err := CoffeeMenu(context.Background(), db); err != nil {
		t.Fatalf("expected second seed to succeed, got %v", err)
	}

	var count int64
	if err := db.Model(&models.Product{}).Count(&count).Error; err != nil {
		t.Fatalf("expected count query to succeed, got %v", err)
	}

	if count != int64(len(CoffeeProducts)) {
		t.Fatalf("expected %d products after reseed, got %d", len(CoffeeProducts), count)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&models.Product{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}
