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
5. [Contribuições do GitHub Copilot](#contribuições-do-github-copilot)
6. [Recursos Adicionais](#recursos-adicionais)


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

## Contribuições do GitHub Copilot

Este projeto utilizou o GitHub Copilot para auxiliar no diagnóstico e correção de problemas técnicos específicos.

### Correção de Workflows CI/CD
O Copilot foi utilizado para identificar e corrigir problemas nos workflows de CI/CD:
- **Correção de Execução de Testes**: Alterou comandos de teste para executar todos os testes (`./...`) ao invés de apenas um pacote
- **Correção de Sintaxe YAML**: Corrigiu triggers de tags no workflow docker-build
- **Remoção de Dependências Inválidas**: Removeu dependências de jobs que causavam falhas nos workflows

Para informações detalhadas sobre as contribuições do Copilot, consulte [COPILOT_INSTRUCTIONS.md](./COPILOT_INSTRUCTIONS.md).

---

## Recursos Adicionais

- **[CONTRIBUTING.md](./CONTRIBUTING.md)**: Guias de desenvolvimento, convenções e instruções detalhadas de setup
- **[CHANGELOG.md](./CHANGELOG.md)**: Revisões de sprints e histórico do projeto
- **[COPILOT_INSTRUCTIONS.md](./COPILOT_INSTRUCTIONS.md)**: Rastreamento completo das contribuições do GitHub Copilot

