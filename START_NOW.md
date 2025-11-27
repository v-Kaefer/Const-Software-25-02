# 🚀 COMEÇAR AGORA - Exemplo Prático (30 min)

## O Que Vamos Fazer (em 30 minutos)

Vamos criar o **mínimo viável** para ter 2 novas entidades funcionando:

1. **Serviço** (tipo de serviço)
2. **Agendamento** (reserva de um serviço)

---

## Passo 1: Criar Arquivo `pkg/servico/model.go`

**Crie o arquivo:** `pkg/servico/model.go`

```go
package servico

import "time"

// Servico representa um serviço oferecido pelo sistema
// Ex: Corte de cabelo, Massagem, Consulta médica, etc.
type Servico struct {
	ID          uint      `gorm:"primaryKey"`
	Nome        string    `gorm:"size:255;not null"`           // Ex: "Corte de Cabelo"
	Descricao   string    `gorm:"type:text"`                   // Descrição detalhada
	Duracao     int       `gorm:"not null"`                    // Minutos
	Preco       float64   `gorm:"type:decimal(10,2);not null"` // R$ 50.00
	Ativo       bool      `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName especifica o nome da tabela
func (Servico) TableName() string {
	return "servicos"
}
```

---

## Passo 2: Criar Arquivo `pkg/agendamento/model.go`

**Crie o arquivo:** `pkg/agendamento/model.go`

```go
package agendamento

import (
	"time"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/servico"
)

// Status tipos de status para um agendamento
type Status string

const (
	Pendente  Status = "pendente"
	Aprovado  Status = "aprovado"
	Concluido Status = "concluido"
	Cancelado Status = "cancelado"
)

// Agendamento representa uma reserva de um cliente para um serviço
type Agendamento struct {
	ID        uint      `gorm:"primaryKey"`
	ClienteID uint      `gorm:"not null"`           // FK para User
	ServicoID uint      `gorm:"not null"`           // FK para Servico
	DataHora  time.Time `gorm:"not null;index"`     // Quando vai acontecer
	Status    Status    `gorm:"type:varchar(20);default:'pendente'"` // Estado
	Notas     string    `gorm:"type:text"`          // Notas extras
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relacionamentos GORM (opcional, mas útil)
	Cliente *user.User     `gorm:"foreignKey:ClienteID;constraint:OnDelete:CASCADE"`
	Servico *servico.Servico `gorm:"foreignKey:ServicoID;constraint:OnDelete:RESTRICT"`
}

func (Agendamento) TableName() string {
	return "agendamentos"
}
```

---

## Passo 3: Criar Migrações SQL

**Crie o arquivo:** `migrations/0002_create_servicos.sql`

```sql
CREATE TABLE IF NOT EXISTS servicos (
  id          SERIAL PRIMARY KEY,
  nome        VARCHAR(255) NOT NULL,
  descricao   TEXT,
  duracao     INTEGER NOT NULL,
  preco       DECIMAL(10, 2) NOT NULL,
  ativo       BOOLEAN DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_servicos_ativo ON servicos(ativo);
```

**Crie o arquivo:** `migrations/0003_create_agendamentos.sql`

```sql
CREATE TABLE IF NOT EXISTS agendamentos (
  id          SERIAL PRIMARY KEY,
  cliente_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  servico_id  INTEGER NOT NULL REFERENCES servicos(id) ON DELETE RESTRICT,
  data_hora   TIMESTAMPTZ NOT NULL,
  status      VARCHAR(20) NOT NULL DEFAULT 'pendente',
  notas       TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agendamentos_cliente_id ON agendamentos(cliente_id);
CREATE INDEX idx_agendamentos_servico_id ON agendamentos(servico_id);
CREATE INDEX idx_agendamentos_data_hora ON agendamentos(data_hora);
CREATE INDEX idx_agendamentos_status ON agendamentos(status);
```

---

## Passo 4: Criar Repositórios Simples

**Crie o arquivo:** `pkg/servico/repo.go`

```go
package servico

import (
	"context"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, s *Servico) error
	GetByID(ctx context.Context, id uint) (*Servico, error)
	List(ctx context.Context, offset, limit int) ([]Servico, int64, error)
	Update(ctx context.Context, s *Servico) error
	Delete(ctx context.Context, id uint) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) Repository {
	return &repo{db}
}

func (r *repo) Create(ctx context.Context, s *Servico) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *repo) GetByID(ctx context.Context, id uint) (*Servico, error) {
	var s Servico
	err := r.db.WithContext(ctx).First(&s, id).Error
	return &s, err
}

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

func (r *repo) Update(ctx context.Context, s *Servico) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *repo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Servico{}, id).Error
}
```

**Crie o arquivo:** `pkg/agendamento/repo.go`

```go
package agendamento

