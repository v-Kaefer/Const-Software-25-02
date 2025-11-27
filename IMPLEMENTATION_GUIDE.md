# 🛠️ GUIA DE IMPLEMENTAÇÃO - O QUE FAZER AGORA

---

## 1️⃣ ESCOLHER O DOMÍNIO DE NEGÓCIO

A decisão **mais importante** é escolher qual domínio implementar. Aqui estão 3 opções bem estruturadas:

### Opção A: Sistema de Agendamento de Serviços ⭐ RECOMENDADO

**Por quê?** Já tem migrações de "alunos" que podem ser reaproveitadas como "clientes"

**Entidades (3+):**
1. **Cliente** (Usuario existente + campos adicionais)
2. **Serviço** (tipo de serviço, duração, preço)
3. **Agendamento** (cliente + serviço + data/hora)
4. **Avaliação** (opcional - cliente avalia serviço)

**Fluxos de Negócio:**
1. **Agendar Serviço**
   - POST /api/v1/agendamentos (criar) - role: user/reviewer
   - Validar: cliente existe, data no futuro, não há conflito
   - Gerar aviso para admin

2. **Aprovar Agendamento**
   - PATCH /api/v1/agendamentos/{id} (status: aprovado) - role: admin
   - Validar: status = pendente
   - Notificar cliente

3. **Cancelar Agendamento**
   - DELETE /api/v1/agendamentos/{id} - role: admin/owner
   - Validar: status != cancelado e != concluído

---

### Opção B: E-commerce Simplificado

**Entidades:**
1. **Produto**
2. **Pedido**
3. **ItemPedido**
4. **Carrinho** (opcional)

**Fluxos:**
1. Criar pedido → Processar pagamento → Enviar
2. Cancelar pedido

---

### Opção C: Gestão de Matrículas em Cursos

**Usa a migration existente!**

**Entidades:**
1. **Aluno** (já existe)
2. **Curso**
3. **Matricula** (já existe na migration)

**Fluxos:**
1. Criar matrícula → Aprovar → Ativar
2. Cancelar matrícula

---

## 2️⃣ ESTRUTURA RECOMENDADA (Exemplo: Agendamentos)

### Criar a estrutura de diretórios:

```
pkg/
├── user/          (já existe)
├── servico/       (NOVO)
│   ├── model.go
│   ├── repo.go
│   ├── service.go
│   ├── repo_test.go
│   └── service_test.go
└── agendamento/   (NOVO)
    ├── model.go
    ├── repo.go
    ├── service.go
    ├── repo_test.go
    └── service_test.go
```

---

## 3️⃣ IMPLEMENTAÇÃO PASSO A PASSO

### Step 1: Criar Models

**`pkg/servico/model.go`:**
```go
package servico

import "time"

type Servico struct {
    ID          uint
    Nome        string    // ex: "Corte de Cabelo"
    Descricao   string
    Duracao     int       // minutos
    Preco       float64
    Ativo       bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (Servico) TableName() string {
    return "servicos"
}
```

**`pkg/agendamento/model.go`:**
```go
package agendamento

import "time"

type Status string

const (
    Pendente  Status = "pendente"
    Aprovado  Status = "aprovado"
    Concluido Status = "concluido"
    Cancelado Status = "cancelado"
)

type Agendamento struct {
    ID        uint
    ClienteID uint
    ServicoID uint
    DataHora  time.Time
    Status    Status
    Notas     string
    CreatedAt time.Time
    UpdatedAt time.Time
    
    // Relacionamentos (GORM)
    Cliente   *User     `gorm:"foreignKey:ClienteID"`
    Servico   *Servico  `gorm:"foreignKey:ServicoID"`
}

func (Agendamento) TableName() string {
    return "agendamentos"
}
```

---

### Step 2: Criar Migrations SQL

**`migrations/0002_create_servicos.sql`:**
```sql
CREATE TABLE IF NOT EXISTS servicos (
  id            SERIAL PRIMARY KEY,
  nome          VARCHAR(255) NOT NULL,
  descricao     TEXT,
  duracao       INTEGER NOT NULL,
  preco         DECIMAL(10, 2) NOT NULL,
  ativo         BOOLEAN DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_servicos_ativo ON servicos(ativo);
```

**`migrations/0003_create_agendamentos.sql`:**
```sql
CREATE TABLE IF NOT EXISTS agendamentos (
  id            SERIAL PRIMARY KEY,
  cliente_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  servico_id    INTEGER NOT NULL REFERENCES servicos(id) ON DELETE RESTRICT,
  data_hora     TIMESTAMPTZ NOT NULL,
  status        VARCHAR(20) NOT NULL DEFAULT 'pendente',
  notas         TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agendamentos_cliente_id ON agendamentos(cliente_id);
CREATE INDEX idx_agendamentos_servico_id ON agendamentos(servico_id);
CREATE INDEX idx_agendamentos_data_hora ON agendamentos(data_hora);
CREATE INDEX idx_agendamentos_status ON agendamentos(status);

-- Prevenir duplicação de slots de horário
CREATE UNIQUE INDEX idx_agendamentos_slot 
ON agendamentos(servico_id, data_hora) 
WHERE status != 'cancelado';
```

