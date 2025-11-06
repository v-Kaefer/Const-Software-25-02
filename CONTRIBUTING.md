# User Service – Go + PostgreSQL

**Documentação de Desenvolvimento**

> Serviço completo para o domínio **User**, com especificação **OpenAPI**, infraestrutura Docker, migração SQL, autenticação RBAC com AWS Cognito e CI/CD em GitHub Actions.

## Sumário
1. [Objetivo](#objetivo)
2. [Stack definida](#stack-definida)
3. [Pré-requisitos](#pré-requisitos)
4. [Como rodar com Docker Compose](#como-rodar-com-docker-compose)
5. [Como rodar localmente (sem Docker)](#como-rodar-localmente-sem-docker)
6. [Estrutura do repositório](#estrutura-do-repositório)
7. [Variáveis de ambiente](#variáveis-de-ambiente)
8. [Migrações de banco](#migrações-de-banco)
9. [Testes (go test)](#testes-go-test)
10. [CI no GitHub Actions](#ci-no-github-actions)
11. [Documentação da API (Swagger)](#documentação-da-api-swagger)
12. [Convenções e decisões](#convenções-e-decisões)
13. [Roadmap](#roadmap)
14. [Troubleshooting](#troubleshooting)

## Objetivo
Serviço REST completo para gerenciamento de usuários com autenticação RBAC, persistência em PostgreSQL, infraestrutura AWS e testes automatizados.

## Stack definida
- **Linguagem:** Go 1.22+
- **Framework web:** net/http standard library (Go 1.22+ routing)
- **Banco:** PostgreSQL 16
- **ORM:** GORM
- **Autenticação:** AWS Cognito (JWT tokens)
- **Infra:** Docker + Docker Compose
- **Docs:** OpenAPI 3.1 (Swagger UI via container)
- **Testes:** `go test` (unitários e integração)
- **IaC:** Terraform (AWS)

## Pré-requisitos
- Docker Desktop/Engine e Docker Compose
- Go 1.22+ (para desenvolvimento local fora do container)

## Como rodar com Docker Compose
1. Crie seu `.env` a partir do exemplo:
   ```bash
   cp .env.example .env
    ```
2. Suba **apenas o banco** inicialmente:

   ```bash
   docker compose up -d db
   ```
3. Aplique a migração inicial:

   ```bash
   docker compose exec -T db psql -U app -d app -f /migrations/0001_init.sql
   ```
4. (Opcional nesta sprint) Suba API e Swagger:

   ```bash
   docker compose up -d api swagger
   # API:    http://localhost:8080
   # Swagger http://localhost:8081
   ```
5. Acompanhe logs (útil quando a API estiver implementada):

   ```bash
   docker compose logs -f api
   ```

## Como rodar localmente (sem Docker)

1. Garanta um PostgreSQL local acessível.
2. Configure `DATABASE_URL` (ver [Variáveis de ambiente](#variáveis-de-ambiente)).
3. Aplique a migração:

   ```bash
   psql "$DATABASE_URL" -f migrations/0001_init.sql
   ```
4. Rode a aplicação:

   ```bash
   go run ./cmd/api
   ```

## Estrutura do repositório

```
.
├── cmd/
│   ├── api/                  # Servidor HTTP principal (main.go)
│   └── tests/                # Testes auxiliares
├── internal/
│   ├── auth/                 # Middleware de autenticação RBAC
│   ├── config/               # Leitura de configuração (env vars)
│   ├── db/                   # Conexão e migração GORM
│   └── http/                 # Handlers HTTP e roteamento
├── pkg/
│   └── user/                 # Domínio User (model, service, repository)
├── migrations/
│   └── 0001_init.sql         # Migração SQL inicial
├── infra/                    # Infraestrutura Terraform (AWS)
├── docs/                     # Documentação adicional
├── openapi.yaml              # Especificação OpenAPI 3.1
├── .github/workflows/        # Pipelines CI/CD
├── Dockerfile
├── docker-compose.yaml
├── .env.example
└── README.md
```

## Variáveis de ambiente

| Variável              | Exemplo                                          | Descrição                       |
| --------------------- | ------------------------------------------------ | ------------------------------- |
| `APP_ENV`             | `development`                                    | Ambiente da aplicação           |
| `APP_PORT`            | `8080`                                           | Porta da API                    |
| `DB_HOST`             | `localhost`                                      | Host do PostgreSQL              |
| `DB_PORT`             | `5432`                                           | Porta do PostgreSQL             |
| `DB_USER`             | `app`                                            | Usuário do banco                |
| `DB_PASS`             | `app`                                            | Senha do banco                  |
| `DB_NAME`             | `app`                                            | Nome do banco de dados          |
| `DB_SSLMODE`          | `disable`                                        | Modo SSL do PostgreSQL          |
| `COGNITO_REGION`      | `us-east-1`                                      | Região AWS do Cognito           |
| `COGNITO_USER_POOL_ID`| `us-east-1_xxxxxxxxx`                            | ID do Cognito User Pool         |

> **Dica:** no Compose, `DATABASE_URL` aponta para o host `db` (nome do serviço).

## Migrações de banco

O projeto usa **migração automática com GORM** em ambiente de desenvolvimento. Para produção, as migrações SQL estão em `migrations/`.

### Aplicar migrações manualmente:

```bash
# em ambiente Docker Compose
docker compose exec -T db psql -U app -d app -f /migrations/0001_init.sql

# localmente (sem Docker)
psql "$DATABASE_URL" -f migrations/0001_init.sql
```

### Migração automática (desenvolvimento):

A aplicação executa `AutoMigrate()` automaticamente quando `APP_ENV != production`.

## Testes (go test)

### Unitários

Executar todos os testes com detector de corrida e cobertura:

```bash
go test -v -race -covermode=atomic -coverprofile=coverage.out ./...
```

Relatório resumido de cobertura:

```bash
go tool cover -func=coverage.out
```

Relatório HTML:

```bash
go tool cover -html=coverage.out -o coverage.html
```

### Integração

O projeto inclui testes de integração usando SQLite em memória:

```bash
# executa todos os testes incluindo integração
go test -v ./...
```

Testes E2E estão em `internal/http/handler_e2e__test.go` e testam o fluxo completo da API com banco em memória.

## CI no GitHub Actions

O projeto possui múltiplos workflows em `./.github/workflows/`:

* **ci.yaml**: Verifica **formatação** (`gofmt`), roda `go vet`
* **tests.yaml**: Executa **go test** com `-race` e **cobertura**, publica artefatos
* **build.yaml**: Compila a aplicação e valida build
* **docker-build.yaml**: Constrói e valida a imagem Docker

Todos os workflows executam automaticamente em PRs e pushes para branches principais.

## Documentação da API (Swagger)

* A especificação OpenAPI está em `openapi.yaml` (raiz do projeto)
* A UI do Swagger sobe em `http://localhost:8081` via Docker Compose
* Endpoints implementados:

  * `POST   /users` – cria usuário (requer autenticação admin)
  * `GET    /users?email=...` – obtém usuário por email (público)

### Exemplos de cURL

```bash
# Criar usuário (requer token JWT admin)
curl -i -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <jwt-token>' \
  -d '{"email":"usuario@example.com","name":"João Silva"}'

# Buscar usuário por email (público)
curl -i http://localhost:8080/users?email=usuario@example.com
```

## Convenções e decisões

* **Autenticação:** AWS Cognito com JWT tokens (RBAC)
* **Framework:** net/http standard library (Go 1.22+ routing)
* **ORM:** GORM com auto-migration em desenvolvimento
* **Erros:** HTTP status codes padrão com mensagens de erro
* **Testes:** Mock middleware para testes sem Cognito real
* **Qualidade:** `gofmt`, `go vet` validados em CI

## Roadmap

* [ ] Adicionar mais endpoints (UPDATE, DELETE)
* [ ] Implementar paginação para listagem de usuários
* [ ] Configurar lint adicional (golangci-lint)
* [ ] Adicionar observabilidade (logs estruturados, métricas)
* [ ] Expandir cobertura de testes

## Troubleshooting

* **"connection refused" ao aplicar migração**: aguarde o `healthcheck` do Postgres (Compose) concluir; tente novamente.
* **Swagger vazio**: confirme o volume `./openapi.yaml` no serviço `swagger` e a porta `8081`.
* **Testes falhando**: certifique-se de que todas as dependências foram baixadas com `go mod download`.
