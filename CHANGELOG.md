# Changelog

Todas as modificações e entregas de sprints para esse projeto, estão documentadas nesse arquivo.

## Sprint 2 Updates

### Infrastructure Consolidation (2025-11-05)

#### EC2 support in LocalStack free tier (2025-11-05)

* **Correção: EC2 é suportado no LocalStack free tier**:
  - Revertida separação de recursos EC2 - não é necessária
  - EC2 resources movidos de volta para `infra/main.tf`
  - Removido `infra/ec2.tf` (não é mais necessário)
  - Removida lógica de exclusão automática de ec2.tf dos comandos tflocal
  - LocalStack free tier suporta EC2 conforme documentação oficial
  - tflocal funciona com EC2 usando implementação mock do LocalStack

#### Consolidação da estrutura de infraestrutura

* **Remoção de infra-localstack**:
  - Removido diretório `infra-localstack` completo
  - Consolidado todos os testes locais para usar `tflocal` no diretório `infra/`
  - Scripts cognito-local movidos para `infra/`
  
* **Uso unificado do tflocal**:
  - Comandos `make tflocal-*` agora operam em `infra/` com tflocal
  - `tflocal` detecta automaticamente endpoints do LocalStack
  - Não requer configuração manual de endpoints ou arquivos separados
  - Mesma estrutura para testes locais e produção
  
* **Comandos Makefile simplificados**:
  - `make tflocal-init/plan/apply/destroy` - Para testes locais com LocalStack
  - `make infra-prod-init/plan/apply/destroy` - Para deploy em produção na AWS
  - `make infra-up/infra-down` - Atalhos para iniciar/parar tudo
  - Removidos comandos `terraform-*` obsoletos
  
* **Documentação atualizada**:
  - Atualizado `README.md` principal com fluxo simplificado
  - Atualizado `infra/README.md` removendo referências a infra-localstack
  - Atualizado `.gitignore` para refletir nova estrutura

### Previous Infrastructure Updates (2025-11-05)

#### Atualização da estrutura de infraestrutura

* **Sincronização infra com infra-localstack**:
  - Adicionado `cognito.tf` ao diretório `infra/` com recursos completos do Cognito
  - Adicionado `credentials.tf.example` para configuração de usuários Cognito
  - Mantida configuração de produção em `infra/main.tf` (sem credenciais "test")
  
* **Suporte a tflocal (terraform-local)**:
  - Adicionados comandos `make tflocal-*` para uso com LocalStack
  - `tflocal` detecta automaticamente endpoints do LocalStack
  - Não requer configuração manual de endpoints
  
* **Comandos Makefile organizados**:
  - `make tflocal-init/plan/apply/destroy` - Para testes com LocalStack usando tflocal
  - `make infra-prod-init/plan/apply/destroy` - Para deploy em produção na AWS
  
* **Documentação atualizada**:
  - Criado `infra/README.md` com guia completo de produção
  - Atualizado `README.md` principal com três opções de infraestrutura:
    1. cognito-local (gratuito, para testes)
    2. LocalStack com tflocal (gratuito, sem Cognito Pro)
    3. AWS Produção (deploy real)
  - Atualizado `.gitignore` para incluir `infra/credentials.tf`

# Entregas da Sprint 2 (D.O.D.)

* **Autenticação**


## Add Autenticação 

https://aws.amazon.com/pt/getting-started/hands-on/build-serverless-web-app-lambda-apigateway-s3-dynamodb-cognito/

## Arquitetura do AWS Cognito
O Cognito tem dois componentes principais:

* User Pools: Autenticação (quem você é)
* Identity Pools: Autorização (o que você pode fazer)

### Estrutura

* AWS Amplify
Hospedagem de site estático (HTML, CSS, JavaScript, etc.)
--
* Amazon Cognito
Gerenciamento de usuários
--
* Amazon API Gateway
Back-end sem servidor
--
* Amazon DynamoDB
AI RESTful


## Add LocalStack

https://docs.localstack.cloud/aws/getting-started/installation/

---

# **Sprint 1 – Setup de Infraestrutura com Terraform (IaC), para AWS**

Infrastructure - Terraform + AWS + Github Actions + Docker

## Objetivo:
Ampliar o projeto atual de provisionamento e configuração de infraestrutura para incluir a definição e o gerenciamento da infraestrutura como código (IaC) utilizando ferramentas como Terraform, AWS SAM, Serverless Framework, Pulumi, ou outra solução adequada.

> **O objetivo é definir e automatizar a criação de todos os recursos necessários na AWS para hospedar e executar o projeto de software.**

## Pré-requisitos:
* Conta AWS com permissões necessárias.
* Projeto de software hospedado no GitHub.
* Conhecimento básico de AWS, GitHub Actions e Docker.
* Conhecimento básico em IaC.
* Terraform (apenas para desenvolvimento e deploy de infra)

## Entregas da Sprint 1 (Definition of Done)

**Infraestrutura do projeto definida e gerenciada na AWS:**
* Scripts e configurações IaC.
* Todos os artefatos necessários para a configuração da infraestrutura na AWS.
* Projeto no GitHub contendo a pasta (infra)

**Entrega Final:**
O trabalho deve ser entregue em um arquivo .zip contendo o repositório de fontes completo, incluindo a pasta infra.

---

# **Sprint 0 – Setup de Time, Stack e Projeto**

Este pacote entrega um **arquivo fonte OpenAPI** para o domínio `User` (com **POST**, **PATCH** e **PUT**), um **README** passo‑a‑passo, além de arquivos básicos de infraestrutura (Docker/Docker Compose e migração SQL) para iniciar o projeto com Go, Gin e PostgreSQL.

## User Service – Go + Gin + PostgreSQL

> Serviço base para o domínio **User**, com especificação **OpenAPI**, infraestrutura Docker, migração SQL e CI simples em GitHub Actions.

## Entregas da Sprint 0 (Definition of Done)

* **Stack definida** (Go, Gin, PostgreSQL)
* **Repositório Git com estrutura** (diretórios e arquivos guia)
* **Docker + docker-compose com banco rodando** (serviço `db`, `api` e `swagger` prontos)
* **CRUD para User (definição OpenAPI)** com **POST**, **PATCH** e **PUT** detalhados (GET/DELETE incluídos)
* **README** com instruções de build/run/test