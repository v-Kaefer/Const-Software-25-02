package user

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
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

// RegisterWithPassword cria um novo usuário com senha
func (s *Service) RegisterWithPassword(ctx context.Context, email, name, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var out *User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		r := s.repo.WithTx(tx)
		u := &User{
			Email:    email,
			Name:     name,
			Password: string(hashedPassword),
		}
		if err := r.Create(ctx, u); err != nil {
			return err
		}
		out = u
		return nil
	})
	return out, err
}

// Authenticate verifica as credenciais do usuário
func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verifica a senha
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.FindByEmail(ctx, email)
}