import (
	"context"
	"time"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, a *Agendamento) error
	GetByID(ctx context.Context, id uint) (*Agendamento, error)
	ListByCliente(ctx context.Context, clienteID uint, offset, limit int) ([]Agendamento, int64, error)
	ListAll(ctx context.Context, offset, limit int) ([]Agendamento, int64, error)
	Update(ctx context.Context, a *Agendamento) error
	Delete(ctx context.Context, id uint) error
	CheckConflict(ctx context.Context, servicoID uint, dataHora time.Time) (bool, error)
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) Repository {
	return &repo{db}
}

func (r *repo) Create(ctx context.Context, a *Agendamento) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repo) GetByID(ctx context.Context, id uint) (*Agendamento, error) {
	var a Agendamento
	err := r.db.WithContext(ctx).
		Preload("Cliente").
		Preload("Servico").
		First(&a, id).Error
	return &a, err
}

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

func (r *repo) Update(ctx context.Context, a *Agendamento) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *repo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Agendamento{}, id).Error
}

func (r *repo) CheckConflict(ctx context.Context, servicoID uint, dataHora time.Time) (bool, error) {
	var count int64
	// Buscar agendamentos NÃO cancelados no mesmo horário
	err := r.db.WithContext(ctx).
		Model(&Agendamento{}).
		Where("servico_id = ? AND data_hora = ? AND status != ?", servicoID, dataHora, Cancelado).
		Count(&count).Error
	return count > 0, err
}
```

---

## Passo 5: Criar Serviços (Lógica de Negócio)

**Crie o arquivo:** `pkg/servico/service.go`

```go
package servico

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

func (s *Service) Create(ctx context.Context, nome, descricao string, duracao int, preco float64) (*Servico, error) {
	if nome == "" {
		return nil, fmt.Errorf("nome é obrigatório")
	}
	if duracao <= 0 {
		return nil, fmt.Errorf("duração deve ser maior que 0")
	}
	if preco <= 0 {
		return nil, fmt.Errorf("preço deve ser maior que 0")
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

func (s *Service) GetByID(ctx context.Context, id uint) (*Servico, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]Servico, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.List(ctx, offset, limit)
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
```

**Crie o arquivo:** `pkg/agendamento/service.go`

```go
package agendamento

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

// AgendarServico cria um novo agendamento (fluxo 1)
func (s *Service) AgendarServico(ctx context.Context, clienteID, servicoID uint, dataHora time.Time) (*Agendamento, error) {
	// Validação 1: Data deve estar no futuro
	if dataHora.Before(time.Now()) {
		return nil, fmt.Errorf("data deve estar no futuro")
	}

	// Validação 2: Verificar conflito de horário
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

// AprovarAgendamento muda status para aprovado (fluxo 2 - admin)
func (s *Service) AprovarAgendamento(ctx context.Context, id uint) (*Agendamento, error) {
	agendamento, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if agendamento.Status != Pendente {
		return nil, fmt.Errorf("apenas agendamentos pendentes podem ser aprovados")
	}

	agendamento.Status = Aprovado
	if err := s.repo.Update(ctx, agendamento); err != nil {
		return nil, err
	}

	return agendamento, nil
}

// CancelarAgendamento cancela um agendamento (fluxo 3)
func (s *Service) CancelarAgendamento(ctx context.Context, id uint) (*Agendamento, error) {
	agendamento, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if agendamento.Status == Concluido {
		return nil, fmt.Errorf("não pode cancelar agendamento concluído")
	}

	if agendamento.Status == Cancelado {
		return nil, fmt.Errorf("agendamento já foi cancelado")
	}

	agendamento.Status = Cancelado
	if err := s.repo.Update(ctx, agendamento); err != nil {
		return nil, err
	}

	return agendamento, nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*Agendamento, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByCliente(ctx context.Context, clienteID uint, offset, limit int) ([]Agendamento, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.ListByCliente(ctx, clienteID, offset, limit)
}

func (s *Service) ListAll(ctx context.Context, offset, limit int) ([]Agendamento, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.ListAll(ctx, offset, limit)
}
```

---

## Próximos Passos

### Hoje (30 min completado ✅)
- [x] Criar models
- [x] Criar migrations
- [x] Criar repos
- [x] Criar services

### Amanhã (Handlers + Router)
- [ ] Criar handlers em `internal/http/servico_handler.go`
- [ ] Criar handlers em `internal/http/agendamento_handler.go`
- [ ] Registrar rotas em `internal/http/handler.go`
- [ ] Testar com Docker

### Dia 3 (Testes + Docs)
- [ ] Criar testes unitários
- [ ] Atualizar OpenAPI
- [ ] Criar exemplos curl
- [ ] Atualizar README

---

## Verificar

Depois que criar os arquivos acima, teste se compila:

```bash
cd c:\Users\Administrador\Documents\cs\Const-Software-25-02
go build ./cmd/api
```

Se der erro sobre imports, crie os diretórios:
```
mkdir -p pkg/servico
mkdir -p pkg/agendamento
```

---

**Parabéns! Você tem os modelos e a lógica de negócio pronta!** 🎉

Próximo: criar os handlers HTTP para expor essa lógica como API.
