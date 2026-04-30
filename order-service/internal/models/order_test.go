package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestOrderBeforeCreateSetsID(t *testing.T) {
	order := Order{}

	err := order.BeforeCreate(nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if order.ID == uuid.Nil {
		t.Fatal("expected order ID to be set")
	}

	if order.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestLineItemBeforeCreate(t *testing.T) {
	tests := []struct {
		name    string
		item    LineItem
		wantErr bool
	}{
		{
			name: "valid line item",
			item: LineItem{
				ProductID:    uuid.New(),
				Quantity:     2,
				PriceInKurus: 7500,
			},
			wantErr: false,
		},
		{
			name: "zero quantity",
			item: LineItem{
				ProductID:    uuid.New(),
				Quantity:     0,
				PriceInKurus: 7500,
			},
			wantErr: true,
		},
		{
			name: "negative quantity",
			item: LineItem{
				ProductID:    uuid.New(),
				Quantity:     -1,
				PriceInKurus: 7500,
			},
			wantErr: true,
		},
		{
			name: "negative price",
			item: LineItem{
				ProductID:    uuid.New(),
				Quantity:     1,
				PriceInKurus: -100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.BeforeCreate(nil)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !tt.wantErr && tt.item.ID == uuid.Nil {
				t.Fatal("expected line item ID to be set")
			}
		})
	}
}
