# Construção de Software 2025/02
Grupo L

https://github.com/v-Kaefer/Const-Software-25-02

![CI](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/ci.yaml/badge.svg)
![Tests](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/tests.yaml/badge.svg)
![Build](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/build.yaml/badge.svg)
![Docker Build](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/docker-build.yaml/badge.svg)

# User Service – Go + Gin + PostgreSQL

> Serviço base para o domínio **User**, com especificação **OpenAPI**, infraestrutura Docker, migração SQL e CI simples em GitHub Actions.

## Sumário
1. [Objetivo](#objetivo)
2. [Pré-requisitos](#pré-requisitos)
3. [Como rodar com Docker Compose](#como-rodar-com-docker-compose)
4. [Como rodar localmente (sem Docker)](#como-rodar-localmente-sem-docker)
5. [Como testar a infraestrutura localmente (Localstack)](#como-testar-a-infraestrutura-localmente-localstack)
6. [Contribuições do GitHub Copilot](#contribuições-do-github-copilot)
7. [Recursos Adicionais](#recursos-adicionais)


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


## Como testar a infraestrutura localmente (Localstack)

**Usando o Makefile (Recomendado):**

```bash
# Ver todos os comandos disponíveis
make help

# Iniciar LocalStack e aplicar Terraform
make infra-up

# Testar a infraestrutura
make infra-test

# Destruir tudo
make infra-down
```

**Manualmente:**

1. No terminal, inicialize o localstack
   ```bash
   localstack start
   ```

2. Na pasta ``infra-localstack``, configure as variáveis (opcional):
   ```bash
   cd infra-localstack
   cp terraform.tfvars.example terraform.tfvars
   # Edite terraform.tfvars conforme necessário
   ```

3. Execute o deploy com o terraform:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

>**⚠️ IMPORTANTE**: Cognito requer LocalStack Pro. Para usar o free tier, renomeie temporariamente `cognito.tf` para `cognito.tf.disabled`. Veja detalhes em [infra-localstack/README.md](./infra-localstack/README.md).

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
