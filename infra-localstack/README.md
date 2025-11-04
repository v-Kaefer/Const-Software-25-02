# Definições Localstack (Para testes de infraestrutura pré deploy)

* Para uma visualização geral da infra definida aqui utilize o comando ```terraform graph```

## Teste local
>Você pode executar o terraform, mesmo sem executar o Localstack, mas vai retornar erros nos serviços: DynamoDB, IAM e VPC.

### Para realizar o teste local é necessário

* [Localstack CLI](https://app.localstack.cloud/getting-started)
* [Docker]()
* [Terraform]()

## 1. Execute o Localstack
> Recomendação: Execute o localstack no terminal ao invés de diretamente no vscode.

Execute: ```localstack start```

## 2. Execute o Terraform
Vá até a pasta ```infra-localstack```

Execute: 
```bash
terraform init
terraform plan
terraform apply
```

## 3. Configure as variáveis JWT

Após o apply, configure as variáveis de ambiente da API:

```bash
# Para Localstack, use:
JWT_ISSUER=http://localhost:4566
JWT_AUDIENCE=<client-id-do-output>
JWKS_URI=http://localhost:4566/.well-known/jwks.json
```

Obtenha o client-id com:
```bash
terraform output cognito_client_id
```

## 4. Obter um token JWT

Use o AWS CLI apontando para Localstack:

```bash
# Autenticar usuário
aws cognito-idp initiate-auth \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id $(terraform output -raw cognito_client_id) \
  --auth-parameters USERNAME=user@example.com,PASSWORD=PassTemp123! \
  --endpoint-url http://localhost:4566

# Extrair o token do resultado (AuthenticationResult.IdToken)
```

## 5. Testar a API

```bash
# Iniciar a API
cd ..
go run ./cmd/api

# Em outro terminal, testar com o token
curl -H "Authorization: Bearer <seu-token>" \
  http://localhost:8080/users?email=test@example.com
```

## 6. Interação visual com a Infraestrutura
Você pode visualizar e interagir com a infraestrutura da mesma forma que a AWS, no [Dashboard do Localstack](https://app.localstack.cloud/inst/default/resources).

## Outputs Disponíveis

- `cognito_user_pool_id` - ID do User Pool
- `cognito_client_id` - ID do App Client (use como JWT_AUDIENCE)
- `jwt_issuer` - Issuer do JWT (use como JWT_ISSUER)
- `jwks_uri` - URI do JWKS (use como JWKS_URI)