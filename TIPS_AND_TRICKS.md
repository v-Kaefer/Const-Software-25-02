# 🛠️ DICAS PRÁTICAS PARA DESENVOLVIMENTO

Aqui estão dicas e truques que vão acelerar sua implementação.

---

## 🚀 Startup Rápido (Clone do User para Novas Entidades)

Ao invés de criar tudo do zero, use o `User` como template:

### Passo 1: Copiar estrutura existente

```bash
# Copiar user como base para servico
cp -r pkg/user pkg/servico

# Copiar handlers
cp internal/http/handler.go internal/http/servico_handler.go
```

### Passo 2: Renomear structs

Em `pkg/servico/model.go`:
```go
// De:
type User struct {
    ID    uint
    Email string
    Name  string
}

// Para:
type Servico struct {
    ID    uint
    Nome  string
    Preco float64
    Duracao int
}
```

### Passo 3: Atualizar métodos

Em `pkg/servico/service.go`:
```go
// Renomear
Register → Create
GetByEmail → GetByID (ou GetByNome)
List → List (reutilizar)
```

---

## 💻 Testes Rápidos sem Docker

Se Docker não estiver funcionando, use SQLite local:

### Arquivo `.env.dev`:
```bash
APP_ENV=development
DB_HOST=
DB_USER=
DB_PASS=
DB_NAME=app.db      # Arquivo SQLite local
DB_SSLMODE=disable
```

### Rodar com SQLite:
```bash
export $(cat .env.dev | xargs)
go run cmd/api/main.go
```

Banco fica em `app.db` - fácil de deletar e recrear!

---

## 🧪 Testes com curl (sem Postman)

### Script de teste simples:

```bash
#!/bin/bash

# URL base
BASE_URL="http://localhost:8080"

# Criar token (mock)
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 1. Criar serviço
curl -X POST $BASE_URL/api/v1/servicos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "Corte",
    "duracao": 30,
    "preco": 50
  }' | jq .

# 2. Listar serviços
curl -X GET "$BASE_URL/api/v1/servicos?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 3. Agendar
curl -X POST $BASE_URL/api/v1/agendamentos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "servico_id": 1,
    "data_hora": "2025-12-15T10:00:00Z"
  }' | jq .
```

---

## 🧠 Padrão Repository Pattern

Todos os repositories seguem o mesmo padrão:

```go
type Repository interface {
    Create(ctx context.Context, model *Model) error
    GetByID(ctx context.Context, id uint) (*Model, error)
    List(ctx context.Context, offset, limit int) ([]Model, int64, error)
    Update(ctx context.Context, model *Model) error
    Delete(ctx context.Context, id uint) error
}

type repo struct {
    db *gorm.DB
}

// Implementar cada método...
```

**Reutilize este padrão para todas as entidades!**

---

## 📝 Template de Service

Todos os services têm:

```go
type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo}
}

func (s *Service) SomeBusinessLogic(ctx context.Context) error {
    // 1. Validações
    if !isValid(data) {
        return fmt.Errorf("validação falhou")
    }

    // 2. Verificar conflitos
    if exists, err := s.repo.CheckConflict(ctx); exists {
        return fmt.Errorf("conflito detectado")
    }

    // 3. Executar ação
    return s.repo.Create(ctx, model)
}
```

**Use como base para todas as services!**

---

## 🔐 Proteger Rotas (Autorização)

### Padrão básico:

```go
// Admin only
r.mux.Handle("DELETE /api/v1/servicos/{id}",
    authMiddleware.Authenticate(
        authMiddleware.RequireRole(auth.RoleAdmin)(
            http.HandlerFunc(handler))))

// Próprio usuário ou admin
r.mux.Handle("GET /api/v1/agendamentos/{id}",
    authMiddleware.Authenticate(
        http.HandlerFunc(handler)))  // Verificar ownership no handler
```

### No handler:

```go
func (r *Router) handleGetAgendamento(w http.ResponseWriter, req *http.Request) {
    // Verificar ownership
    roles, _ := auth.GetRolesFromContext(req.Context())
    username, _ := auth.GetUserFromContext(req.Context())
    
    isAdmin := hasRole(roles, "admin")
    isOwner := checkIfOwner(username, id)
    
    if !isAdmin && !isOwner {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }
    
    // Continuar...
}
```

---

## 📊 Estrutura de Response com Paginação

Sempre usar a mesma estrutura:

```go
type ListResponse struct {
    Data  interface{} `json:"data"`
    Total int64       `json:"total"`
    Page  int         `json:"page"`
    Limit int         `json:"limit"`
    Pages int         `json:"pages"`  // total / limit
}

// No handler:
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(ListResponse{
    Data:  items,
    Total: total,
    Page:  page,
    Limit: limit,
    Pages: int((total + int64(limit) - 1) / int64(limit)),
})
```

---

## ❌ Tratamento de Erros Consistente

### Padrão:

```go
func (r *Router) handleGetItem(w http.ResponseWriter, req *http.Request) {
    id, err := strconv.ParseUint(req.PathValue("id"), 10, 32)
    if err != nil {
        // 400 - Erro do cliente
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

    item, err := r.svc.GetByID(req.Context(), uint(id))
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // 404 - Não encontrado
            http.Error(w, "not found", http.StatusNotFound)
        } else {
            // 500 - Erro do servidor
            http.Error(w, "internal error", http.StatusInternalServerError)
        }
        return
    }

    // 200 - Sucesso
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(item)
}
```

