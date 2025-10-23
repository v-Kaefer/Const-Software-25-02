# 🔐 Alternativas ao Cognito para Testes Locais

Como o AWS Cognito não está disponível no LocalStack free tier, este guia apresenta alternativas para desenvolvimento e testes locais.

## 📊 Comparação de Opções

| Opção | Custo | Compatibilidade | Complexidade | Recomendado para |
|-------|-------|-----------------|--------------|------------------|
| LocalStack Pro | Pago ($$ 💰) | Alta (oficial) | Baixa | Produção-like |
| cognito-local | Grátis ✅ | Média | Média | Desenvolvimento |
| Mock/Fake | Grátis ✅ | Baixa | Baixa | Testes unitários |
| AWS Free Tier | Grátis* ✅ | Alta (real) | Baixa | Testes limitados |

*AWS Free Tier tem limites mensais

## 1️⃣ LocalStack Pro (Recomendado)

### Vantagens
- ✅ Compatibilidade completa com AWS Cognito
- ✅ Funciona com a configuração Terraform existente
- ✅ Fácil de usar
- ✅ Suporte oficial

### Desvantagens
- ❌ Requer licença paga
- ❌ ~$30-40/mês (preço pode variar)

### Como usar
```bash
# 1. Obter API Key em https://app.localstack.cloud/
export LOCALSTACK_API_KEY=sua-chave

# 2. Usar normalmente
make infra-up
```

## 2️⃣ cognito-local (Alternativa Open Source)

