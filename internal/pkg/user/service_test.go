package user_test

import (
	"context"
	"testing"
	"github.com/v-Kaefer/Const-Software-25-02/internal/pkg/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSvc(t *testing.T) *user.Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	if err := db.AutoMigrate(&user.User{}); err != nil { t.Fatal(err) }
	return user.NewService(db, user.NewRepo(db))
}

func TestService_RegisterAndGet(t *testing.T) {
	svc := newSvc(t)

	ctx := context.Background()
	u, err := svc.Register(ctx, "b@b.com", "Bob")
	if err != nil { t.Fatalf("register: %v", err) }

	got, err := svc.GetByEmail(ctx, "b@b.com")
	if err != nil { t.Fatalf("get: %v", err) }

	if got.ID == 0 || got.Email != u.Email {
		t.Fatalf("unexpected: got=%+v", got)
	}
}
