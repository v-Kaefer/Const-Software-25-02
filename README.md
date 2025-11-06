# Construção de Software 2025/02
Grupo L

https://github.com/v-Kaefer/Const-Software-25-02

![CI](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/ci.yaml/badge.svg)
![Tests](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/tests.yaml/badge.svg)
![Build](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/build.yaml/badge.svg)
![Docker Build](https://github.com/v-Kaefer/Const-Software-25-02/actions/workflows/docker-build.yaml/badge.svg)

# User Service – Go + PostgreSQL

> Serviço completo para o domínio **User**, com especificação **OpenAPI**, infraestrutura Docker, migração SQL, autenticação RBAC com AWS Cognito e CI/CD em GitHub Actions.

## Sumário
1. [Objetivo](#objetivo)
2. [Pré-requisitos](#pré-requisitos)
3. [Como rodar com Docker Compose](#como-rodar-com-docker-compose)
4. [Como rodar localmente (sem Docker)](#como-rodar-localmente-sem-docker)
5. [Como testar a infraestrutura localmente (Localstack)](#como-testar-a-infraestrutura-localmente-localstack)
6. [Autenticação RBAC](#autenticação-rbac)
7. [Contribuições do GitHub Copilot](#contribuições-do-github-copilot)
8. [Recursos Adicionais](#recursos-adicionais)


## Objetivo
Serviço REST completo para gerenciamento de usuários com autenticação RBAC, persistência em PostgreSQL, infraestrutura AWS e testes automatizados.

## Pré-requisitos
- Docker Desktop/Engine e Docker Compose
- Go 1.22+ (para desenvolvimento local fora do container)
- Terraform (para deploy de infraestrutura)
- AWS CLI (para testes com Cognito)

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
4. Suba API e Swagger:

   ```bash
   docker compose up -d api swagger
   # API:    http://localhost:8080
   # Swagger http://localhost:8081
   ```
5. Acompanhe logs:

   ```bash
   docker compose logs -f api
   ```

## Como rodar localmente (sem Docker)

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


## Como testar a infraestrutura localmente

### 🔥 Opção 1: cognito-local (RECOMENDADO - 100% Gratuito)

**📋 Pré-requisitos:**
- Docker e Docker Compose instalados
- AWS CLI instalado: `pip install awscli` ou `brew install awscli`

**Teste completo do Cognito localmente sem custos:**

```bash
# Passo 1: Iniciar cognito-local
make cognito-local-start

# Passo 2: Configurar (cria estrutura igual ao Terraform cognito.tf)
make cognito-local-setup

# Passo 3: Testar
make cognito-local-test

# Passo 4: Parar quando terminar
make cognito-local-stop
```

**O que é criado:**
- ✅ User Pool com políticas de senha
- ✅ App Client
- ✅ 3 Grupos (admin, reviewers, user)
- ✅ 3 Usuários de exemplo
- ✅ Arquivo de configuração JSON para integração

---

### Opção 2: LocalStack com tflocal (S3 + DynamoDB + IAM + VPC + Cognito)

**Usando o Makefile com tflocal (recomendado):**

```bash
# Ver todos os comandos disponíveis
make help

# Iniciar LocalStack
make localstack-start

# Aplicar Terraform usando tflocal (detecta automaticamente o LocalStack)
make tflocal-init
make tflocal-apply

# Testar a infraestrutura
make infra-test

# Destruir tudo
make tflocal-destroy
make localstack-stop
```

**Atalho com comando combinado:**

```bash
# Iniciar tudo de uma vez (LocalStack + tflocal init + tflocal apply)
make infra-up

# Testar a infraestrutura
make infra-test

# Destruir tudo (tflocal destroy + para LocalStack)
make infra-down
```

**Configuração das credenciais Cognito:**

Para criar usuários no Cognito, configure as credenciais antes de aplicar:
```bash
cd infra
cp credentials.tf.example credentials.tf
# Edite credentials.tf com seus usuários
```

>**⚠️ IMPORTANTE**: Cognito requer LocalStack Pro (pago). Para testar Cognito gratuitamente, use **cognito-local** (Opção 1 acima). Se usar LocalStack free, o Cognito não funcionará mas os outros recursos (S3, DynamoDB, IAM, VPC) funcionarão normalmente.

---

### Opção 3: Deploy na AWS (Produção)

**Usando o Makefile:**

```bash
# Configurar credenciais AWS (criar .aws/credentials no diretório infra/)
# e configurar usuários Cognito (copiar credentials.tf.example)

# Inicializar e aplicar
make infra-prod-init
make infra-prod-plan
make infra-prod-apply

# Destruir (cuidado!)
make infra-prod-destroy
```

>**📖 Documentação completa**: [infra/README.md](./infra/README.md)

---

## Autenticação RBAC

Este projeto implementa autenticação Role-Based Access Control (RBAC) integrada com AWS Cognito.

### Recursos de Autenticação

- **JWT Token Verification**: Validação automática de tokens Cognito
- **Role-Based Access Control**: Três níveis de acesso (admin, reviewer, user)
- **Middleware Protection**: Proteção de endpoints com autenticação e autorização
- **Context Integration**: Informações de usuário e roles disponíveis no contexto da requisição

### Configuração Rápida

1. Configure as variáveis de ambiente:
   ```bash
   COGNITO_REGION=us-east-1
   COGNITO_USER_POOL_ID=your-user-pool-id
   ```

2. Para testes locais, use cognito-local:
   ```bash
   make cognito-local-start
   make cognito-local-setup
   ```

3. Faça requisições autenticadas:
   ```bash
   curl -H "Authorization: Bearer <jwt-token>" http://localhost:8080/users
   ```

### Documentação Completa

Para detalhes completos sobre implementação, uso e troubleshooting, consulte:
**[docs/RBAC_AUTHENTICATION.md](./docs/RBAC_AUTHENTICATION.md)**

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
- **[docs/RBAC_AUTHENTICATION.md](./docs/RBAC_AUTHENTICATION.md)**: Documentação completa de autenticação RBAC com Cognito
