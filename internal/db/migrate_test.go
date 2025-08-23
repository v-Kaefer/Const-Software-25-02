package db_test

import (
	"testing"
	appdb "github.com/v-Kaefer/Const-Software-25-02/internal/db"
	"github.com/v-Kaefer/Const-Software-25-02/internal/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAutoMigrate(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil { t.Fatal(err) }

	if err := appdb.AutoMigrate(gdb); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	if !gdb.Migrator().HasTable(&user.User{}) {
		t.Fatalf("expected users table")
	}
}
