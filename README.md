# Construção de Software 2025/02
Grupo L

https://github.com/v-Kaefer/Const-Software-25-02

# Sprint 0 – Setup de Time, Stack e Projeto

Este pacote entrega um **arquivo fonte OpenAPI** para o domínio `User` (com **POST**, **PATCH** e **PUT**), um **README** passo‑a‑passo, além de arquivos básicos de infraestrutura (Docker/Docker Compose e migração SQL) para iniciar o projeto com Go, Gin e PostgreSQL.

---

## 📦 Estrutura do repositório

```
.
├── cmd/
│   └── api/
│       └── main.go                # (stub futuro) inicialização do servidor Gin
├── internal/
│   ├── config/                    # (stub futuro) leitura de envs/config
│   ├── http/                      # (stub futuro) middlewares e roteamento
│   └── user/                      # (stub futuro) handlers, service e repository
├── migrations/
│   └── 0001_init.sql              # criação da tabela users
├── openapi/
│   └── openapi.yaml               # especificação da API
├── Dockerfile                     # build da API
├── docker-compose.yml             # orquestração (db, api, swagger)
├── .env.example                   # variáveis de ambiente padrão
└── README.md                      # instruções de build/run/test
```

---


# User Service – Go + Gin + PostgreSQL

**Sprint 0 – Setup de Time, Stack e Projeto**

> Serviço base para o domínio **User**, com especificação **OpenAPI**, infraestrutura Docker, migração SQL e CI simples em GitHub Actions.

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
Preparar o ambiente e a estrutura mínima para iniciar o desenvolvimento do domínio `User` com **CRUD** completo definido em OpenAPI.

## Stack definida
- **Linguagem:** Go 1.22+
- **Framework web:** Gin (v1.10+)
- **Banco:** PostgreSQL 16
- **Infra:** Docker + Docker Compose
- **Docs:** OpenAPI 3.0 (Swagger UI via container)
- **Testes:** `go test` (unitários e base para integração)

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

> Útil a partir da Sprint 1, quando o servidor Gin for implementado.

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
├── cmd/api/                 # main.go (servidor Gin) – Sprint 1
├── internal/                # handlers, services, repositories – Sprint 1
├── migrations/              # SQL de migração
│   └── 0001_init.sql
├── openapi/
│   └── openapi.yaml         # especificação da API
├── .github/workflows/
│   └── ci.yaml              # pipeline de testes (go test)
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── README.md
```

## Variáveis de ambiente

| Variável       | Exemplo                                          | Descrição                       |
| -------------- | ------------------------------------------------ | ------------------------------- |
| `APP_PORT`     | `8080`                                           | Porta da API                    |
| `DATABASE_URL` | `postgres://app:app@db:5432/app?sslmode=disable` | String de conexão PostgreSQL    |
| `GIN_MODE`     | `release`                                        | Modo do Gin (`debug`/`release`) |

> **Dica:** no Compose, `DATABASE_URL` aponta para o host `db` (nome do serviço).

## Migrações de banco

Nesta sprint usamos **SQL puro**. Para aplicar:

```bash
# em ambiente Docker Compose
docker compose exec -T db psql -U app -d app -f /migrations/0001_init.sql

# localmente (sem Docker)
psql "$DATABASE_URL" -f migrations/0001_init.sql
```

Futuramente, é possível integrar ferramentas como `golang-migrate` ou `goose`.

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

### Integração (opcional a partir da Sprint 1)

Sugestão: marcar testes de integração com a build tag `integration` e usar um banco efêmero (Docker/Testcontainers).

```bash
# exemplo: executa somente testes marcados com a tag "integration"
go test -v -tags=integration ./...
```

Se usar Docker Compose, garanta que o serviço `db` esteja ativo.

## CI no GitHub Actions

Há um workflow em `./.github/workflows/ci.yaml` que:

* Verifica **formatação** (`gofmt`), roda `go vet`.
* Executa **go test** com `-race` e **cobertura**.
* Publica **artefatos** de cobertura (`coverage.out`, `coverage.txt`).
* Inclui um job (desabilitado por padrão) para **testes de integração** com serviço PostgreSQL.

