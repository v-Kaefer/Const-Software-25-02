# JWT Authentication Examples

This directory contains practical examples demonstrating JWT authentication with the Terraform-provisioned AWS Cognito infrastructure.

## Available Examples

### 1. jwt-auth-example.go

A complete example demonstrating:
- User authentication with Cognito
- JWT token retrieval (ID, Access, and Refresh tokens)
- Token refresh mechanism
- User information retrieval using access tokens

**Prerequisites:**
```bash
# Install AWS SDK dependencies (only needed for JWT examples)
go get github.com/aws/aws-sdk-go-v2/aws@latest
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider@latest

# Start cognito-local
make cognito-local-start

# Setup Cognito infrastructure
make cognito-local-setup
```

**Run the example:**
```bash
# Install dependencies first (if not already installed)
go get github.com/aws/aws-sdk-go-v2/aws@latest
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider@latest

# Run the example
go run examples/jwt-auth-example.go
```

**What it does:**
1. Loads Cognito configuration from `infra-localstack/cognito-local-config/config.json`
2. Authenticates user with username/password
3. Receives JWT tokens from Cognito
4. Displays token information and saves to `tokens.json`
5. Retrieves user information using the access token
6. Demonstrates token refresh

**Expected output:**
```
🚀 JWT Authentication Example with Terraform Cognito
====================================================

1️⃣  Logging in as: user@example.com
🔐 Authenticating user with Cognito...

✅ Authentication Successful!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Token Type: Bearer
Expires In: 3600 seconds (60 minutes)

📝 JWT Tokens:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
...
```

## JWT Token Structure

The example demonstrates the three types of JWT tokens issued by Cognito:

### ID Token
Contains user identity and attributes:
- User ID (`sub`)
- Email address
- Username
- Custom attributes (name, role)
- Group memberships

### Access Token
Used to authorize API requests:
- Scopes and permissions
- Client ID
- Token expiration

### Refresh Token
Used to obtain new ID and Access tokens:
- Long-lived (typically 30 days)
- Can be revoked
- Single-use (new refresh token issued with each refresh)

## Using JWT Tokens in Your Application

### 1. Protect API Routes

```go
import (
    "github.com/gin-gonic/gin"
)

func setupRoutes(router *gin.Engine) {
    // Public routes
    router.POST("/auth/login", handleLogin)
    
    // Protected routes (require JWT)
    protected := router.Group("/api")
    protected.Use(JWTAuthMiddleware())
    {
        protected.GET("/profile", getProfile)
        protected.POST("/data", createData)
    }
}
```

### 2. Validate Tokens

See `JWT-WITH-TERRAFORM.md` for complete validation examples including:
- Signature verification using JWKS
- Expiration checking
- Issuer and audience validation
- Claims extraction

### 3. Extract User Information

```go
func getProfile(c *gin.Context) {
    // Middleware sets these from validated JWT
    userID := c.GetString("user_id")
    email := c.GetString("email")
    groups := c.Get("groups")
    
    c.JSON(200, gin.H{
        "user_id": userID,
        "email": email,
        "groups": groups,
    })
}
```

## Testing the Example

### Prerequisites
- Go 1.22+
- Docker and Docker Compose
- AWS CLI (for cognito-local setup)

### Quick Start

```bash
# 1. Clone and navigate to repository
cd /path/to/Const-Software-25-02

# 2. Start cognito-local
make cognito-local-start

# 3. Setup Cognito infrastructure (creates users, groups, etc.)
make cognito-local-setup

# 4. Install dependencies (if running the example)
go get github.com/aws/aws-sdk-go-v2/aws@latest
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider@latest

# 5. Run the example
go run examples/jwt-auth-example.go

# 6. View the generated tokens
cat tokens.json

# 7. Decode tokens at https://jwt.io/
```

### Troubleshooting

**Error: Configuration file not found**
```
Solution: Run 'make cognito-local-setup' to generate config.json
```

**Error: Connection refused**
```
Solution: Ensure cognito-local is running with 'docker ps | grep cognito-local'
```

**Error: Invalid credentials**
```
Solution: Check username/password in the example matches those created during setup
Default: user@example.com / PassTemp123!
```

## Default Test Users

The Terraform setup creates three test users:

| Email | Password | Group |
|-------|----------|-------|
| admin@example.com | AdminTemp123! | admin-group |
| reviewer@example.com | PassTemp123! | reviewers-group |
| user@example.com | PassTemp123! | user-group |

Modify the example to test different users and group-based authorization.

## Next Steps

1. **Implement JWT validation** - See `JWT-WITH-TERRAFORM.md` for complete examples
2. **Create middleware** - Protect your API routes with JWT authentication
3. **Add role-based access control** - Use Cognito groups for authorization
4. **Handle token refresh** - Implement automatic token refresh in your client
5. **Secure token storage** - Use httpOnly cookies or secure storage

## Resources

- [JWT-WITH-TERRAFORM.md](../infra-localstack/JWT-WITH-TERRAFORM.md) - Complete JWT implementation guide
- [cognito.tf](../infra-localstack/cognito.tf) - Terraform configuration for Cognito
- [COGNITO-LOCAL-SETUP.md](../infra-localstack/COGNITO-LOCAL-SETUP.md) - Local testing guide
- [AWS Cognito Documentation](https://docs.aws.amazon.com/cognito/)
- [JWT.io](https://jwt.io/) - JWT decoder and debugger
