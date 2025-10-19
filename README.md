# Construção de Software 2025/02
Grupo L

https://github.com/v-Kaefer/Const-Software-25-02

![Tests](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/tests.yaml/badge.svg)
![Docker Build](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/docker-build.yaml/badge.svg)

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

# User Service – Go + Gin + PostgreSQL

> Serviço base para o domínio **User**, com especificação **OpenAPI**, infraestrutura Docker, migração SQL e CI simples em GitHub Actions.

## Sumário
1. [Objetivo](#objetivo)
2. [Pré-requisitos](#pré-requisitos)
3. [Como rodar com Docker Compose](#como-rodar-com-docker-compose)
4. [Como rodar localmente (sem Docker)](#como-rodar-localmente-sem-docker)
5. [GitHub Copilot Contributions](#github-copilot-contributions)
6. [Additional Resources](#additional-resources)


## Objetivo
Preparar o ambiente e a estrutura mínima para iniciar o desenvolvimento do domínio `User` com **CRUD** completo definido em OpenAPI.

## Pré-requisitos
- Docker Desktop/Engine e Docker Compose
- Go 1.22+ (para desenvolvimento local fora do container)
- Terraform (apenas para desenvolvimento e deploy de infra)

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

> Útil quando o servidor Gin for implementado.

1. Garanta um PostgreSQL local acessível.
2. Configure `DATABASE_URL` (ver [Variáveis de ambiente](./CONTRIBUTING.md)).
3. Aplique a migração:

   ```bash
   psql "$DATABASE_URL" -f migrations/0001_init.sql
   ```
4. Rode a aplicação:

   ```bash
   go run ./cmd/api
   ```

---

## GitHub Copilot Contributions

This project leverages GitHub Copilot to accelerate development and maintain high code quality. Below is a summary of the key areas where Copilot has contributed to the project:

### Core Application Components
- **Go + Gin Framework Setup**: Complete REST API structure with handlers, services, and repositories
- **User Domain Implementation**: Full CRUD operations for user management
- **Database Integration**: PostgreSQL connection, GORM ORM, and migration tools
- **Configuration Management**: Environment-based configuration system

### Infrastructure & DevOps
- **Containerization**: Dockerfile and docker-compose.yaml for multi-service orchestration
- **AWS Infrastructure**: Terraform configurations for VPC, S3, DynamoDB, and IAM
- **LocalStack Setup**: Local AWS simulation for development and testing
- **CI/CD Pipelines**: GitHub Actions workflows for build, test, and deployment

### Testing & Quality
- **Comprehensive Test Suite**: Unit tests, integration tests, and E2E tests
- **Test Coverage**: Automated coverage reporting in CI/CD
- **Mock Implementations**: In-memory database for testing

### API Documentation
- **OpenAPI Specification**: Complete API documentation with request/response schemas
- **Swagger UI**: Interactive API documentation via Docker

### Development Tools
- **Database Migrations**: SQL-based migration system
- **Code Quality Tools**: Go formatting, linting, and best practices
- **Documentation**: Contributing guidelines, changelog, and setup instructions

For detailed information about all Copilot contributions, see [COPILOT_INSTRUCTIONS.md](./COPILOT_INSTRUCTIONS.md).

---

## Additional Resources

- **[CONTRIBUTING.md](./CONTRIBUTING.md)**: Development guidelines, conventions, and detailed setup instructions
- **[CHANGELOG.md](./CHANGELOG.md)**: Sprint reviews and project history
- **[COPILOT_INSTRUCTIONS.md](./COPILOT_INSTRUCTIONS.md)**: Complete tracking of GitHub Copilot contributions

