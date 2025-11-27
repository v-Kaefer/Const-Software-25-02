package agendamento

import (
	"context"
	"fmt"
	"time"
)

// Service encapsula a lógica de negócio para Agendamento
type Service struct {
	repo Repository
}

// NewService cria uma nova instância de service
func NewService(repo Repository) *Service {
	return &Service{repo}
}

// Agendar cria um novo agendamento (FLUXO 1)
func (s *Service) Agendar(ctx context.Context, clienteID, servicoID uint, dataHora time.Time) (*Agendamento, error) {
	// Validação 1: IDs válidos
	if clienteID == 0 {
		return nil, fmt.Errorf("clienteID inválido")
	}
	if servicoID == 0 {
		return nil, fmt.Errorf("servicoID inválido")
	}

	// Validação 2: Data deve estar no futuro
	if dataHora.Before(time.Now()) {
		return nil, fmt.Errorf("data deve estar no futuro")
	}

	// Validação 3: Verificar conflito de horário
	hasConflict, err := s.repo.CheckConflict(ctx, servicoID, dataHora)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar conflito: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("já existe agendamento para este serviço neste horário")
	}

	agendamento := &Agendamento{
		ClienteID: clienteID,
		ServicoID: servicoID,
		DataHora:  dataHora,
		Status:    Pendente,
	}

	if err := s.repo.Create(ctx, agendamento); err != nil {
		return nil, err
	}

	return agendamento, nil
}

// Aprovar aprova um agendamento (FLUXO 2 - Admin only)
func (s *Service) Aprovar(ctx context.Context, id uint) (*Agendamento, error) {
	agendamento, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validação: apenas agendamentos pendentes podem ser aprovados
	if agendamento.Status != Pendente {
		return nil, fmt.Errorf("apenas agendamentos pendentes podem ser aprovados (status atual: %s)", agendamento.Status)
	}

	agendamento.Status = Aprovado
	if err := s.repo.Update(ctx, agendamento); err != nil {
		return nil, err
	}

	return agendamento, nil
}

// Concluir marca um agendamento como concluído
func (s *Service) Concluir(ctx context.Context, id uint) (*Agendamento, error) {
	agendamento, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validação: apenas agendamentos aprovados podem ser concluídos
	if agendamento.Status != Aprovado {
		return nil, fmt.Errorf("apenas agendamentos aprovados podem ser concluídos")
	}

	agendamento.Status = Concluido
	if err := s.repo.Update(ctx, agendamento); err != nil {
		return nil, err
	}

	return agendamento, nil
}

// Cancelar cancela um agendamento (FLUXO 3)
func (s *Service) Cancelar(ctx context.Context, id uint) (*Agendamento, error) {
	agendamento, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validação: não pode cancelar se já foi concluído
	if agendamento.Status == Concluido {
		return nil, fmt.Errorf("não pode cancelar agendamento concluído")
	}

	// Validação: não pode cancelar se já foi cancelado
	if agendamento.Status == Cancelado {
		return nil, fmt.Errorf("agendamento já foi cancelado")
	}

	agendamento.Status = Cancelado
	if err := s.repo.Update(ctx, agendamento); err != nil {
		return nil, err
	}

	return agendamento, nil
}

// GetByID busca um agendamento por ID
func (s *Service) GetByID(ctx context.Context, id uint) (*Agendamento, error) {
	if id == 0 {
		return nil, fmt.Errorf("id inválido")
	}
	return s.repo.GetByID(ctx, id)
}

// ListByCliente lista agendamentos de um cliente
func (s *Service) ListByCliente(ctx context.Context, clienteID uint, offset, limit int) ([]Agendamento, int64, error) {
	if clienteID == 0 {
		return nil, 0, fmt.Errorf("clienteID inválido")
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.ListByCliente(ctx, clienteID, offset, limit)
}

// ListAll lista todos os agendamentos
func (s *Service) ListAll(ctx context.Context, offset, limit int) ([]Agendamento, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.ListAll(ctx, offset, limit)
}

// ListByStatus lista agendamentos por status
func (s *Service) ListByStatus(ctx context.Context, status Status, offset, limit int) ([]Agendamento, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.ListByStatus(ctx, status, offset, limit)
}
