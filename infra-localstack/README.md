# Definições Localstack (Para testes de infraestrutura pré deploy)

* Para uma visualização geral da infra definida aqui utilize o comando ```terraform graph```

## 🔥 RECOMENDADO: cognito-local (100% Gratuito)

**📋 Pré-requisitos para cognito-local:**
- ✅ Docker e Docker Compose
- ✅ **AWS CLI** - Necessário para configurar o cognito-local
  - **Instalar**: `pip install awscli` ou `brew install awscli` (macOS)
  - **Verificar**: `aws --version`

**Para testar Cognito localmente SEM CUSTOS, use cognito-local:**

```bash
# Do diretório raiz do projeto:

# 1. Iniciar cognito-local
make cognito-local-start

# 2. Configurar (replica estrutura do cognito.tf)
make cognito-local-setup

# 3. Testar
make cognito-local-test

# 4. Parar
make cognito-local-stop
```

**⚠️ Nota Importante**:
- O script de setup limpa automaticamente User Pools existentes para evitar conflitos
- Não conflita com arquivos Terraform (usa AWS CLI diretamente no cognito-local)
- Para recomeçar do zero: `make cognito-local-clean && make cognito-local-start && make cognito-local-setup`

**📖 Guia completo**: [COGNITO-LOCAL-SETUP.md](./COGNITO-LOCAL-SETUP.md)

---

## ⚠️ IMPORTANTE: Limitação do Cognito no LocalStack Free Tier

**O Cognito NÃO está disponível no LocalStack free tier!** Para usar Cognito com LocalStack, você precisa:
1. **LocalStack Pro** (pago) - [Saiba mais](https://localstack.cloud/pricing/)
2. **Alternativa GRATUITA**: Use **cognito-local** (veja acima) ✅

### Opções para testar a infraestrutura:

#### Opção A: cognito-local (GRATUITO - Recomendado)
```bash
# Veja seção acima "RECOMENDADO: cognito-local"
make cognito-local-start
make cognito-local-setup
make cognito-local-test
```

#### Opção B: LocalStack Pro (Cognito completo - Pago)
```bash
export LOCALSTACK_API_KEY=seu-api-key
make infra-up
```

#### Opção C: Free Tier (S3 e DynamoDB apenas - Sem Cognito)
```bash
# Renomear cognito.tf temporariamente
mv cognito.tf cognito.tf.disabled
make infra-up
# Depois de testar, restaurar: mv cognito.tf.disabled cognito.tf
```

## 🚀 Início Rápido com Makefile

**Do diretório raiz do projeto**, você pode usar o Makefile para gerenciar toda a infraestrutura:

```bash
# Ver todos os comandos disponíveis (inclui cognito-local)
make help

# Para Cognito (GRATUITO):
make cognito-local-start   # Inicia cognito-local
make cognito-local-setup   # Configura
make cognito-local-test    # Testa

# Para LocalStack (sem Cognito):
make infra-up              # Iniciar LocalStack
make infra-test            # Testar recursos
make infra-down            # Destruir tudo
```

## Teste local (Manual)
>Você pode executar o terraform, mesmo sem executar o Localstack, mas vai retornar erros nos serviços: DynamoDB, IAM, VPC e Cognito.

### Para realizar o teste local é necessário

* [Localstack CLI](https://app.localstack.cloud/getting-started)
* [AWS CLI](https://aws.amazon.com/cli/) - Para testar os recursos
* [Docker](https://www.docker.com/)
* [Terraform](https://www.terraform.io/)

### 1. Configure as credenciais

```bash
cd infra-localstack
cp credentials.tf.example credentials.tf
# Edite credentials.tf com seus usuários (opcional - tem valores padrão)
```

### 2. Execute o Localstack
> Recomendação: Execute o localstack no terminal ao invés de diretamente no vscode.

```bash
localstack start
# Ou em background: localstack start -d
```

### 3. Execute o Terraform
Vá até a pasta ```infra-localstack```

```bash
terraform init
terraform plan
terraform apply
```

## 📦 Recursos Criados

### ✅ Disponíveis no Free Tier
- **S3**: Bucket `my-localstack-bucket`
- **DynamoDB**: Tabela `GrupoLConstSoftSprint1DynamoDB`

### ❌ Requerem LocalStack Pro
- **Cognito User Pool**: `CognitoUserPool`
- **Cognito Identity Pool**: `MyIdentityPool`
- **Cognito User Groups**: admin, reviewer, user
- **IAM Roles**: Para cada grupo de usuários

## 🧪 Testando os Recursos

### S3
```bash
aws --endpoint-url=http://localhost:4566 s3 ls
aws --endpoint-url=http://localhost:4566 s3 mb s3://test-bucket
```

### DynamoDB
```bash
aws --endpoint-url=http://localhost:4566 dynamodb list-tables
aws --endpoint-url=http://localhost:4566 dynamodb describe-table --table-name GrupoLConstSoftSprint1DynamoDB
```

### Cognito (apenas com Pro)
```bash
aws --endpoint-url=http://localhost:4566 cognito-idp list-user-pools --max-results 10
```

## 4. Interação visual com a Infraestrutura
Você pode visualizar e interagir com a infraestrutura da mesma forma que a AWS, no [Dashboard do Localstack](https://app.localstack.cloud/inst/default/resources).

## 🔧 Comandos Make Disponíveis

| Comando | Descrição |
|---------|-----------|
| `make help` | Mostra todos os comandos disponíveis |
| `make localstack-start` | Inicia o LocalStack |
| `make localstack-stop` | Para o LocalStack |
| `make localstack-status` | Verifica o status |
| `make terraform-init` | Inicializa o Terraform |
| `make terraform-plan` | Executa terraform plan |
| `make terraform-apply` | Aplica a infraestrutura |
| `make terraform-destroy` | Destrói a infraestrutura |
| `make infra-up` | Start + Init + Apply |
| `make infra-down` | Destroy + Stop |
| `make infra-test` | Testa todos os recursos |

## 🐛 Troubleshooting

### Erro: "Cognito not available"
Este é o comportamento esperado no free tier. Veja as opções acima.

### LocalStack não inicia
```bash
localstack status
docker ps -a | grep localstack
make localstack-clean  # Limpa containers antigos
```

### Terraform não conecta
```bash
curl http://localhost:4566/_localstack/health
localstack logs
```