## Documentação da API (Swagger)

* A UI do Swagger sobe em `http://localhost:8081` e lê `openapi/openapi.yaml`.
* Endpoints principais:

  * `POST   /v1/users` – cria usuário (201 + `Location`)
  * `GET    /v1/users` – lista paginada
  * `GET    /v1/users/{id}` – obtém por ID
  * `PUT    /v1/users/{id}` – substituição completa
  * `PATCH  /v1/users/{id}` – atualização parcial (merge‑patch)
  * `DELETE /v1/users/{id}` – remove

### Exemplos de cURL

```bash
# Criar
curl -i -X POST http://localhost:8080/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ana","email":"ana@example.com","password":"S3nh@Segura!"}'

# Atualização parcial (PATCH)
curl -i -X PATCH http://localhost:8080/v1/users/{id} \
  -H 'Content-Type: application/merge-patch+json' \
  -d '{"name":"Ana M."}'

# Substituição completa (PUT)
curl -i -X PUT http://localhost:8080/v1/users/{id} \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ana Maria","email":"ana.maria@example.com","isActive":true}'
```

## Convenções e decisões

* **PATCH:** `application/merge-patch+json` (RFC 7396).
* **PUT:** requer representação completa do recurso.
* **Paginação:** `page` (>=1) e `pageSize` (<=100), resposta `{ data, meta }`.
* **Erros:** `{ code, message, details? }`.
* **Auth:** `bearerAuth` definido na OpenAPI; não obrigatório nesta sprint.
* **Qualidade:** `gofmt`, `go vet`; considerar `golangci-lint` em sprints futuras.

## Roadmap

* Implementar handlers Gin e camadas service/repository.
* Adicionar testes unitários e de integração (tags e/ou Testcontainers).
* Configurar lint (golangci-lint) e cobertura mínima obrigatória.
* Publicar imagem Docker em registry.

## Troubleshooting

* **"connection refused" ao aplicar migração**: aguarde o `healthcheck` do Postgres (Compose) concluir; tente novamente.
* **Swagger vazio**: confirme o volume `./openapi/openapi.yaml` no serviço `swagger` e a porta `8081`.

```
bash
# Criar
curl -i -X POST http://localhost:8080/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ana","email":"ana@example.com","password":"S3nh@Segura!"}'

# Atualização parcial (PATCH)
curl -i -X PATCH http://localhost:8080/v1/users/{id} \
  -H 'Content-Type: application/merge-patch+json' \
  -d '{"name":"Ana M."}'

# Substituição completa (PUT)
curl -i -X PUT http://localhost:8080/v1/users/{id} \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ana Maria","email":"ana.maria@example.com","isActive":true}'
```

## Convenções e decisões (resumo)

* **PATCH:** `application/merge-patch+json` (RFC 7396) – simples e direto no Go.
* **PUT:** requer representação completa do recurso.
* **Paginação:** `page` (>=1) e `pageSize` (<=100), resposta com `{ data, meta }`.
* **Erros:** payload `{ code, message, details? }`.
* **Auth:** `bearerAuth` definido, **não obrigatório** por padrão nesta fase.

## Roadmap

* [ ] Implementar handlers Gin conforme OpenAPI
* [ ] Implementar service + repository (PostgreSQL)
* [ ] Cobertura de testes (unit e integração) para CRUD de `User`
* [ ] Configurar CI (lint, build, test)


### ✅ Entregas da Sprint 0 atendidas

* **Stack definida** (Go, Gin, PostgreSQL)
* **Repositório Git com estrutura** (diretórios e arquivos guia)
* **Docker + docker-compose com banco rodando** (serviço `db`, `api` e `swagger` prontos)
* **CRUD para User (definição OpenAPI)** com **POST**, **PATCH** e **PUT** detalhados (GET/DELETE incluídos)
* **README** com instruções de build/run/test

> Próximo passo (Sprint 1): codificar os handlers, serviços e repositórios conforme este OpenAPI, adicionar testes e conectar ao PostgreSQL.
