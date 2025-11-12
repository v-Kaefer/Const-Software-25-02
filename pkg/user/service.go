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

func (s *Service) GetByID(ctx context.Context, id uint) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id uint, email, name string) (*User, error) {
	var out *User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		r := s.repo.WithTx(tx)
		u, err := r.FindByID(ctx, id)
		if err != nil {
			return err
		}
		u.Email = email
		u.Name = name
		if err := r.Update(ctx, u); err != nil {
			return err
		}
		out = u
		return nil
	})
	return out, err
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		r := s.repo.WithTx(tx)
		// Check if user exists first
		_, err := r.FindByID(ctx, id)
		if err != nil {
			return err
		}
		return r.Delete(ctx, id)
	})
}
