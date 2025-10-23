# 🚀 Guia Rápido: LocalStack com Cognito

Este guia mostra como testar a infraestrutura localmente com LocalStack.

## 📋 Pré-requisitos

1. Instale o LocalStack:
   ```bash
   pip install localstack
   ```

2. Instale o AWS CLI:
   ```bash
   # macOS
   brew install awscli
   
   # Ubuntu/Debian
   sudo apt install awscli
   
   # Ou via pip
   pip install awscli
   ```

3. Instale o Terraform (>= 1.2):
   - Download: https://www.terraform.io/downloads

## 🎯 Opção 1: Com LocalStack Free Tier (Sem Cognito)

### Passo 1: Desabilitar Cognito temporariamente
```bash
cd infra-localstack
mv cognito.tf cognito.tf.disabled
```

### Passo 2: Iniciar a infraestrutura
```bash
cd ..  # Voltar para o diretório raiz
make infra-up
```

### Passo 3: Testar
```bash
make infra-test
```

Você verá:
- ✅ S3 funcionando
- ✅ DynamoDB funcionando
- ❌ Cognito não disponível (esperado no free tier)

### Passo 4: Limpar
```bash
make infra-down
mv infra-localstack/cognito.tf.disabled infra-localstack/cognito.tf
```

## 💎 Opção 2: Com LocalStack Pro (Cognito Completo)

### Passo 1: Obter API Key
1. Acesse https://app.localstack.cloud/
2. Crie uma conta ou faça login
3. Obtenha sua API Key

### Passo 2: Configurar API Key
```bash
export LOCALSTACK_API_KEY=seu-api-key-aqui
```

### Passo 3: Configurar usuários Cognito (opcional)
```bash
cd infra-localstack
cp terraform.tfvars.example terraform.tfvars
# Edite terraform.tfvars com seus usuários
```

### Passo 4: Iniciar tudo
```bash
cd ..  # Voltar para o diretório raiz
make infra-up
```

### Passo 5: Testar Cognito
```bash
make infra-test

# Testar especificamente o Cognito
aws --endpoint-url=http://localhost:4566 \
    cognito-idp list-user-pools --max-results 10

# Listar usuários
aws --endpoint-url=http://localhost:4566 \
    cognito-idp list-users --user-pool-id <pool-id>
```

### Passo 6: Limpar
```bash
make infra-down
```

## 🔧 Opção 3: Comandos Manuais (Sem Makefile)

Se preferir não usar o Makefile:

### Iniciar LocalStack
```bash
localstack start -d
# Aguarde alguns segundos para inicializar
localstack status
```

### Aplicar Terraform
```bash
cd infra-localstack
terraform init
terraform plan
terraform apply
```

### Testar recursos
```bash
# S3
aws --endpoint-url=http://localhost:4566 s3 ls

# DynamoDB
aws --endpoint-url=http://localhost:4566 dynamodb list-tables

# Cognito (apenas com Pro)
aws --endpoint-url=http://localhost:4566 cognito-idp list-user-pools --max-results 10
```

### Destruir
```bash
terraform destroy
cd ..
localstack stop
```

## 🐛 Problemas Comuns

### Erro: "command not found: localstack"
```bash
pip install localstack
# Ou se precisar de permissões
pip install --user localstack
```

### Erro: "connection refused" ao executar terraform
```bash
# Verificar se LocalStack está rodando
localstack status

# Ver logs para diagnóstico
localstack logs

# Verificar saúde
curl http://localhost:4566/_localstack/health
```

### Erro: "Cognito not available" (Free Tier)
Isso é esperado! Cognito requer LocalStack Pro. Use a Opção 1 acima.

### LocalStack não inicia
```bash
# Verificar Docker
docker ps

# Limpar containers antigos
docker ps -a | grep localstack | awk '{print $1}' | xargs docker rm -f

# Tentar novamente
localstack start -d
```

## 📊 Visualizar Infraestrutura

### Dashboard Web (LocalStack Pro)
Acesse: https://app.localstack.cloud/inst/default/resources

### Diagrama Terraform
```bash
cd infra-localstack
terraform graph | dot -Tpng > graph.png
open graph.png  # macOS
xdg-open graph.png  # Linux
```

## 📚 Próximos Passos

1. **Integrar com aplicação Go**: Configure a aplicação para usar os endpoints do LocalStack
2. **Testes automatizados**: Crie scripts de teste que usem a infraestrutura
3. **CI/CD**: Configure GitHub Actions para testar com LocalStack

## 🔗 Links Úteis

- [LocalStack Docs](https://docs.localstack.cloud/)
- [LocalStack Pricing](https://localstack.cloud/pricing/)
- [AWS CLI Reference](https://docs.aws.amazon.com/cli/latest/reference/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Alternative: cognito-local](https://github.com/jagregory/cognito-local)
