# 🔐 Guia Completo: cognito-local

Este guia mostra como usar **cognito-local** para testar a configuração Cognito do Terraform **gratuitamente**, sem precisar do LocalStack Pro.

## 📋 Pré-requisitos

**IMPORTANTE**: Você precisa ter instalado:

1. **Docker** e **Docker Compose**
   ```bash
   docker --version
   docker-compose --version
   ```

2. **AWS CLI** (necessário para configurar o cognito-local)
   ```bash
   # Verificar se está instalado
   aws --version
   
   # Instalar se necessário:
   # Ubuntu/Debian
   sudo apt install awscli
   
   # macOS
   brew install awscli
   
   # Via pip (todas as plataformas)
   pip install awscli
   ```

## 📋 O que é cognito-local?

[cognito-local](https://github.com/jagregory/cognito-local) é um emulador open-source do AWS Cognito que roda localmente em Docker. Ele permite testar funcionalidades do Cognito sem custo.

### ✅ Vantagens
- Gratuito e open source
- Roda em Docker (fácil de configurar)
- **API compatível com AWS CLI** - usa comandos `aws cognito-idp` padrão
- Suporta User Pools, grupos, usuários, autenticação
- Ideal para desenvolvimento e testes

### ⚠️ Limitações
- Não suporta 100% das features do Cognito (mas cobre os casos comuns)
- Não funciona diretamente com Terraform AWS provider
- Requer configuração manual via AWS CLI com `--endpoint-url`

### 🔑 Importante: Credenciais AWS

O cognito-local **não valida credenciais AWS**, mas o AWS CLI as requer. Os scripts usam credenciais dummy automaticamente:
- `AWS_ACCESS_KEY_ID=local`
- `AWS_SECRET_ACCESS_KEY=local`
- `AWS_DEFAULT_REGION=us-east-1`

Você pode usar suas próprias credenciais AWS se preferir - elas não serão validadas pelo cognito-local.

## 🚀 Início Rápido

### Passo 1: Iniciar cognito-local

```bash
# Usando Make (recomendado)
make cognito-local-start

# Ou manualmente
docker-compose -f docker-compose.cognito-local.yaml up -d
```

O cognito-local estará disponível em: `http://localhost:9229`

### Passo 2: Configurar com base no Terraform

Este script cria automaticamente a mesma estrutura definida no `cognito.tf` usando **comandos AWS CLI padrão** com `--endpoint-url`:

```bash
# Usando Make (recomendado)
make cognito-local-setup

# Ou manualmente
cd infra-localstack
./setup-cognito-local.sh
```

O script irá criar:
- ✅ User Pool com políticas de senha
- ✅ App Client para autenticação
- ✅ 3 Grupos (admin-group, reviewers-group, user-group)
- ✅ 3 Usuários de exemplo
- ✅ Associação de usuários aos grupos

### Passo 3: Testar a configuração

```bash
# Usando Make (recomendado)
make cognito-local-test

# Ou manualmente
cd infra-localstack
./test-cognito-local.sh
```

Este script valida que tudo foi criado corretamente.

## 📊 Estrutura Criada

O setup cria a seguinte estrutura (equivalente ao Terraform):

### User Pool
- Nome: `CognitoUserPool`
- Política de senha: 8 caracteres, maiúsculas, minúsculas, números
- Atributos: email, name, role
- MFA: Opcional

### App Client
- Nome: `my-app-client`
- Auth flows: `ALLOW_USER_PASSWORD_AUTH`, `ALLOW_REFRESH_TOKEN_AUTH`

### Grupos
| Grupo | Descrição |
|-------|-----------|
| admin-group | Administradores do sistema |
| reviewers-group | Revisores/Avaliadores |
| user-group | Usuários padrão |

### Usuários Criados
| Email | Senha Temporária | Grupo |
|-------|------------------|-------|
| admin@example.com | AdminTemp123! | admin-group |
| reviewer@example.com | PassTemp123! | reviewers-group |
| user@example.com | PassTemp123! | user-group |

## 🔧 Configuração

### Arquivo de Configuração

Após executar o setup, um arquivo `cognito-local-config/config.json` é criado com:

```json
{
  "userPoolId": "us-east-1_xxxxxxxxx",
  "clientId": "xxxxxxxxxxxxxxxxxxxx",
  "endpoint": "http://localhost:9229",
  "region": "us-east-1"
}
```

Use estes valores na sua aplicação.

### Customizar Usuários

Para criar usuários personalizados, copie `credentials.tf.example` para `credentials.tf` e edite:

```bash
cd infra-localstack
cp credentials.tf.example credentials.tf
# Edite credentials.tf com seus usuários
```

Exemplo de `credentials.tf`:
```hcl
variable "admin_users" {
  default = [
    {
      email = "seu.admin@empresa.com"
      name  = "Seu Admin"
    }
  ]
}

variable "reviewer_users" {
  default = [
    {
      email = "revisor@empresa.com"
      name  = "Revisor"
    }
  ]
}

variable "user_cognito" {
  default = {
    email = "usuario@empresa.com"
    name  = "Usuario Padrão"
  }
}
```

Depois execute novamente:
```bash
make cognito-local-clean  # Limpar configuração anterior
make cognito-local-setup  # Aplicar nova configuração
```

## 💻 Integrando com sua Aplicação Go

### Exemplo: Configurar AWS SDK

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

// CognitoConfig representa a configuração do cognito-local
type CognitoConfig struct {
    UserPoolID string `json:"userPoolId"`
    ClientID   string `json:"clientId"`
    Endpoint   string `json:"endpoint"`
    Region     string `json:"region"`
}

// LoadCognitoConfig carrega a configuração do arquivo
func LoadCognitoConfig(path string) (*CognitoConfig, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg CognitoConfig
    err = json.Unmarshal(data, &cfg)
    return &cfg, err
}

// NewCognitoClient cria um cliente configurado para cognito-local
func NewCognitoClient(ctx context.Context, cognitoConfig *CognitoConfig) (*cognitoidentityprovider.Client, error) {
    // Carregar configuração AWS
    cfg, err := config.LoadDefaultConfig(ctx,
        config.WithRegion(cognitoConfig.Region),
        // Configurar endpoint customizado para cognito-local
        config.WithEndpointResolverWithOptions(
            aws.EndpointResolverWithOptionsFunc(
                func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                    if service == cognitoidentityprovider.ServiceID {
                        return aws.Endpoint{
                            URL:           cognitoConfig.Endpoint,
                            SigningRegion: cognitoConfig.Region,
                        }, nil
                    }
                    return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
                },
            ),
        ),
    )
    if err != nil {
        return nil, err
    }

    return cognitoidentityprovider.NewFromConfig(cfg), nil
}

