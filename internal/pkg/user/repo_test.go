package user_test

import (
	"context"
	"testing"
	"github.com/v-Kaefer/Const-Software-25-02/internal/pkg/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	if err := db.AutoMigrate(&user.User{}); err != nil { t.Fatal(err) }
	return db
}

func TestRepo_CreateAndFind(t *testing.T) {
	db := newTestDB(t)
	repo := user.NewRepo(db)

	ctx := context.Background()
	u := &user.User{Email: "a@b.com", Name: "Alice"}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.FindByEmail(ctx, "a@b.com")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Email != u.Email || got.Name != u.Name {
		t.Fatalf("mismatch: got=%+v want=%+v", got, u)
	}
}