[cognito-local](https://github.com/jagregory/cognito-local) é um emulador open-source do AWS Cognito.

### Vantagens
- ✅ Gratuito e open source
- ✅ Suporta muitas operações do Cognito
- ✅ Roda em Docker
- ✅ API compatível com AWS SDK

### Desvantagens
- ❌ Não suporta 100% das features do Cognito
- ❌ Requer modificação do endpoint nos clientes
- ❌ Não funciona diretamente com Terraform AWS provider

### Instalação e Uso

#### Passo 1: Adicionar ao docker-compose.yaml

Crie um novo arquivo `docker-compose.cognito-local.yaml`:

```yaml
version: '3.8'

services:
  cognito-local:
    image: jagregory/cognito-local:latest
    container_name: cognito-local
    ports:
      - "9229:9229"
    volumes:
      - ./cognito-local-config:/app/.cognito
    environment:
      - COGNITO_LOCAL_DATABASE=/app/.cognito/db.json
```

#### Passo 2: Iniciar
```bash
docker-compose -f docker-compose.cognito-local.yaml up -d
```

#### Passo 3: Configurar manualmente via AWS CLI

```bash
# Configurar endpoint
export AWS_ENDPOINT=http://localhost:9229

# Criar User Pool
aws cognito-idp create-user-pool \
  --pool-name TestUserPool \
  --endpoint-url $AWS_ENDPOINT

# Criar usuário
aws cognito-idp admin-create-user \
  --user-pool-id us-east-1_xxxxxx \
  --username user@example.com \
  --endpoint-url $AWS_ENDPOINT
```

#### Passo 4: Usar na aplicação

```go
// Configurar SDK para usar cognito-local
cfg, err := config.LoadDefaultConfig(context.TODO(),
    config.WithEndpointResolver(aws.EndpointResolverFunc(
        func(service, region string) (aws.Endpoint, error) {
            if service == cognitoidentityprovider.ServiceID {
                return aws.Endpoint{
                    URL: "http://localhost:9229",
                }, nil
            }
            return aws.Endpoint{}, &aws.EndpointNotFoundError{}
        },
    )),
)
```

## 3️⃣ Mocks para Testes Unitários

Para testes unitários, use mocks ao invés de um Cognito real.

### Exemplo com Go

```go
// Criar interface
type CognitoClient interface {
    SignUp(context.Context, *cognitoidentityprovider.SignUpInput) (*cognitoidentityprovider.SignUpOutput, error)
    InitiateAuth(context.Context, *cognitoidentityprovider.InitiateAuthInput) (*cognitoidentityprovider.InitiateAuthOutput, error)
}

// Mock para testes
type MockCognitoClient struct {
    mock.Mock
}

func (m *MockCognitoClient) SignUp(ctx context.Context, input *cognitoidentityprovider.SignUpInput) (*cognitoidentityprovider.SignUpOutput, error) {
    args := m.Called(ctx, input)
    return args.Get(0).(*cognitoidentityprovider.SignUpOutput), args.Error(1)
}

// Usar nos testes
func TestUserSignup(t *testing.T) {
    mockClient := new(MockCognitoClient)
    mockClient.On("SignUp", mock.Anything, mock.Anything).Return(&cognitoidentityprovider.SignUpOutput{
        UserConfirmed: false,
    }, nil)
    
    // Testar sua lógica...
}
```

### Bibliotecas úteis
- [testify/mock](https://github.com/stretchr/testify) - Para criar mocks
- [go-mock](https://github.com/golang/mock) - Geração automática de mocks

## 4️⃣ AWS Free Tier

Para testes mais realistas, use o próprio AWS Cognito no free tier.

### Vantagens
- ✅ Cognito real da AWS
- ✅ Gratuito dentro dos limites
- ✅ Perfeito para homologação

### Desvantagens
- ❌ Requer conta AWS
- ❌ Limites do free tier (50.000 MAUs)
- ❌ Pode gerar custos se exceder limites
- ❌ Mais lento que opções locais

### Free Tier Limits
- 50.000 MAUs (Monthly Active Users) gratuitos
- Operações adicionais fora do MAU

### Como usar
```bash
# 1. Configurar credenciais AWS
aws configure

# 2. Usar diretório infra/ ao invés de infra-localstack/
cd infra
terraform init
terraform apply
```

## 5️⃣ Solução Híbrida (Recomendada para Free Tier)

Para quem tem LocalStack free tier, uma boa estratégia é:

### Durante Desenvolvimento
1. **LocalStack** para S3 e DynamoDB (gratuito)
2. **Mocks** para Cognito nos testes unitários
3. **Autenticação simplificada** ou JWT manual para dev

### Configuração

```bash
# 1. Desabilitar Cognito no Terraform
cd infra-localstack
mv cognito.tf cognito.tf.disabled

# 2. Iniciar LocalStack sem Cognito
make infra-up

# 3. Para autenticação local, usar JWT simples
# Exemplo de geração de token JWT para dev:
```

```go
// auth_dev.go
package auth

import (
    "github.com/golang-jwt/jwt/v5"
    "time"
)

// Para desenvolvimento local apenas!
func GenerateDevToken(userID, email string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":   userID,
        "email": email,
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
    })
    
    return token.SignedString([]byte("dev-secret-key"))
}
```

### Para Homologação/Staging
- Usar **AWS Cognito real** no free tier
- Aplicar o Terraform do diretório `infra/`

## 📝 Resumo de Recomendações

### Para quem tem orçamento
→ **LocalStack Pro** - Vale o investimento para ambiente development completo

### Para desenvolvimento open-source/pessoal
→ **Solução Híbrida**: LocalStack free + Mocks + AWS free tier para staging

### Para CI/CD
→ **Mocks** nos testes unitários, AWS real em staging

### Para aprender/estudar
→ **cognito-local** - Ótimo para entender como Cognito funciona

## 🔗 Links Úteis

- [LocalStack Pricing](https://localstack.cloud/pricing/)
- [cognito-local GitHub](https://github.com/jagregory/cognito-local)
- [AWS Cognito Free Tier](https://aws.amazon.com/cognito/pricing/)
- [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/)
- [testify - Testing toolkit](https://github.com/stretchr/testify)

## ❓ Qual escolher?

```
┌─────────────────────────────────────────────────────┐
│ Você tem orçamento para LocalStack Pro?            │
└────────────┬─────────────────────────┬──────────────┘
             │                         │
          ✅ Sim                    ❌ Não
             │                         │
             v                         v
    ┌─────────────────┐    ┌──────────────────────┐
    │ LocalStack Pro  │    │ Precisa testar       │
    │                 │    │ Cognito localmente?  │
    └─────────────────┘    └──────┬──────┬────────┘
                                   │      │
                                ✅ Sim  ❌ Não
                                   │      │
                                   v      v
                        ┌──────────────┐ ┌──────────────┐
                        │cognito-local │ │ Mocks +      │
                        │ou AWS free   │ │ LocalStack   │
                        │tier          │ │ free tier    │
                        └──────────────┘ └──────────────┘
```

## 💡 Dica Final

Para este projeto (free tier), recomendamos:
1. **Desenvolvimento**: LocalStack free (sem Cognito) + Mocks
2. **Testes**: Mocks + testes de integração mínimos
3. **Staging**: AWS Cognito no free tier (use com cuidado para não exceder limites)
4. **Documentação**: Manter cognito.tf para referência futura