func main() {
    ctx := context.Background()

    // Carregar configuração
    cognitoConfig, err := LoadCognitoConfig("infra-localstack/cognito-local-config/config.json")
    if err != nil {
        log.Fatalf("Erro ao carregar configuração: %v", err)
    }

    // Criar cliente Cognito
    client, err := NewCognitoClient(ctx, cognitoConfig)
    if err != nil {
        log.Fatalf("Erro ao criar cliente: %v", err)
    }

    // Exemplo: Listar usuários
    result, err := client.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
        UserPoolId: aws.String(cognitoConfig.UserPoolID),
    })
    if err != nil {
        log.Fatalf("Erro ao listar usuários: %v", err)
    }

    fmt.Printf("Usuários encontrados: %d\n", len(result.Users))
    for _, user := range result.Users {
        fmt.Printf("- %s\n", *user.Username)
    }
}
```

### Exemplo: Autenticar Usuário

```go
func AuthenticateUser(ctx context.Context, client *cognitoidentityprovider.Client, userPoolID, clientID, username, password string) (*cognitoidentityprovider.InitiateAuthOutput, error) {
    return client.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
        AuthFlow: types.AuthFlowTypeUserPasswordAuth,
        ClientId: aws.String(clientID),
        AuthParameters: map[string]string{
            "USERNAME": username,
            "PASSWORD": password,
        },
    })
}

// Uso
auth, err := AuthenticateUser(ctx, client, cognitoConfig.UserPoolID, cognitoConfig.ClientID, "user@example.com", "PassTemp123!")
if err != nil {
    log.Fatalf("Erro na autenticação: %v", err)
}

fmt.Printf("Token de acesso: %s\n", *auth.AuthenticationResult.AccessToken)
```

## 🧪 Comandos de Teste Manual

### Listar User Pools
```bash
aws cognito-idp list-user-pools \
  --max-results 10 \
  --endpoint-url http://localhost:9229 \
  --region us-east-1
```

### Listar Usuários
```bash
aws cognito-idp list-users \
  --user-pool-id <USER_POOL_ID> \
  --endpoint-url http://localhost:9229 \
  --region us-east-1
```

### Criar Novo Usuário
```bash
aws cognito-idp admin-create-user \
  --user-pool-id <USER_POOL_ID> \
  --username "novo@example.com" \
  --user-attributes "Name=email,Value=novo@example.com" "Name=name,Value=Novo Usuario" \
  --temporary-password "TempPass123!" \
  --message-action SUPPRESS \
  --endpoint-url http://localhost:9229 \
  --region us-east-1
```

### Testar Autenticação
```bash
aws cognito-idp admin-initiate-auth \
  --user-pool-id <USER_POOL_ID> \
  --client-id <CLIENT_ID> \
  --auth-flow ADMIN_NO_SRP_AUTH \
  --auth-parameters "USERNAME=user@example.com,PASSWORD=PassTemp123!" \
  --endpoint-url http://localhost:9229 \
  --region us-east-1
