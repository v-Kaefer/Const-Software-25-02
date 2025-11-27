package servico

import (
	"context"

	"gorm.io/gorm"
)

// Repository define a interface para operações com Servico
type Repository interface {
	Create(ctx context.Context, s *Servico) error
	GetByID(ctx context.Context, id uint) (*Servico, error)
	GetByNome(ctx context.Context, nome string) (*Servico, error)
	List(ctx context.Context, offset, limit int) ([]Servico, int64, error)
	ListAtivos(ctx context.Context, offset, limit int) ([]Servico, int64, error)
	Update(ctx context.Context, s *Servico) error
	Delete(ctx context.Context, id uint) error
}

type repo struct {
	db *gorm.DB
}

// NewRepo cria uma nova instância de repository
func NewRepo(db *gorm.DB) Repository {
	return &repo{db}
}

// Create insere um novo serviço
func (r *repo) Create(ctx context.Context, s *Servico) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetByID busca um serviço por ID
func (r *repo) GetByID(ctx context.Context, id uint) (*Servico, error) {
	var s Servico
	err := r.db.WithContext(ctx).First(&s, id).Error
	return &s, err
}

// GetByNome busca um serviço por nome
func (r *repo) GetByNome(ctx context.Context, nome string) (*Servico, error) {
	var s Servico
	err := r.db.WithContext(ctx).Where("nome = ?", nome).First(&s).Error
	return &s, err
}

// List lista todos os serviços com paginação
func (r *repo) List(ctx context.Context, offset, limit int) ([]Servico, int64, error) {
	var servicos []Servico
	var total int64

	err := r.db.WithContext(ctx).
		Model(&Servico{}).
		Count(&total).
		Offset(offset).
		Limit(limit).
		Order("id DESC").
		Find(&servicos).Error

	return servicos, total, err
}

// ListAtivos lista apenas serviços ativos
func (r *repo) ListAtivos(ctx context.Context, offset, limit int) ([]Servico, int64, error) {
	var servicos []Servico
	var total int64

	err := r.db.WithContext(ctx).
		Model(&Servico{}).
		Where("ativo = true").
		Count(&total).
		Offset(offset).
		Limit(limit).
		Order("id DESC").
		Find(&servicos).Error

	return servicos, total, err
}

// Update atualiza um serviço
func (r *repo) Update(ctx context.Context, s *Servico) error {
	return r.db.WithContext(ctx).Save(s).Error
}

// Delete deleta um serviço
func (r *repo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Servico{}, id).Error
}