---

### Step 3: Criar Repositórios

**`pkg/agendamento/repo.go`:**
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
    ListPendentes(ctx context.Context, offset, limit int) ([]Agendamento, int64, error)
    Update(ctx context.Context, a *Agendamento) error
    Delete(ctx context.Context, id uint) error
    CheckConflict(ctx context.Context, servicoID uint, dataHora time.Time) (bool, error)
}

type repo struct{ db *gorm.DB }

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
        Preload("Cliente").
        Preload("Servico").
        Find(&agendamentos).Error
    
    return agendamentos, total, err
}

// ... mais métodos
```

---

### Step 4: Criar Serviços

**`pkg/agendamento/service.go`:**
```go
package agendamento

import (
    "context"
    "fmt"
    "time"
    "gorm.io/gorm"
)

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo}
}

func (s *Service) AgendarServico(ctx context.Context, clienteID, servicoID uint, dataHora time.Time) (*Agendamento, error) {
    // Validação 1: Data no futuro
    if dataHora.Before(time.Now()) {
        return nil, fmt.Errorf("data deve ser no futuro")
    }

    // Validação 2: Verificar conflito
    hasConflict, err := s.repo.CheckConflict(ctx, servicoID, dataHora)
    if err != nil {
        return nil, err
    }
    if hasConflict {
        return nil, fmt.Errorf("horário já ocupado")
    }

    // Criar agendamento
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

func (s *Service) AprovarAgendamento(ctx context.Context, id uint) (*Agendamento, error) {
    agendamento, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if agendamento.Status != Pendente {
        return nil, fmt.Errorf("apenas agendamentos pendentes podem ser aprovados")
    }

    agendamento.Status = Aprovado
    return agendamento, s.repo.Update(ctx, agendamento)
}
```

---

### Step 5: Criar Handlers HTTP

**`internal/http/agendamento_handler.go`:**
```go
package http

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"
    "github.com/v-Kaefer/Const-Software-25-02/internal/auth"
    "github.com/v-Kaefer/Const-Software-25-02/pkg/agendamento"
)

func (r *Router) registerAgendamentoRoutes() {
    // POST /api/v1/agendamentos
    r.mux.Handle("POST /api/v1/agendamentos",
        r.authMiddleware.Authenticate(
            http.HandlerFunc(r.handleCreateAgendamento),
        ))

    // GET /api/v1/agendamentos (admin lista tudo, user lista os seus)
    r.mux.Handle("GET /api/v1/agendamentos",
        r.authMiddleware.Authenticate(
            http.HandlerFunc(r.handleListAgendamentos),
        ))

    // GET /api/v1/agendamentos/{id}
    r.mux.Handle("GET /api/v1/agendamentos/{id}",
        r.authMiddleware.Authenticate(
            http.HandlerFunc(r.handleGetAgendamento),
        ))

    // PATCH /api/v1/agendamentos/{id} (mudar status)
    r.mux.Handle("PATCH /api/v1/agendamentos/{id}",
        r.authMiddleware.Authenticate(
            r.authMiddleware.RequireRole(auth.RoleAdmin)(
                http.HandlerFunc(r.handleUpdateAgendamento),
            )))

    // DELETE /api/v1/agendamentos/{id}
    r.mux.Handle("DELETE /api/v1/agendamentos/{id}",
        r.authMiddleware.Authenticate(
            http.HandlerFunc(r.handleDeleteAgendamento),
        ))
}

