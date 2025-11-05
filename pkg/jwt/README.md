# JWT Authentication Package

Este pacote fornece autenticação JWT (JSON Web Token) para a aplicação.

## Funcionalidades

- Geração de tokens JWT com claims customizados
- Validação de tokens JWT
- Middleware HTTP para proteger endpoints
- Suporte a expiração de tokens
- Hashing seguro de senhas com bcrypt

## Uso

### Gerando um Token

```go
import "github.com/v-Kaefer/Const-Software-25-02/pkg/jwt"
import "time"

// Criar gerador JWT
generator := jwt.NewGenerator("sua-chave-secreta")

// Gerar token válido por 24 horas
token, err := generator.GenerateToken(userID, email, 24*time.Hour)
if err != nil {
    // Tratar erro
}
```

### Validando um Token

```go
claims, err := generator.ValidateToken(tokenString)
if err != nil {
    if err == jwt.ErrExpiredToken {
        // Token expirado
    } else {
        // Token inválido
    }
}

// Usar claims
userID := claims.UserID
email := claims.Email
```

### Protegendo Endpoints com Middleware

```go
// Criar handler protegido
protectedHandler := generator.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Extrair claims do contexto
    claims, ok := jwt.GetClaimsFromContext(r.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Usar claims
    userID := claims.UserID
    // ...
}))
```

## Endpoints de Autenticação

### Registrar Novo Usuário

```bash
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "name": "Nome do Usuário",
  "password": "senha-segura"
}
```

Resposta (201 Created):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "Nome do Usuário"
}
```

### Login

```bash
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "senha-segura"
}
```

Resposta (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "Nome do Usuário"
  }
}
```

### Usando o Token

Para acessar endpoints protegidos, inclua o token no header `Authorization`:

```bash
POST /users
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "email": "newuser@example.com",
  "name": "Novo Usuário"
}
```

## Configuração

Configure a chave secreta JWT através da variável de ambiente:

```bash
JWT_SECRET=sua-chave-secreta-aqui
```

**IMPORTANTE**: Nunca use a chave padrão em produção. Sempre configure uma chave forte e aleatória.

## Segurança

- Tokens são assinados usando HMAC-SHA256
- Senhas são hasheadas com bcrypt antes de serem armazenadas
- Tokens incluem timestamp de expiração
- Validação automática de expiração e assinatura
- Campo de senha nunca é retornado em respostas JSON

## Testes

Execute os testes do pacote JWT:

```bash
go test -v ./pkg/jwt/...
```

Todos os testes incluem:
- Geração e validação de tokens
- Validação de tokens expirados
- Validação de tokens inválidos
- Middleware de autenticação
- Cenários de erro
