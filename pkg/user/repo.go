package user

import (
	"context"
	"gorm.io/gorm"
)

// Repo define operações de persistência de usuários.
type Repo interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uint) error
	WithTx(tx *gorm.DB) Repo
}

type repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) Repo { return &repo{db: db} }

func (r *repo) WithTx(tx *gorm.DB) Repo { return &repo{db: tx} }

func (r *repo) Create(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *repo) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repo) FindByID(ctx context.Context, id uint) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repo) List(ctx context.Context) ([]User, error) {
	var users []User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (r *repo) Update(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *repo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}