**Sempre retornar o código HTTP correto!**

---

## 🧪 Testes Rápidos (TDD)

### Estrutura básica:

```go
func TestCreateServico_Success(t *testing.T) {
    repo := &mockRepo{}
    svc := NewService(repo)
    
    resultado, err := svc.Create(context.Background(), "Corte", 30, 50.0)
    
    if err != nil {
        t.Fatalf("erro inesperado: %v", err)
    }
    if resultado.Nome != "Corte" {
        t.Errorf("esperava 'Corte', got %q", resultado.Nome)
    }
}

func TestCreateServico_ValidationError(t *testing.T) {
    repo := &mockRepo{}
    svc := NewService(repo)
    
    _, err := svc.Create(context.Background(), "", 30, 50.0)
    
    if err == nil {
        t.Error("esperava erro para nome vazio")
    }
}
```

---

## 📚 Documentar no OpenAPI

Padrão para cada endpoint:

```yaml
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
              required:
                - servico_id
                - data_hora
      responses:
        '201':
          description: "Agendamento criado"
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Agendamento'
        '400':
          description: "Validação falhou"
        '401':
          description: "Não autenticado"
        '403':
          description: "Acesso negado"
        '500':
          description: "Erro interno"
```

---

## 🎯 Checklist de Cada Entidade Nova

```
Servico (ou similar)
├─ [ ] model.go (struct + TableName)
├─ [ ] migrations SQL
├─ [ ] repo.go (interface + implementação)
├─ [ ] service.go (lógica de negócio)
├─ [ ] handlers (create/read/update/delete)
├─ [ ] rotas em handler.go
├─ [ ] OpenAPI schemas
├─ [ ] OpenAPI paths
├─ [ ] testes repo
├─ [ ] testes service
└─ [ ] exemplo curl no README
```

---

## 🚀 Comandos Úteis

### Build & Run
```bash
# Compilar
go build ./cmd/api

# Rodar localmente
go run ./cmd/api/main.go

# Com debug
go run -race ./cmd/api/main.go
```

### Testes
```bash
# Rodar todos os testes
go test ./...

# Com verbose
go test ./... -v

# Com cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Apenas um pacote
go test ./pkg/agendamento/...
```

### Lint & Format
```bash
# Formatar
go fmt ./...

# Vet (problemas comuns)
go vet ./...

# Imports
go mod tidy
```

### Docker
```bash
# Build da imagem
docker build -t userservice .

# Rodar com docker-compose
docker compose up

# Ver logs
docker compose logs -f api

# Limpar
docker compose down
```

---

## 💪 Produtividade

### Atalho 1: Template Go rápido
Crie um arquivo `template_model.go`:
```go
// Template para novo model
// Copie, renomeie, customize

package myentity

import "time"

type MyEntity struct {
    ID        uint
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (MyEntity) TableName() string {
    return "myentities"
}
```

### Atalho 2: Criar múltiplas files
```bash
# Criar estrutura completa para nova entidade
mkdir -p pkg/minha_entidade
touch pkg/minha_entidade/{model,repo,service,repo_test,service_test}.go
```

### Atalho 3: Auto-complete no VS Code
Instale extensão `Go` - permite auto-complete, refactoring, testes.

---

## 📝 Exemplo Real (Colar e Customizar)

### Novo handler simples:
```go
func (r *Router) handleCreateAgendamento(w http.ResponseWriter, req *http.Request) {
    // 1. Parse request
    type req struct {
        ServicoID uint   `json:"servico_id"`
        DataHora  string `json:"data_hora"`
    }
    var body req
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // 2. Parse timestamp
    dataHora, err := time.Parse(time.RFC3339, body.DataHora)
    if err != nil {
        http.Error(w, "invalid data_hora", http.StatusBadRequest)
        return
    }

    // 3. Get user from context
    username, ok := auth.GetUserFromContext(req.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    // 4. Execute business logic
    agendamento, err := r.agendamentoSvc.AgendarServico(
        req.Context(), clienteID, body.ServicoID, dataHora)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // 5. Return success
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(agendamento)
}
```

---

## 🎓 Recursos Úteis

- GORM: https://gorm.io/docs/
- Go Testing: https://golang.org/pkg/testing/
- OpenAPI 3.1: https://spec.openapis.org/
- HTTP Status Codes: https://httpwg.org/specs/rfc9110.html#status.codes
- REST Best Practices: https://restfulapi.net/

---

## ⚡ Otimizações Futuras

Depois que tudo funcionar:

- [ ] Adicionar logging estruturado (slog)
- [ ] Adicionar metrics (prometheus)
- [ ] Adicionar tracing (otel)
- [ ] Adicionar cache (redis)
- [ ] Adicionar rate limiting
- [ ] Adicionar validação de schemas (jsonschema)
- [ ] Adicionar soft deletes
- [ ] Adicionar audit logs

---

**Dica final:** Não tente ser perfeito na primeira tentativa. Faça funcionar, depois refatore!

Boa codificação! 🚀
