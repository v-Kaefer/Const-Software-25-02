package agendamento

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository define a interface para operações com Agendamento
type Repository interface {
	Create(ctx context.Context, a *Agendamento) error
	GetByID(ctx context.Context, id uint) (*Agendamento, error)
	ListByCliente(ctx context.Context, clienteID uint, offset, limit int) ([]Agendamento, int64, error)
	ListAll(ctx context.Context, offset, limit int) ([]Agendamento, int64, error)
	ListByStatus(ctx context.Context, status Status, offset, limit int) ([]Agendamento, int64, error)
	Update(ctx context.Context, a *Agendamento) error
	Delete(ctx context.Context, id uint) error
	CheckConflict(ctx context.Context, servicoID uint, dataHora time.Time) (bool, error)
}

type repo struct {
	db *gorm.DB
}

// NewRepo cria uma nova instância de repository
func NewRepo(db *gorm.DB) Repository {
	return &repo{db}
}

// Create insere um novo agendamento
func (r *repo) Create(ctx context.Context, a *Agendamento) error {
	return r.db.WithContext(ctx).Create(a).Error
}

// GetByID busca um agendamento por ID
func (r *repo) GetByID(ctx context.Context, id uint) (*Agendamento, error) {
	var a Agendamento
	err := r.db.WithContext(ctx).
		Preload("Cliente").
		Preload("Servico").
		First(&a, id).Error
	return &a, err
}

// ListByCliente lista agendamentos de um cliente
func (r *repo) ListByCliente(ctx context.Context, clienteID uint, offset, limit int) ([]Agendamento, int64, error) {
	var agendamentos []Agendamento
	var total int64

	err := r.db.WithContext(ctx).
		Model(&Agendamento{}).
		Where("cliente_id = ?", clienteID).
		Count(&total).
		Offset(offset).
		Limit(limit).
		Order("data_hora DESC").
		Preload("Servico").
		Find(&agendamentos).Error

	return agendamentos, total, err
}

// ListAll lista todos os agendamentos
func (r *repo) ListAll(ctx context.Context, offset, limit int) ([]Agendamento, int64, error) {
	var agendamentos []Agendamento
	var total int64

	err := r.db.WithContext(ctx).
		Model(&Agendamento{}).
		Count(&total).
		Offset(offset).
		Limit(limit).
		Order("data_hora DESC").
		Preload("Cliente").
		Preload("Servico").
		Find(&agendamentos).Error

	return agendamentos, total, err
}

// ListByStatus lista agendamentos por status
func (r *repo) ListByStatus(ctx context.Context, status Status, offset, limit int) ([]Agendamento, int64, error) {
	var agendamentos []Agendamento
	var total int64

	err := r.db.WithContext(ctx).
		Model(&Agendamento{}).
		Where("status = ?", status).
		Count(&total).
		Offset(offset).
		Limit(limit).
		Order("data_hora DESC").
		Preload("Cliente").
		Preload("Servico").
		Find(&agendamentos).Error

	return agendamentos, total, err
}

// Update atualiza um agendamento
func (r *repo) Update(ctx context.Context, a *Agendamento) error {
	return r.db.WithContext(ctx).Save(a).Error
}

// Delete deleta um agendamento
func (r *repo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Agendamento{}, id).Error
}

// CheckConflict verifica se já existe um agendamento no mesmo horário
func (r *repo) CheckConflict(ctx context.Context, servicoID uint, dataHora time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Agendamento{}).
		Where("servico_id = ? AND data_hora = ? AND status != ?", servicoID, dataHora, Cancelado).
		Count(&count).Error
	return count > 0, err
}
