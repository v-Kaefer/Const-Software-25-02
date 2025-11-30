# Infraestrutura de Produção (AWS)

Esta pasta contém as definições Terraform para a infraestrutura de produção na AWS.

## 📋 Pré-requisitos

- [Terraform](https://www.terraform.io/) >= 1.2
- [AWS CLI](https://aws.amazon.com/cli/) configurado com credenciais válidas
- Credenciais AWS configuradas em `.aws/credentials` (ver seção de configuração)

## 🚀 Início Rápido

### 1. Configure as credenciais AWS

Crie o arquivo `.aws/credentials` neste diretório:
```bash
mkdir -p .aws
cat > .aws/credentials << EOF
[default]
aws_access_key_id = SUA_ACCESS_KEY
aws_secret_access_key = SUA_SECRET_KEY
EOF
```

### 2. Configure as credenciais do Cognito

```bash
cp credentials.tf.example credentials.tf
# Edite credentials.tf com os usuários que deseja criar
```

### 3. Execute o Terraform

**Do diretório raiz do projeto:**

```bash
# Inicializar
make infra-prod-init

# Ver o plano de execução
make infra-prod-plan

# Aplicar a infraestrutura
make infra-prod-apply

# Destruir a infraestrutura (cuidado!)
make infra-prod-destroy
```

**Ou manualmente:**

```bash
cd infra
terraform init
terraform plan
terraform apply
terraform destroy
```

## 📦 Recursos Criados

### Compute
- **EC2 Instance**: `grupo-l-sprint1` (t2.micro)
- **Key Pair**: `grupo-l-key`

### Storage
- **S3 Bucket**: `grupo-l-terraform`
- **DynamoDB Table**: `GrupoLConstSoftSprint1DynamoDB`

### Networking
- **Security Group**: `allow-http`
  - Inbound: SSH (22), HTTP (8080), ICMP
  - Outbound: ICMP

### IAM
- **IAM Role**: `ec2_role` (com permissões para S3 e DynamoDB)

### Cognito
- **User Pool**: `CognitoUserPool`
- **Identity Pool**: `MyIdentityPool`
- **User Groups**: admin-group, reviewers-group, user-group
- **IAM Roles**: Para cada grupo de usuários
- **Senhas temporárias**: geradas automaticamente (veja `terraform output admin_temp_password`, `reviewer_temp_password`, `user_temp_password` após o apply)

## 🔧 Comandos Make Disponíveis

| Comando | Descrição |
|---------|-----------|
| `make infra-prod-init` | Inicializa o Terraform |
| `make infra-prod-plan` | Executa terraform plan |
| `make infra-prod-apply` | Aplica a infraestrutura |
| `make infra-prod-destroy` | Destrói a infraestrutura |

## 🧪 Testes Locais

Para testar a infraestrutura localmente antes de aplicar na AWS, use o LocalStack com tflocal:

```bash
# Opção 1: Usar comando combinado (recomendado)
make infra-up  # Inicia LocalStack, cognito-local e aplica infra

# Opção 2: Passo a passo
make localstack-start
make cognito-local-start
make tflocal-init
make cognito-local-setup
make tflocal-apply

# Testar os recursos
make infra-test

# Destruir quando terminar
make infra-down  # Para tudo automaticamente
```

**💡 Notas sobre recursos**: 
- **EC2**: Suportado no LocalStack free tier com AMI mock automático
- **Cognito**: Usa cognito-local (alternativa gratuita) - automaticamente excluído do tflocal
- **S3, DynamoDB, IAM, VPC**: Todos funcionam com LocalStack free tier
- **Configuração automática**: Os comandos `tflocal-*` automaticamente:
  - Usam AMI mock (`ami-ff0fea8310f3`) para EC2
  - Excluem Cognito (substituído por cognito-local)
  - Em produção, tudo funciona normalmente com recursos reais

**Testando Cognito separadamente:**
```bash
make cognito-local-test
```

Ver [../README.md](../README.md) para mais detalhes sobre as opções de teste.

## 📝 Notas

- O arquivo `credentials.tf` contém informações sensíveis e está no `.gitignore`
- Sempre use `credentials.tf.example` como referência para criar seu `credentials.tf`
- O arquivo `.aws/credentials` também está no `.gitignore` por segurança
- Revise sempre o `terraform plan` antes de aplicar mudanças na produção

## 🐛 Troubleshooting

### Erro de autenticação AWS
Verifique se suas credenciais AWS estão configuradas corretamente:
```bash
aws configure list
aws sts get-caller-identity
```

### Erro ao criar Cognito User Pool
Certifique-se de que o arquivo `credentials.tf` existe e está configurado corretamente:
```bash
cp credentials.tf.example credentials.tf
# Edite o arquivo com seus dados
```

### Conflito de recursos
Se recursos já existem na AWS, use `terraform import` ou ajuste os nomes nos arquivos `.tf`.
