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
3. [Autenticação e Autorização](#autenticação-e-autorização)
4. [Como rodar com Docker Compose](#como-rodar-com-docker-compose)
5. [Como rodar localmente (sem Docker)](#como-rodar-localmente-sem-docker)
6. [Como testar a infraestrutura localmente (Localstack)](#como-testar-a-infraestrutura-localmente-localstack)
7. [Contribuições do GitHub Copilot](#contribuições-do-github-copilot)
8. [Recursos Adicionais](#recursos-adicionais)


## Objetivo
Preparar o ambiente e a estrutura mínima para iniciar o desenvolvimento do domínio `User` com **CRUD** completo definido em OpenAPI.

## Pré-requisitos
- Docker Desktop/Engine e Docker Compose
- Go 1.22+ (para desenvolvimento local fora do container)
- Terraform (apenas para desenvolvimento e deploy de infra)

## Autenticação e Autorização

Esta API utiliza **JWT (JSON Web Tokens)** para autenticação e **RBAC (Role-Based Access Control)** para autorização.

### Configuração JWT

As seguintes variáveis de ambiente são necessárias (veja `.env.example`):

```bash
JWT_ISSUER=https://cognito-idp.us-east-1.amazonaws.com/us-east-1_XXXXXXXXX
JWT_AUDIENCE=your-app-client-id
JWKS_URI=https://cognito-idp.us-east-1.amazonaws.com/us-east-1_XXXXXXXXX/.well-known/jwks.json
```

### Como obter um Access Token

#### Opção 1: AWS Cognito com Localstack (Desenvolvimento Local)

1. Inicie o Localstack (veja seção [Como testar a infraestrutura localmente](#como-testar-a-infraestrutura-localmente-localstack))

2. Deploy da infraestrutura Cognito:
   ```bash
   cd infra-localstack
   terraform init
   terraform apply
   ```

3. Obtenha um token usando o AWS CLI:
   ```bash
   # Autenticar um usuário
   aws cognito-idp initiate-auth \
     --auth-flow USER_PASSWORD_AUTH \
     --client-id <your-client-id> \
     --auth-parameters USERNAME=user@example.com,PASSWORD=YourPassword123! \
     --endpoint-url http://localhost:4566

   # O token estará em: AuthenticationResult.IdToken
   ```

4. Configure as variáveis para Localstack:
   ```bash
   JWT_ISSUER=http://localhost:4566
   JWT_AUDIENCE=<your-client-id-from-terraform-output>
   JWKS_URI=http://localhost:4566/.well-known/jwks.json
   ```

#### Opção 2: AWS Cognito em Produção

1. Deploy da infraestrutura (veja `infra/README.md`)

2. Use o Hosted UI ou Client Credentials:
   ```bash
   # Via Hosted UI (navegador):
   https://<your-cognito-domain>.auth.us-east-1.amazoncognito.com/oauth2/authorize?client_id=<client-id>&response_type=token&scope=openapi&redirect_uri=<redirect-uri>

   # Via Client Credentials:
   aws cognito-idp initiate-auth \
     --auth-flow USER_PASSWORD_AUTH \
     --client-id <your-client-id> \
     --auth-parameters USERNAME=user@example.com,PASSWORD=YourPassword
   ```

#### Opção 3: Mock Token para Testes (Desenvolvimento)

Para testes locais sem IdP, você pode desabilitar a autenticação não fornecendo as variáveis JWT. A API permitirá todas as requisições:

```bash
# Não defina JWT_ISSUER, JWT_AUDIENCE, JWKS_URI
# A API irá logar um warning e permitir acesso sem autenticação
```

### Usando o Token

Inclua o token no header `Authorization`:

```bash
curl -H "Authorization: Bearer <your-jwt-token>" \
  http://localhost:8080/users
```

### Controle de Acesso (RBAC)

Os seguintes roles são suportados (no claim `cognito:groups`):

- **admin-group**: Acesso total a todos os recursos
- **user-group**: Acesso limitado aos próprios recursos

Regras de autorização por endpoint:

| Endpoint | Método | Permissão Necessária |
|----------|--------|---------------------|
| `/users` | GET | Admin apenas |
| `/users/{id}` | GET | Dono do recurso ou Admin |
| `/users/{id}` | PUT | Dono do recurso ou Admin |
| `/users/{id}` | PATCH | Dono do recurso ou Admin |
| `/users/{id}` | DELETE | Admin apenas |
| `/users` | POST | Qualquer usuário autenticado |

### Exemplo de Request Autenticado

```bash
# Criar usuário
curl -X POST http://localhost:8080/users \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"Test User"}'

# Listar usuários (admin apenas)
curl -H "Authorization: Bearer eyJhbGc..." \
  http://localhost:8080/users?email=user@example.com

# Buscar usuário específico
curl -H "Authorization: Bearer eyJhbGc..." \
  http://localhost:8080/users/123
```

---

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

1. No terminal, inicialize o localstack
   ```bash
   localstack start
   ```

2. Na pasta ``infra-localstack``, execute o deploy com o terraform

   ```bash
   terraform plan
   ```
>Aqui, você já deve receber a confirmação visual, das estruturas que serão criadas ou possíveis erros encontrados.

---

## Testes

### Testes Unitários e E2E

Execute todos os testes:
```bash
go test ./... -v
```

Com cobertura:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Testes de Integração com Cognito-Local

O projeto inclui testes de integração que validam a autenticação JWT com Cognito usando Localstack:

```bash
# Executar testes de integração
go test -tags=integration ./internal/auth/... -v -timeout 10m
```

**Características dos testes de integração:**
- Iniciam um container Localstack automaticamente
- Criam user pool, app client e grupos do Cognito
- Testam autenticação de usuários e geração de tokens JWT
- Validam a estrutura dos tokens

**Nota**: O suporte completo ao Cognito requer Localstack Pro. Os testes pulam graciosamente na versão gratuita.

Para mais detalhes, consulte [internal/auth/INTEGRATION_TESTS.md](./internal/auth/INTEGRATION_TESTS.md).

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
