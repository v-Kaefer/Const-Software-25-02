# Infrastructure as Code - Production

Este diretório contém a infraestrutura Terraform para deploy em **produção** na AWS.

## Recursos Provisionados

- **Cognito User Pool**: Provedor de identidade (IdP) para autenticação JWT
- **Cognito User Pool Client**: Aplicação cliente para obtenção de tokens
- **Cognito User Groups**: Grupos para RBAC (admin-group, user-group)
- **Cognito Identity Pool**: Pool de identidades para acesso a recursos AWS (opcional)
- **DynamoDB, S3, VPC, IAM**: Outros recursos da infraestrutura

## Pré-requisitos

1. **AWS CLI** configurado com credenciais válidas:
   ```bash
   aws configure
   ```

2. **Terraform** 1.0+ instalado:
   ```bash
   terraform version
   ```

3. Permissões IAM necessárias:
   - `cognito-idp:*`
   - `cognito-identity:*`
   - `iam:*` (para roles e políticas)
   - Outras permissões conforme recursos adicionais

## Como Usar

### 1. Inicializar Terraform

```bash
cd infra
terraform init
```

### 2. Configurar Variáveis

Crie um arquivo `terraform.tfvars`:

```hcl
aws_region           = "us-east-1"
environment          = "production"
cognito_domain_prefix = "your-app-name"
project_name         = "user-service"
```

Ou use variáveis de ambiente:

```bash
export TF_VAR_aws_region="us-east-1"
export TF_VAR_environment="production"
```

### 3. Planejar Deploy

```bash
terraform plan
```

Revise os recursos que serão criados.

### 4. Aplicar Infraestrutura

```bash
terraform apply
```

Digite `yes` para confirmar.

### 5. Obter Outputs

Após o apply, capture os outputs importantes:

```bash
terraform output jwt_issuer
terraform output jwks_uri
terraform output cognito_client_id
```

Configure essas variáveis no `.env` da aplicação:

```bash
JWT_ISSUER=$(terraform output -raw jwt_issuer)
JWT_AUDIENCE=$(terraform output -raw cognito_client_id)
JWKS_URI=$(terraform output -raw jwks_uri)
```

## Variáveis Importantes

| Variável | Descrição | Default |
|----------|-----------|---------|
| `aws_region` | Região AWS para deploy | `us-east-1` |
| `environment` | Ambiente (dev/staging/prod) | `development` |
| `cognito_domain_prefix` | Prefixo para Hosted UI | `user-service` |
| `project_name` | Nome do projeto | `user-service` |

## Outputs Importantes

| Output | Descrição | Uso |
|--------|-----------|-----|
| `jwt_issuer` | Issuer do JWT | Configurar como `JWT_ISSUER` |
| `jwks_uri` | URI do JWKS | Configurar como `JWKS_URI` |
| `cognito_client_id` | ID do cliente | Configurar como `JWT_AUDIENCE` |
| `cognito_user_pool_id` | ID do User Pool | Gerenciamento de usuários |
| `cognito_domain` | URL do Hosted UI | Login via navegador |

## Gerenciamento de Usuários

### Criar Usuário Admin via AWS CLI

```bash
# Criar usuário
aws cognito-idp admin-create-user \
  --user-pool-id <USER_POOL_ID> \
  --username admin@example.com \
  --user-attributes Name=email,Value=admin@example.com Name=name,Value="Admin User" \
  --temporary-password "TempPass123!" \
  --message-action SUPPRESS

# Adicionar ao grupo admin
aws cognito-idp admin-add-user-to-group \
  --user-pool-id <USER_POOL_ID> \
  --username admin@example.com \
  --group-name admin-group

# Definir senha permanente (após primeiro login)
aws cognito-idp admin-set-user-password \
  --user-pool-id <USER_POOL_ID> \
  --username admin@example.com \
  --password "SecurePassword123!" \
  --permanent
```

### Obter Token JWT

```bash
# Via USER_PASSWORD_AUTH
aws cognito-idp initiate-auth \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id <CLIENT_ID> \
  --auth-parameters USERNAME=user@example.com,PASSWORD=YourPassword123!

# O token estará em AuthenticationResult.IdToken
```

## Configuração do Hosted UI (Opcional)

Para habilitar login via navegador, configure:

1. No Cognito Console ou via Terraform, adicione:
   - Callback URLs: `http://localhost:3000/callback`
   - Logout URLs: `http://localhost:3000/logout`

2. Acesse:
   ```
   https://<cognito_domain_prefix>.auth.<region>.amazoncognito.com/login?client_id=<client_id>&response_type=token&scope=openid&redirect_uri=<callback_url>
   ```

## Limpeza

Para destruir a infraestrutura:

```bash
terraform destroy
```

**Atenção**: Isso irá deletar todos os recursos, incluindo usuários e dados.

## Troubleshooting

### Erro: "User pool domain already exists"

Se o domínio já existir, altere `cognito_domain_prefix` para um valor único.

### Erro: "Invalid token"

Verifique:
1. `JWT_ISSUER` corresponde ao output `jwt_issuer`
2. `JWT_AUDIENCE` corresponde ao output `cognito_client_id`
3. `JWKS_URI` está acessível e correto
4. Token não expirou (validade padrão: 60 minutos)

## Estrutura de Arquivos

```
infra/
├── cognito.tf       # Configuração do Cognito (IdP)
├── dynamodb.tf      # Tabelas DynamoDB
├── iam.tf           # Roles e policies
├── main.tf          # Provider AWS
├── s3.tf            # Buckets S3
├── terraform.tf     # Configuração do Terraform
├── variables.tf     # Variáveis de entrada
├── vpc.tf           # Networking
└── README.md        # Esta documentação
```

## Referências

- [AWS Cognito Documentation](https://docs.aws.amazon.com/cognito/)
- [Terraform AWS Provider - Cognito](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cognito_user_pool)
- [JWT.io](https://jwt.io/) - Debug de tokens JWT
