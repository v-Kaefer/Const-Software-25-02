package user

import (
	"context"
	"gorm.io/gorm"
)

// Service orquestra casos de uso para usuários.
type Service struct {
	db   *gorm.DB
	repo Repo
}

func NewService(db *gorm.DB, repo Repo) *Service {
	return &Service{db: db, repo: repo}
}

func (s *Service) Register(ctx context.Context, email, name string) (*User, error) {
	var out *User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		r := s.repo.WithTx(tx)
		u := &User{Email: email, Name: name}
		if err := r.Create(ctx, u); err != nil {
			return err
		}
		out = u
		return nil
	})
	return out, err
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.FindByEmail(ctx, email)
}
