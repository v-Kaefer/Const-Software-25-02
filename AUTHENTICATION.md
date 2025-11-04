# Authentication and Authorization Guide

This document provides a comprehensive guide to the JWT authentication and RBAC authorization system implemented in this project.

## Overview

The User Service API now implements:
- **JWT (JSON Web Tokens)** for authentication
- **RBAC (Role-Based Access Control)** for authorization
- **AWS Cognito** as the Identity Provider (IdP)

## Architecture

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   Client    │─────▶│  API Gateway │─────▶│  User API   │
└─────────────┘      └──────────────┘      └─────────────┘
       │                                           │
       │ 1. Login                                  │
       ▼                                           ▼
┌─────────────┐                          ┌─────────────┐
│   Cognito   │                          │ JWT Verify  │
│  (AWS IdP)  │                          │  + RBAC     │
└─────────────┘                          └─────────────┘
       │                                           │
       │ 2. JWT Token                              │
       └───────────────────────────────────────────┘
              3. Include in Authorization header
```

## Quick Start

### 1. Configure Environment Variables

```bash
# Copy the example
cp .env.example .env

# Edit and configure
JWT_ISSUER=https://cognito-idp.us-east-1.amazonaws.com/us-east-1_XXXXXXXXX
JWT_AUDIENCE=your-app-client-id
JWKS_URI=https://cognito-idp.us-east-1.amazonaws.com/us-east-1_XXXXXXXXX/.well-known/jwks.json
```

### 2. Deploy Infrastructure

**For Production:**
```bash
cd infra
terraform init
terraform apply
# Note the outputs: jwt_issuer, jwks_uri, cognito_client_id
```

**For Local Development (Localstack):**
```bash
# Start Localstack
localstack start

# Deploy to Localstack
cd infra-localstack
terraform init
terraform apply
```

### 3. Create Users

```bash
# Create an admin user
aws cognito-idp admin-create-user \
  --user-pool-id <USER_POOL_ID> \
  --username admin@example.com \
  --user-attributes Name=email,Value=admin@example.com Name=name,Value="Admin User" \
  --temporary-password "TempPass123!" \
  --message-action SUPPRESS

# Add to admin group
aws cognito-idp admin-add-user-to-group \
  --user-pool-id <USER_POOL_ID> \
  --username admin@example.com \
  --group-name admin-group
```

### 4. Obtain a Token

```bash
# Authenticate and get token
TOKEN=$(aws cognito-idp initiate-auth \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id <CLIENT_ID> \
  --auth-parameters USERNAME=admin@example.com,PASSWORD=YourPassword123! \
  | jq -r '.AuthenticationResult.IdToken')
```

### 5. Make Authenticated Requests

```bash
# List users (admin only)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/users

# Get specific user
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/users/123

# Create user
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"new@example.com","name":"New User"}' \
  http://localhost:8080/users
```

## JWT Token Structure

The JWT token contains the following claims:

```json
{
  "sub": "user-id-123",
  "email": "user@example.com",
  "cognito:groups": ["admin-group"],
  "iss": "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_XXXXX",
  "aud": "your-client-id",
  "exp": 1699999999,
  "nbf": 1699900000,
  "iat": 1699900000
}
```

### Key Claims:
- `sub`: Subject (user ID)
- `cognito:groups`: User's roles/groups
- `iss`: Issuer (must match JWT_ISSUER)
- `aud`: Audience (must match JWT_AUDIENCE)
- `exp`: Expiration timestamp
- `nbf`: Not before timestamp

## RBAC Rules

### Roles

- **admin-group**: Full administrative access
- **user-group**: Limited access to own resources

### Endpoint Permissions

| Endpoint | Method | Permission Required |
|----------|--------|-------------------|
| `/users` | GET | Admin only |
| `/users` | POST | Any authenticated user |
| `/users/{id}` | GET | Owner or Admin |
| `/users/{id}` | PUT | Owner or Admin |
| `/users/{id}` | PATCH | Owner or Admin |
| `/users/{id}` | DELETE | Admin only |

### Permission Logic

1. **Admin Access**: Users in `admin-group` have full access to all resources
2. **Owner Access**: Users can access/modify their own resources (when `sub` matches resource ID)
3. **Forbidden**: All other combinations return 403 Forbidden

## Token Validation

The middleware validates:

1. **Presence**: Token must exist in `Authorization: Bearer <token>` header
2. **Format**: Must be a valid JWT structure
3. **Signature**: Verified using JWKS from `JWKS_URI`
4. **Issuer**: Must match `JWT_ISSUER`
5. **Audience**: Must match `JWT_AUDIENCE`
6. **Expiration**: Token must not be expired
7. **Not Before**: Token must be valid (nbf < now)

## Security Best Practices

### Token Storage
- ✅ Store tokens in memory or secure storage
- ❌ Never store tokens in localStorage (XSS risk)
- ✅ Use httpOnly cookies when possible

### Token Transmission
- ✅ Always use HTTPS in production
- ✅ Include token in Authorization header
- ❌ Never pass tokens in URL query parameters

### Token Lifecycle
- Tokens expire after 60 minutes (configurable)
- Use refresh tokens for long-lived sessions
- Implement token revocation for logout

### Rate Limiting
- Implement rate limiting per user/IP
- Monitor for brute force attacks
- Log authentication failures

## Development & Testing

### Testing Without Authentication

For local development and testing, you can run the API without JWT:

```bash
# Don't set JWT_* environment variables
unset JWT_ISSUER JWT_AUDIENCE JWKS_URI

# Run the API
go run ./cmd/api
```

The API will log a warning and allow all requests without authentication.

### Mock Tokens

For integration tests, create mock tokens:

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "time"
)

func createMockToken() string {
    claims := jwt.MapClaims{
        "sub": "test-user-123",
        "cognito:groups": []string{"admin-group"},
        "iss": "http://localhost:4566",
        "aud": "test-client-id",
        "exp": time.Now().Add(1 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte("test-secret"))
    return tokenString
}
```

### Unit Tests

Run tests with:

```bash
# All tests
go test ./...

# Auth tests only
go test ./internal/auth/...

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Troubleshooting

### "invalid token" Error

**Causes:**
- Token signature verification failed
- Token expired
- Wrong issuer or audience
- JWKS not accessible

**Solutions:**
1. Check that `JWT_ISSUER` matches token's `iss` claim
2. Verify `JWT_AUDIENCE` matches token's `aud` claim
3. Ensure `JWKS_URI` is accessible
4. Check token expiration with https://jwt.io

### "forbidden: admin access required" Error

**Cause:** User doesn't have admin role

**Solution:** Add user to admin-group:
```bash
aws cognito-idp admin-add-user-to-group \
  --user-pool-id <USER_POOL_ID> \
  --username user@example.com \
  --group-name admin-group
```

### JWKS Fetch Errors

**Causes:**
- Network connectivity issues
- Invalid JWKS URI
- Cognito service down

**Solutions:**
1. Test JWKS URI manually: `curl $JWKS_URI`
2. Check network/firewall rules
3. Verify Cognito configuration

## Additional Resources

- [AWS Cognito Documentation](https://docs.aws.amazon.com/cognito/)
- [JWT.io - Debug Tokens](https://jwt.io/)
- [OpenAPI Specification](./openapi.yaml)
- [Infrastructure Guide](./infra/README.md)
- [Localstack Guide](./infra-localstack/README.md)

## Support

For issues or questions:
1. Check this documentation
2. Review the OpenAPI spec
3. Check application logs
4. Open an issue on GitHub
