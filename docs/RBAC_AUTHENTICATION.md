# RBAC Authentication

This document describes the Role-Based Access Control (RBAC) authentication implementation integrated with AWS Cognito.

## Overview

The API uses AWS Cognito for authentication and authorization, implementing RBAC through Cognito User Groups. JWT tokens are verified and role-based access control is enforced at the API level.

## Architecture

### Components

1. **AWS Cognito User Pool**: Manages user authentication
2. **Cognito User Groups**: Define roles (admin-group, reviewers-group, user-group)
3. **JWT Middleware**: Validates Cognito JWT tokens
4. **RBAC Middleware**: Enforces role-based access control

### Roles

Three roles are defined in the system:

- `admin-group`: Full access to all resources
- `reviewers-group`: Read access to resources
- `user-group`: Limited user-level access

## Configuration

### Environment Variables

Add the following environment variables to your `.env` file:

```bash
COGNITO_REGION=us-east-1
COGNITO_USER_POOL_ID=your-user-pool-id
```

### Cognito Setup

The infrastructure is already defined in `infra/cognito.tf`. To deploy:

```bash
# For local testing with cognito-local
make cognito-local-start
make cognito-local-setup

# For production
cd infra
terraform apply
```

## Usage

### Protected Endpoints

Endpoints can be protected by wrapping them with authentication and role middleware:

```go
// Require authentication only
r.mux.Handle("GET /protected", 
    authMiddleware.Authenticate(http.HandlerFunc(handler)))

// Require authentication + specific role
r.mux.Handle("POST /admin-only",
    authMiddleware.Authenticate(
        authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(handler))))

// Require authentication + any of multiple roles
r.mux.Handle("GET /resource",
    authMiddleware.Authenticate(
        authMiddleware.RequireRole(auth.RoleAdmin, auth.RoleReviewer)(
            http.HandlerFunc(handler))))
```

### Making Authenticated Requests

Include a JWT token in the Authorization header:

```bash
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/users
```

### Getting a Token

To obtain a JWT token from Cognito:

```bash
# Using AWS CLI
aws cognito-idp initiate-auth \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id <your-client-id> \
  --auth-parameters USERNAME=user@example.com,PASSWORD=YourPassword123! \
  --region us-east-1
```

Or for local testing with cognito-local:

```bash
# See infra/test-cognito-local.sh for example
make cognito-local-test
```

## Implementation Details

### JWT Token Verification

The middleware performs the following steps:

1. Extract Bearer token from Authorization header
2. Fetch JWKS (JSON Web Key Set) from Cognito
3. Verify token signature using RSA public key
4. Validate token claims (exp, iss, token_use)
5. Extract user information and groups from claims

### Role Extraction

Roles are extracted from the `cognito:groups` claim in the JWT token. The token looks like:

```json
{
  "sub": "user-uuid",
  "cognito:username": "user@example.com",
  "cognito:groups": ["admin-group"],
  "email": "user@example.com",
  "token_use": "access",
  ...
}
```

### Context Values

After successful authentication, the following values are added to the request context:

- **User**: Retrieved via `auth.GetUserFromContext(ctx)`
- **Roles**: Retrieved via `auth.GetRolesFromContext(ctx)`

Example:

```go
func (r *Router) handleProtected(w http.ResponseWriter, req *http.Request) {
    user, _ := auth.GetUserFromContext(req.Context())
    roles, _ := auth.GetRolesFromContext(req.Context())
    
    // Use user and roles information
    fmt.Printf("User: %s, Roles: %v\n", user, roles)
}
```

## Testing

### Unit Tests

The auth middleware includes comprehensive unit tests:

```bash
go test ./internal/auth/... -v
```

### Mock Middleware

For testing endpoints that require authentication, use the mock middleware:

```go
mockAuth := auth.NewMockMiddleware()
router := http.NewRouter(userSvc, mockAuth)
```

The mock middleware:
- Skips JWT validation
- Injects test user and admin role into context
- Allows testing without real Cognito tokens

### Integration Testing

For integration tests with real Cognito:

1. Set up cognito-local or use a test Cognito User Pool
2. Create test users and obtain real tokens
3. Test with actual JWT tokens

## Security Considerations

### Token Validation

- Tokens are validated against Cognito's public keys (JWKS)
- Signature verification ensures token authenticity
- Expiration is checked automatically by the JWT library
- Only tokens with valid `token_use` claim are accepted

### JWKS Caching

- Public keys are cached for 24 hours
- Reduces API calls to Cognito
- Automatic refresh when cache expires

### Error Handling

- Invalid tokens return 401 Unauthorized
- Missing required roles return 403 Forbidden
- Detailed error messages in development, generic in production

## Troubleshooting

### "missing authorization header"

Ensure you're including the Authorization header:
```bash
curl -H "Authorization: Bearer TOKEN" ...
```

### "invalid token: ..."

- Check token is not expired
- Verify COGNITO_USER_POOL_ID matches the token issuer
- Ensure token is from the correct user pool

### "insufficient permissions"

User's groups don't include required role. Check:
- User is in the correct Cognito group
- Token includes `cognito:groups` claim
- Role name matches exactly (e.g., "admin-group")

## Migration from Non-Auth

If you have existing code without authentication:

1. Add `COGNITO_REGION` and `COGNITO_USER_POOL_ID` to environment
2. Update handler initialization to include auth middleware
3. Wrap protected routes with authentication
4. Add role requirements as needed
5. Update tests to use mock middleware

## Example

See `internal/http/handler.go` for a complete example of integrating RBAC:

```go
func (r *Router) routes() {
    // Protected - admin only
    r.mux.Handle("POST /users", 
        r.authMiddleware.Authenticate(
            r.authMiddleware.RequireRole(auth.RoleAdmin)(
                http.HandlerFunc(r.handleCreateUser))))
    
    // Public endpoint
    r.mux.HandleFunc("GET /users", r.handleGetUserByEmail)
}
```
