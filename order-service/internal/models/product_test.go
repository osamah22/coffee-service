package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateProduct(t *testing.T) {
	tests := []struct {
		name    string
		product Product
		wantErr bool
	}{
		{
			name: "valid hot product",
			product: Product{
				Name:         "Latte",
				Category:     Hot,
				PriceInKurus: 7500,
				ImagePath:    "/images/products/latte.jpg",
				Available:    true,
			},
			wantErr: false,
		},
		{
			name: "valid cold product",
			product: Product{
				Name:         "Iced Americano",
				Category:     Cold,
				PriceInKurus: 6500,
				ImagePath:    "/images/products/iced-americano.jpg",
				Available:    true,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			product: Product{
				Name:         "",
				Category:     Hot,
				PriceInKurus: 7500,
				ImagePath:    "/images/products/latte.jpg",
				Available:    true,
			},
			wantErr: true,
		},
		{
			name: "negative price",
			product: Product{
				Name:         "Latte",
				Category:     Hot,
				PriceInKurus: -1,
				ImagePath:    "/images/products/latte.jpg",
				Available:    true,
			},
			wantErr: true,
		},
		{
			name: "invalid category",
			product: Product{
				Name:         "Latte",
				Category:     Category("invalid"),
				PriceInKurus: 7500,
				ImagePath:    "/images/products/latte.jpg",
				Available:    true,
			},
			wantErr: true,
		},
		{
			name: "empty image path",
			product: Product{
				Name:         "Latte",
				Category:     Hot,
				PriceInKurus: 7500,
				ImagePath:    "",
				Available:    true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProduct(&tt.product)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestCreateProductWithGorm(t *testing.T) {
	db := newTestDB(t)

	product := Product{
		Name:         "Cappuccino",
		Category:     Hot,
		PriceInKurus: 8000,
		ImagePath:    "/images/products/cappuccino.jpg",
		Available:    true,
	}

	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("expected product create to succeed, got %v", err)
	}

	if product.ID == uuid.Nil {
		t.Fatal("expected product ID to be set")
	}

	var found Product
	if err := db.First(&found, "id = ?", product.ID).Error; err != nil {
		t.Fatalf("expected product to be found, got %v", err)
	}

	if found.Name != product.Name {
		t.Fatalf("expected name %q, got %q", product.Name, found.Name)
	}
}