```

## 🔄 Workflow Completo

### Para Desenvolvimento Diário

```bash
# 1. Iniciar cognito-local
make cognito-local-start

# 2. Configurar (primeira vez ou após mudanças)
make cognito-local-setup

# 3. Desenvolver sua aplicação...

# 4. Testar quando necessário
make cognito-local-test

# 5. Parar quando terminar
make cognito-local-stop
```

### Para Limpar e Recomeçar

```bash
# Limpar tudo e começar do zero
make cognito-local-clean
make cognito-local-start
make cognito-local-setup
```

## 🐛 Troubleshooting

### Container para logo após iniciar

Se o container `cognito-local` para/encerra logo após iniciar:

```bash
# 1. Verificar logs do container
docker logs cognito-local

# 2. Verificar se o container está rodando
docker ps -a | grep cognito-local

# 3. Se estiver com status "Exited", verificar o erro
docker logs cognito-local

# 4. Causas comuns:
# - Porta 9229 já em uso
# - Problema com volume/permissões

# 5. Solução: Limpar e reiniciar
make cognito-local-clean
make cognito-local-start

# 6. Verificar status novamente
docker ps | grep cognito-local
```

**Se o problema persistir:**
```bash
# Executar manualmente para ver erros em tempo real
docker-compose -f docker-compose.cognito-local.yaml up

# Ou ver logs continuamente
docker-compose -f docker-compose.cognito-local.yaml logs -f
```

### Erro: "Connection refused"
```bash
# Verificar se o container está rodando
docker ps | grep cognito-local

# Se não estiver, ver logs
docker logs cognito-local

# Ver logs do compose
docker-compose -f docker-compose.cognito-local.yaml logs

# Reiniciar
make cognito-local-stop
make cognito-local-start
```

### Erro: "User pool already exists" ou conflito com Terraform

**Problema**: O script falha ao criar User Pool porque já existe um com o mesmo nome.

**Causa**: O nome "CognitoUserPool" pode conflitar com:
- User Pool criado por execução anterior do script
- User Pool criado por Terraform (se você usou LocalStack Pro)
- Dados persistentes do cognito-local

**Solução Automática** (implementada no script):
O script agora detecta e remove automaticamente pools existentes antes de criar novos.

**Solução Manual** (se necessário):
```bash
# Limpar e reconfigurar
make cognito-local-clean
make cognito-local-start
make cognito-local-setup
```

**Importante**: O script de setup (`setup-cognito-local.sh`) é independente do Terraform. Ele cria recursos no cognito-local usando AWS CLI, não usando Terraform. Isso significa:
- ✅ Não conflita com arquivos `.tf`
- ✅ Pode rodar mesmo sem Terraform instalado
- ✅ Limpa automaticamente pools existentes
- ⚠️ Não persiste após `make cognito-local-clean`

### Porta 9229 já em uso
```bash
# Verificar o que está usando a porta
lsof -i :9229

# Ou alterar a porta no docker-compose.cognito-local.yaml
# Mude "9229:9229" para "9230:9229"
```

## 📚 Diferenças do Terraform

| Feature | Terraform (LocalStack Pro) | cognito-local |
|---------|---------------------------|---------------|
| User Pool | ✅ Automático | ⚠️ Via script |
| User Groups | ✅ Automático | ⚠️ Via script |
| Usuários | ✅ Automático | ⚠️ Via script |
| IAM Roles | ✅ Sim | ❌ Não |
| Identity Pool | ✅ Sim | ❌ Não |
| MFA Completo | ✅ Sim | ⚠️ Limitado |

## 🎯 Quando Usar

### ✅ Use cognito-local quando:
- Você tem LocalStack free tier
- Precisa testar autenticação Cognito localmente
- Está desenvolvendo features de login/signup
- Quer testar fluxos de usuários

### ❌ Use LocalStack Pro quando:
- Precisa de 100% compatibilidade com AWS
- Precisa de Identity Pool e IAM roles integrados
- Tem orçamento para ferramentas Pro
- Precisa automatizar tudo com Terraform

## 🔗 Recursos Adicionais

- [cognito-local GitHub](https://github.com/jagregory/cognito-local)
- [AWS Cognito Documentation](https://docs.aws.amazon.com/cognito/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)

## ✅ Resumo dos Comandos

```bash
# Iniciar
make cognito-local-start

# Configurar (baseado no Terraform)
make cognito-local-setup

# Testar
make cognito-local-test

# Parar
make cognito-local-stop

# Limpar tudo
make cognito-local-clean
```

---

**Pronto!** Agora você pode testar sua infraestrutura Cognito gratuitamente usando cognito-local! 🎉
