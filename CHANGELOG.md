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


# **Sprint 1 – Definição e Gerenciamento de Infraestrutura como Código (IaC)**

## Objetivo:
Ampliar o projeto atual de provisionamento e configuração de infraestrutura para incluir a definição e o gerenciamento da infraestrutura como código (IaC) utilizando ferramentas como Terraform, AWS SAM, Serverless Framework, Pulumi, ou outra solução adequada.

> **O objetivo é definir e automatizar a criação de todos os recursos necessários na AWS para hospedar e executar o projeto de software.**

### Pré-requisitos:

* Conta AWS com permissões necessárias.

* Projeto de software hospedado no GitHub.

* Conhecimento básico de AWS, GitHub Actions e Docker.

* Conhecimento básico em IaC.


# **Sprint 0 – Setup de Time, Stack e Projeto**

## User Service – Go + Gin + PostgreSQL

Este pacote entrega um **arquivo fonte OpenAPI** para o domínio `User` (com **POST**, **PATCH** e **PUT**), um **README** passo‑a‑passo, além de arquivos básicos de infraestrutura (Docker/Docker Compose e migração SQL) para iniciar o projeto com Go, Gin e PostgreSQL.

> Serviço base para o domínio **User**, com especificação **OpenAPI**, infraestrutura Docker, migração SQL e CI simples em GitHub Actions.