# Autenticação RBAC

Este documento descreve a implementação de autenticação baseada em controle de acesso por funções (RBAC - Role-Based Access Control) integrada com AWS Cognito.

## Visão Geral

A API usa AWS Cognito para autenticação e autorização, implementando RBAC através de Grupos de Usuários do Cognito. Tokens JWT são verificados e o controle de acesso baseado em funções é aplicado no nível da API.

## Arquitetura

### Componentes

1. **AWS Cognito User Pool**: Gerencia autenticação de usuários
2. **Cognito User Groups**: Define funções (admin-group, reviewers-group, user-group)
3. **JWT Middleware**: Valida tokens JWT do Cognito
4. **RBAC Middleware**: Aplica controle de acesso baseado em funções

### Funções

Três funções são definidas no sistema:

- `admin-group`: Acesso completo a todos os recursos
- `reviewers-group`: Acesso de leitura aos recursos
- `user-group`: Acesso limitado ao nível de usuário

## Configuração

### Variáveis de Ambiente

Adicione as seguintes variáveis de ambiente ao seu arquivo `.env`:

```bash
COGNITO_REGION=us-east-1
COGNITO_USER_POOL_ID=seu-user-pool-id
```

### Configuração do Cognito

A infraestrutura já está definida em `infra/cognito.tf`. Para implantar:

```bash
# Para testes locais com cognito-local
make cognito-local-start
make cognito-local-setup

# Para produção
cd infra
terraform apply
```

## Uso

### Endpoints Protegidos

Os endpoints podem ser protegidos envolvendo-os com middleware de autenticação e função:

```go
// Requer apenas autenticação
r.mux.Handle("GET /protected", 
    authMiddleware.Authenticate(http.HandlerFunc(handler)))

// Requer autenticação + função específica
r.mux.Handle("POST /admin-only",
    authMiddleware.Authenticate(
        authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(handler))))

// Requer autenticação + qualquer uma das múltiplas funções
r.mux.Handle("GET /resource",
    authMiddleware.Authenticate(
        authMiddleware.RequireRole(auth.RoleAdmin, auth.RoleReviewer)(
            http.HandlerFunc(handler))))
```

### Fazendo Requisições Autenticadas

Inclua um token JWT no header Authorization:

```bash
curl -H "Authorization: Bearer <seu-jwt-token>" \
     http://localhost:8080/users
```

### Obtendo um Token

Para obter um token JWT do Cognito:

```bash
# Usando AWS CLI
aws cognito-idp initiate-auth \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id <seu-client-id> \
  --auth-parameters USERNAME=usuario@example.com,PASSWORD=SuaSenha123! \
  --region us-east-1
```

Ou para testes locais com cognito-local:

```bash
# Veja infra/test-cognito-local.sh para exemplo
make cognito-local-test
```

## Detalhes de Implementação

### Verificação de Token JWT

O middleware executa as seguintes etapas:

1. Extrai o token Bearer do header Authorization
2. Busca o JWKS (JSON Web Key Set) do Cognito
3. Verifica a assinatura do token usando chave pública RSA
4. Valida os claims do token (exp, iss, token_use)
5. Extrai informações do usuário e grupos dos claims

### Extração de Funções

As funções são extraídas do claim `cognito:groups` no token JWT. O token se parece com:

```json
{
  "sub": "user-uuid",
  "cognito:username": "usuario@example.com",
  "cognito:groups": ["admin-group"],
  "email": "usuario@example.com",
  "token_use": "access",
  ...
}
```

### Valores de Contexto

Após autenticação bem-sucedida, os seguintes valores são adicionados ao contexto da requisição:

- **User**: Recuperado via `auth.GetUserFromContext(ctx)`
- **Roles**: Recuperado via `auth.GetRolesFromContext(ctx)`

Exemplo:

```go
func (r *Router) handleProtected(w http.ResponseWriter, req *http.Request) {
    user, _ := auth.GetUserFromContext(req.Context())
    roles, _ := auth.GetRolesFromContext(req.Context())
    
    // Use informações de usuário e funções
    fmt.Printf("Usuário: %s, Funções: %v\n", user, roles)
}
```

## Testes

### Testes Unitários

O middleware de autenticação inclui testes unitários abrangentes:

```bash
go test ./internal/auth/... -v
```

### Middleware Mock

Para testar endpoints que requerem autenticação, use o middleware mock:

```go
mockAuth := auth.NewMockMiddleware()
router := http.NewRouter(userSvc, mockAuth)
```

O middleware mock:
- Ignora validação JWT
- Injeta usuário de teste e função admin no contexto
- Permite testes sem tokens Cognito reais

### Testes de Integração

Para testes de integração com Cognito real:

1. Configure cognito-local ou use um User Pool de teste do Cognito
2. Crie usuários de teste e obtenha tokens reais
3. Teste com tokens JWT reais

## Considerações de Segurança

### Validação de Token

- Tokens são validados contra as chaves públicas do Cognito (JWKS)
- Verificação de assinatura garante autenticidade do token
- Expiração é verificada automaticamente pela biblioteca JWT
- Apenas tokens com claim `token_use` válido são aceitos

### Cache de JWKS

- Chaves públicas são armazenadas em cache por 24 horas
- Reduz chamadas de API para o Cognito
- Atualização automática quando o cache expira

### Tratamento de Erros

- Tokens inválidos retornam 401 Unauthorized
- Falta de funções necessárias retorna 403 Forbidden
- Mensagens de erro detalhadas em desenvolvimento, genéricas em produção

## Solução de Problemas

### "missing authorization header"

Certifique-se de estar incluindo o header Authorization:
```bash
curl -H "Authorization: Bearer TOKEN" ...
```

### "invalid token: ..."

- Verifique se o token não está expirado
- Verifique se COGNITO_USER_POOL_ID corresponde ao emissor do token
- Certifique-se de que o token é do user pool correto

### "insufficient permissions"

Os grupos do usuário não incluem a função necessária. Verifique:
- O usuário está no grupo Cognito correto
- O token inclui o claim `cognito:groups`
- O nome da função corresponde exatamente (ex.: "admin-group")

## Migração de Código Sem Autenticação

Se você tem código existente sem autenticação:

1. Adicione `COGNITO_REGION` e `COGNITO_USER_POOL_ID` ao ambiente
2. Atualize a inicialização do handler para incluir middleware de autenticação
3. Envolva rotas protegidas com autenticação
4. Adicione requisitos de função conforme necessário
5. Atualize testes para usar middleware mock

## Exemplo

Veja `internal/http/handler.go` para um exemplo completo de integração RBAC:

```go
func (r *Router) routes() {
    // Protegido - apenas admin
    r.mux.Handle("POST /users", 
        r.authMiddleware.Authenticate(
            r.authMiddleware.RequireRole(auth.RoleAdmin)(
                http.HandlerFunc(r.handleCreateUser))))
    
    // Endpoint público
    r.mux.HandleFunc("GET /users", r.handleGetUserByEmail)
}
```
