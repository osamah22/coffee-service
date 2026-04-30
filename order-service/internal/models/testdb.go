// testdb.go
package models

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// dir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	// db, err := gorm.Open(sqlite.Open(fmt.Sprintf("%s/memory", dir,")), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(&Product{}, &Order{}, &LineItem{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}