func (r *Router) handleCreateAgendamento(w http.ResponseWriter, req *http.Request) {
    type CreateAgendamentoReq struct {
        ServicoID uint   `json:"servico_id"`
        DataHora  string `json:"data_hora"` // "2025-12-01T14:00:00Z"
        Notas     string `json:"notas,omitempty"`
    }

    var body CreateAgendamentoReq
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    // Parse data
    dataHora, err := time.Parse(time.RFC3339, body.DataHora)
    if err != nil {
        http.Error(w, "invalid data_hora format", http.StatusBadRequest)
        return
    }

    // Get user ID from context
    username, ok := auth.GetUserFromContext(req.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    // TODO: Implementar GetClienteByEmail ou usar User existente
    clienteID := uint(1) // placeholder

    a, err := r.agendamentoSvc.AgendarServico(req.Context(), clienteID, body.ServicoID, dataHora)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(a)
}

// ... outros handlers
```

---

### Step 6: Integrar no Router

**Atualizar `internal/http/handler.go`:**
```go
type Router struct {
    userSvc         *user.Service
    agendamentoSvc  *agendamento.Service    // NOVO
    servicoSvc      *servico.Service        // NOVO
    authMiddleware  *auth.Middleware
    mux             *http.ServeMux
}

func NewRouter(
    userSvc *user.Service,
    agendamentoSvc *agendamento.Service,
    servicoSvc *servico.Service,
    authMiddleware *auth.Middleware,
) *Router {
    r := &Router{
        userSvc:        userSvc,
        agendamentoSvc: agendamentoSvc,
        servicoSvc:     servicoSvc,
        authMiddleware: authMiddleware,
        mux:            http.NewServeMux(),
    }
    r.routes()
    return r
}

func (r *Router) routes() {
    // Versionar todas as rotas
    r.mux.Handle("POST /api/v1/users", ...)
    r.mux.Handle("GET /api/v1/users", ...)
    
    // Registrar rotas de agendamento
    r.registerAgendamentoRoutes()
    r.registerServicoRoutes()
}
```

---

## 4️⃣ ATUALIZAR OpenAPI

**Adicionar ao `openapi/openapi.yaml`:**

```yaml
components:
  schemas:
    Servico:
      type: object
      properties:
        id:
          type: integer
        nome:
          type: string
          example: "Corte de Cabelo"
        duracao:
          type: integer
          description: "Duração em minutos"
        preco:
          type: number
          format: float
      required:
        - nome
        - duracao
        - preco

    Agendamento:
      type: object
      properties:
        id:
          type: integer
        cliente_id:
          type: integer
        servico_id:
          type: integer
        data_hora:
          type: string
          format: date-time
        status:
          type: string
          enum: [pendente, aprovado, concluido, cancelado]
        notas:
          type: string

paths:
  /api/v1/agendamentos:
    post:
      summary: "Agendar serviço"
      tags: [Agendamentos]
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                servico_id:
                  type: integer
                data_hora:
                  type: string
                  format: date-time
                  example: "2025-12-01T14:00:00Z"
                notas:
                  type: string
      responses:
        '201':
          description: "Agendamento criado"
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Agendamento'
```

---

## 5️⃣ IMPLEMENTAR PAGINAÇÃO

Criar utilitário genérico:

**`internal/http/pagination.go`:**
```go
package http

import (
    "net/http"
    "strconv"
)

type PaginationParams struct {
    Page  int
    Limit int
}

func ParsePaginationParams(r *http.Request) PaginationParams {
    page := 1
    if p := r.URL.Query().Get("page"); p != "" {
        if pageNum, err := strconv.Atoi(p); err == nil && pageNum > 0 {
            page = pageNum
        }
    }

    limit := 10
    if l := r.URL.Query().Get("limit"); l != "" {
        if limitNum, err := strconv.Atoi(l); err == nil && limitNum > 0 && limitNum <= 100 {
            limit = limitNum
        }
    }

    return PaginationParams{page, limit}
}

type PaginatedResponse struct {
    Data  interface{} `json:"data"`
    Total int64       `json:"total"`
    Page  int         `json:"page"`
    Limit int         `json:"limit"`
}
```

---

## 6️⃣ TESTES

**`pkg/agendamento/service_test.go`:**
```go
package agendamento

import (
    "context"
    "testing"
    "time"
)

func TestAgendarServico(t *testing.T) {
    // Mock repository
    repo := &mockRepo{}
    svc := NewService(repo)

    ctx := context.Background()
    dataFutura := time.Now().Add(24 * time.Hour)

    agendamento, err := svc.AgendarServico(ctx, 1, 1, dataFutura)
    if err != nil {
        t.Fatalf("erro inesperado: %v", err)
    }

    if agendamento.Status != Pendente {
        t.Errorf("status esperado Pendente, got %v", agendamento.Status)
    }
}

func TestAgendarServico_DataPassada(t *testing.T) {
    repo := &mockRepo{}
    svc := NewService(repo)

    ctx := context.Background()
    dataPassada := time.Now().Add(-24 * time.Hour)

    _, err := svc.AgendarServico(ctx, 1, 1, dataPassada)
    if err == nil {
        t.Error("esperava erro para data passada")
    }
}
```

---

## ✅ CHECKLIST DE IMPLEMENTAÇÃO

```
[ ] Decidir domínio (Agendamento, E-commerce, etc)
[ ] Criar models (Servico, Agendamento)
[ ] Criar migrations SQL
[ ] Criar repositórios (repo.go)
[ ] Criar serviços (service.go)
[ ] Criar handlers HTTP
[ ] Registrar rotas em Router
[ ] Adicionar ao main.go
[ ] Atualizar OpenAPI
[ ] Implementar paginação
[ ] Adicionar testes
[ ] Testar com curl/Postman
[ ] Atualizar README
```

---

## 🔗 REFERÊNCIAS

- GORM Relations: https://gorm.io/docs/associations.html
- OpenAPI 3.1: https://spec.openapis.org/oas/v3.1.0
- Go Testing: https://golang.org/pkg/testing/

---

**Próximo passo:** Executar os passos acima no seu domínio escolhido!
