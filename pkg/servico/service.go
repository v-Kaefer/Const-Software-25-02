package servico

import (
	"context"
	"fmt"
)

// Service encapsula a lógica de negócio para Servico
type Service struct {
	repo Repository
}

// NewService cria uma nova instância de service
func NewService(repo Repository) *Service {
	return &Service{repo}
}

// Create cria um novo serviço com validações
func (s *Service) Create(ctx context.Context, nome, descricao string, duracao int, preco float64) (*Servico, error) {
	// Validação: nome é obrigatório
	if nome == "" {
		return nil, fmt.Errorf("nome é obrigatório")
	}

	// Validação: duração deve ser maior que 0
	if duracao <= 0 {
		return nil, fmt.Errorf("duração deve ser maior que 0 minutos")
	}

	// Validação: preço deve ser maior que 0
	if preco <= 0 {
		return nil, fmt.Errorf("preço deve ser maior que 0")
	}

	// Validação: nome único
	_, err := s.repo.GetByNome(ctx, nome)
	if err == nil {
		return nil, fmt.Errorf("já existe um serviço com este nome")
	}

	servico := &Servico{
		Nome:      nome,
		Descricao: descricao,
		Duracao:   duracao,
		Preco:     preco,
		Ativo:     true,
	}

	if err := s.repo.Create(ctx, servico); err != nil {
		return nil, err
	}

	return servico, nil
}

// GetByID busca um serviço por ID
func (s *Service) GetByID(ctx context.Context, id uint) (*Servico, error) {
	if id == 0 {
		return nil, fmt.Errorf("id inválido")
	}
	return s.repo.GetByID(ctx, id)
}

// List lista todos os serviços com paginação
func (s *Service) List(ctx context.Context, offset, limit int) ([]Servico, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.List(ctx, offset, limit)
}

// ListAtivos lista apenas serviços ativos
func (s *Service) ListAtivos(ctx context.Context, offset, limit int) ([]Servico, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.ListAtivos(ctx, offset, limit)
}

// Update atualiza um serviço
func (s *Service) Update(ctx context.Context, id uint, nome, descricao string, duracao int, preco float64) (*Servico, error) {
	// Validações
	if id == 0 {
		return nil, fmt.Errorf("id inválido")
	}
	if nome == "" {
		return nil, fmt.Errorf("nome é obrigatório")
	}
	if duracao <= 0 {
		return nil, fmt.Errorf("duração deve ser maior que 0")
	}
	if preco <= 0 {
		return nil, fmt.Errorf("preço deve ser maior que 0")
	}

	servico, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	servico.Nome = nome
	servico.Descricao = descricao
	servico.Duracao = duracao
	servico.Preco = preco

	if err := s.repo.Update(ctx, servico); err != nil {
		return nil, err
	}

	return servico, nil
}

// Desativar desativa um serviço
func (s *Service) Desativar(ctx context.Context, id uint) (*Servico, error) {
	servico, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	servico.Ativo = false
	if err := s.repo.Update(ctx, servico); err != nil {
		return nil, err
	}

	return servico, nil
}

// Ativar ativa um serviço
func (s *Service) Ativar(ctx context.Context, id uint) (*Servico, error) {
	servico, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	servico.Ativo = true
	if err := s.repo.Update(ctx, servico); err != nil {
		return nil, err
	}

	return servico, nil
}

// Delete deleta um serviço
func (s *Service) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("id inválido")
	}
	return s.repo.Delete(ctx, id)
}
