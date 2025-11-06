# Construção de Software 2025/02
Grupo L

https://github.com/v-Kaefer/Const-Software-25-02

![CI](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/ci.yaml/badge.svg)
![Tests](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/tests.yaml/badge.svg)
![Build](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/build.yaml/badge.svg)
![Docker Build](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/docker-build.yaml/badge.svg)

# User Service – API REST com Autenticação JWT/RBAC

> Serviço RESTful para gerenciamento de usuários com autenticação AWS Cognito, controle de acesso baseado em funções (RBAC) e infraestrutura como código.

## Sumário
1. [Pré-requisitos](#pré-requisitos)
2. [Início Rápido](#-início-rápido)
3. [Comandos Makefile Essenciais](#-comandos-makefile-essenciais)
4. [Variáveis de Ambiente](#-variáveis-de-ambiente-env)
5. [Autenticação e Autorização](#-autenticação-e-autorização)
6. [Documentação Completa](#-documentação-completa)
7. [Arquitetura](#-arquitetura)
8. [Testes](#-testes)
9. [Infraestrutura](#-infraestrutura)
10. [CI/CD](#-cicd)
11. [Contribuições do GitHub Copilot](#contribuições-do-github-copilot)
12. [Recursos Adicionais](#recursos-adicionais)

## Pré-requisitos
- Docker Desktop/Engine e Docker Compose
- Go 1.22+ (para desenvolvimento local fora do container)
- Terraform (apenas para desenvolvimento e deploy de infra)
- AWS CLI (para testes com Cognito)

## 🚀 Início Rápido

### Configuração Inicial

1. **Configure as variáveis de ambiente:**
   ```bash
   cp .env.example .env
   # Edite .env com suas configurações
   ```

2. **Inicie os serviços:**
   ```bash
   # Banco de dados + API
   docker compose up -d
   ```

3. **Aplique as migrações:**
   ```bash
   docker compose exec -T db psql -U app -d app -f /migrations/0001_init.sql
   ```

4. **Acesse a API:**
   - API: http://localhost:8080
   - Swagger: http://localhost:8081

## 📝 Comandos Makefile Essenciais

### Desenvolvimento Local
```bash
make help                    # Ver todos os comandos disponíveis

# Testes com Cognito Local (Recomendado)
make cognito-local-start     # Inicia cognito-local
make cognito-local-setup     # Configura usuários e grupos
make cognito-local-test      # Testa e obtém tokens JWT

# Infraestrutura Local (LocalStack + Cognito)
make infra-up               # Inicia toda infraestrutura local
make infra-test             # Testa recursos criados
make infra-down             # Para tudo e limpa recursos

# Testes e Build
go test ./...               # Executa todos os testes
go build ./cmd/api          # Compila a aplicação
```

### Deploy em Produção
```bash
make infra-prod-init        # Inicializa Terraform
make infra-prod-plan        # Revisa mudanças
make infra-prod-apply       # Aplica infraestrutura AWS
```

## 🔧 Variáveis de Ambiente (.env)

Copie `.env.example` para `.env` e configure:

### Aplicação
- `APP_ENV` - Ambiente (development/production)
- `APP_PORT` - Porta da API (padrão: 8080)

### Banco de Dados
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME`
- `DB_SSLMODE` - Modo SSL (disable para dev)

### Autenticação JWT/Cognito
- `COGNITO_REGION` - Região AWS (ex: us-east-1)
- `COGNITO_USER_POOL_ID` - ID do User Pool (deixe vazio para mock local)
- `JWT_ISSUER` - URL do emissor JWT (auto-construído se não fornecido)
- `JWT_AUDIENCE` - Client ID da aplicação (opcional)
- `JWKS_URI` - URL das chaves públicas (auto-construído se não fornecido)

**Exemplo para produção:**
```bash
JWT_ISSUER=https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123
JWT_AUDIENCE=seu-client-id
JWKS_URI=https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123/.well-known/jwks.json
```

## 🔐 Autenticação e Autorização

### Rotas da API

| Método | Rota | Permissão | Descrição |
|--------|------|-----------|-----------|
| POST | `/users` | Admin | Criar usuário |
| GET | `/users` | Admin | Listar todos os usuários |
| GET | `/users/{id}` | Admin ou Próprio | Obter usuário por ID |
| PUT | `/users/{id}` | Admin ou Próprio | Atualizar usuário |
| PATCH | `/users/{id}` | Admin ou Próprio | Atualizar parcialmente |
| DELETE | `/users/{id}` | Admin | Deletar usuário |

### Como Obter Token JWT

**Opção 1 - Cognito Local (Desenvolvimento):**
```bash
make cognito-local-start
make cognito-local-setup
make cognito-local-test  # Exibe tokens gerados
```

**Opção 2 - AWS Cognito (Produção):**
```bash
aws cognito-idp initiate-auth \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id <seu-client-id> \
  --auth-parameters USERNAME=admin@example.com,PASSWORD=SuaSenha123! \
  --region us-east-1
```

### Fazendo Requisições

```bash
# Exemplo: Listar usuários (admin apenas)
curl -H "Authorization: Bearer SEU_TOKEN_JWT" \
     http://localhost:8080/users

# Exemplo: Criar usuário
curl -X POST \
     -H "Authorization: Bearer SEU_TOKEN_JWT" \
     -H "Content-Type: application/json" \
     -d '{"email":"novo@example.com","name":"Novo Usuario"}' \
     http://localhost:8080/users
```

## 📚 Documentação Completa

- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Guias de desenvolvimento e convenções
- **[docs/RBAC_AUTHENTICATION.md](./docs/RBAC_AUTHENTICATION.md)** - Documentação detalhada de autenticação RBAC
- **[CHANGELOG.md](./CHANGELOG.md)** - Histórico de mudanças e sprints
- **[COPILOT_INSTRUCTIONS.md](./COPILOT_INSTRUCTIONS.md)** - Contribuições do GitHub Copilot
- **[infra/README.md](./infra/README.md)** - Documentação de infraestrutura e Terraform

## 🏗️ Arquitetura

```
├── cmd/api/              # Ponto de entrada da aplicação
├── internal/
│   ├── auth/            # Middleware JWT/RBAC
│   ├── config/          # Configurações
│   ├── db/              # Conexão e migrações
│   └── http/            # Handlers HTTP
├── pkg/user/            # Domínio User (service, repo)
├── infra/               # Infraestrutura como código (Terraform)
├── docs/                # Documentação adicional
├── migrations/          # Scripts SQL
└── openapi/             # Especificação OpenAPI 3.1
```

## 🧪 Testes

```bash
# Todos os testes
go test ./...

# Com cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Testes específicos
go test ./internal/auth/... -v    # Testes de autenticação
go test ./internal/http/... -v    # Testes de handlers
```

**Cobertura Atual:** 58.3% (74.6% auth, 67.4% http)

## 🛠️ Infraestrutura

### Recursos AWS (Terraform)
- Cognito User Pool com grupos (admin, reviewer, user)
- S3, DynamoDB, VPC, IAM
- Configurável para LocalStack (desenvolvimento)

### Arquivos de Configuração
- `infra/credentials.tf.example` - Template para usuários Cognito (copie para `credentials.tf`)
- `.env.example` - Template de variáveis de ambiente (copie para `.env`)

## 📊 CI/CD

GitHub Actions configurado com:
- ✅ Build e testes automáticos
- ✅ Linting (go vet)
- ✅ Cobertura de código
- ✅ Docker build
- ✅ Execução em push/PR

---

## Contribuições do GitHub Copilot

Este projeto utilizou o GitHub Copilot para auxiliar no diagnóstico e correção de problemas técnicos específicos.

### Correção de Workflows CI/CD
O Copilot foi utilizado para identificar e corrigir problemas nos workflows de CI/CD:
- **Correção de Execução de Testes**: Alterou comandos de teste para executar todos os testes (`./...`) ao invés de apenas um pacote
- **Correção de Sintaxe YAML**: Corrigiu triggers de tags no workflow docker-build
- **Remoção de Dependências Inválidas**: Removeu dependências de jobs que causavam falhas nos workflows

### Implementação de Autenticação JWT/RBAC
O Copilot implementou autenticação JWT completa e controle de acesso baseado em funções (RBAC):
- **Validação JWT com JWKS**: Verificação de claims (iss, aud, exp, nbf) e assinaturas
- **Rotas CRUD Protegidas**: Endpoints com controle de acesso baseado em funções
- **Testes Abrangentes**: 76 testes implementados (JWT validation + RBAC)
- **Documentação Completa**: README, OpenAPI e guias traduzidos para PT-BR

Para informações detalhadas sobre as contribuições do Copilot, consulte [COPILOT_INSTRUCTIONS.md](./COPILOT_INSTRUCTIONS.md).

---

## Recursos Adicionais

- **[CONTRIBUTING.md](./CONTRIBUTING.md)**: Guias de desenvolvimento, convenções e instruções detalhadas de setup
- **[CHANGELOG.md](./CHANGELOG.md)**: Revisões de sprints e histórico do projeto
- **[COPILOT_INSTRUCTIONS.md](./COPILOT_INSTRUCTIONS.md)**: Rastreamento completo das contribuições do GitHub Copilot
- **[docs/RBAC_AUTHENTICATION.md](./docs/RBAC_AUTHENTICATION.md)**: Documentação completa de autenticação RBAC com Cognito

---

Desenvolvido por **Grupo L** com assistência do **GitHub Copilot** para implementação de autenticação JWT/RBAC.
