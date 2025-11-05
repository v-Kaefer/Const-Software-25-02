# 🔐 JWT Implementation with Terraform and AWS Cognito

This guide demonstrates how the Terraform configuration in this repository implements JWT (JSON Web Tokens) authentication using AWS Cognito.

## 📋 Table of Contents
1. [Overview](#overview)
2. [How Terraform Provisions JWT Infrastructure](#how-terraform-provisions-jwt-infrastructure)
3. [JWT Token Flow](#jwt-token-flow)
4. [JWT Token Structure](#jwt-token-structure)
5. [Code Examples](#code-examples)
6. [Testing JWT Locally](#testing-jwt-locally)
7. [Validation and Security](#validation-and-security)

## Overview

The `cognito.tf` file in this directory provisions a complete JWT authentication infrastructure using AWS Cognito. When users authenticate, Cognito automatically issues JWT tokens that can be used to access protected resources.

### What is JWT?

JWT (JSON Web Token) is an open standard (RFC 7519) for securely transmitting information between parties as a JSON object. It consists of three parts:
- **Header**: Token type and signing algorithm
- **Payload**: Claims (user data, permissions, expiration)
- **Signature**: Cryptographic signature to verify token integrity

### Why Cognito for JWT?

AWS Cognito User Pools automatically:
- ✅ Generate JWT tokens on successful authentication
- ✅ Sign tokens with RSA keys (RS256 algorithm)
- ✅ Provide public keys (JWKS) for token validation
- ✅ Handle token refresh and expiration
- ✅ Include user attributes and groups in token claims

## How Terraform Provisions JWT Infrastructure

### 1. User Pool - JWT Token Issuer

From `cognito.tf` lines 5-83:

```hcl
resource "aws_cognito_user_pool" "cognito_pool" {
  name = "CognitoUserPool"
  
  # JWT tokens will be issued after successful authentication
  mfa_configuration = "OPTIONAL"
  
  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_uppercase = true
  }
  
  # User attributes that will be included in JWT claims
  schema {
    name                = "email"
    attribute_data_type = "String"
    required            = true
  }
  
  schema {
    name                = "role"
    attribute_data_type = "String"
    required            = false
  }
}
```

**What this does for JWT:**
- Creates the authority that issues JWT tokens
- Defines user attributes that become JWT claims
- Sets security policies for authentication

### 2. App Client - Enables Authentication Flows

From `cognito.tf` lines 324-332:

```hcl
resource "aws_cognito_user_pool_client" "client" {
  name         = "my-app-client"
  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  
  explicit_auth_flows = [
    "ALLOW_USER_PASSWORD_AUTH",  # Enables username/password login
    "ALLOW_REFRESH_TOKEN_AUTH"   # Enables JWT refresh
  ]
}
```

**What this does for JWT:**
- Enables password-based authentication that returns JWT tokens
- Allows refreshing expired JWT access tokens using refresh tokens
- Client ID becomes the `aud` (audience) claim in JWT

### 3. User Groups - JWT Authorization

From `cognito.tf` lines 152-174:

```hcl
resource "aws_cognito_user_group" "admin" {
  name         = "admin-group"
  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  precedence   = 1
  role_arn     = aws_iam_role.cognito_admin_group_role.arn
}

resource "aws_cognito_user_group" "reviewer" {
  name         = "reviewers-group"
  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  precedence   = 2
}

resource "aws_cognito_user_group" "main" {
  name         = "user-group"
  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  precedence   = 3
}
```

**What this does for JWT:**
- User groups are included in JWT as `cognito:groups` claim
- Enables role-based access control (RBAC)
- Precedence determines which role applies when user is in multiple groups

### 4. Identity Pool - AWS Credentials from JWT

From `cognito.tf` lines 312-321:

```hcl
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "MyIdentityPool"
  allow_unauthenticated_identities = false
  
  cognito_identity_providers {
    client_id               = aws_cognito_user_pool_client.client.id
    provider_name           = aws_cognito_user_pool.cognito_pool.endpoint
  }
}
```

**What this does for JWT:**
- Exchanges Cognito JWT tokens for temporary AWS credentials
- Enables authenticated users to access AWS services (S3, DynamoDB, etc.)
- Maps JWT claims to IAM roles

## JWT Token Flow

### 1. Authentication Flow

```
┌──────────┐                  ┌─────────────┐                  ┌──────────┐
│          │   1. Login with  │             │   2. Validate    │          │
│  Client  │─────────────────>│   Cognito   │─────────────────>│   User   │
│          │   email/password │  User Pool  │   credentials    │   Pool   │
└──────────┘                  └─────────────┘                  └──────────┘
     │                               │
     │    3. Return JWT Tokens       │
     │<──────────────────────────────┤
     │   - ID Token (user info)      │
     │   - Access Token (API access) │
     │   - Refresh Token (renew)     │
     │                               │
     │    4. API Request with JWT    │
     │──────────────────────────────>│
     │       Authorization: Bearer   │
     │         <ID or Access Token>  │
     │                               │
     │    5. Verify JWT signature    │
     │       & claims                │
     │<──────────────────────────────┤
     │                               │
     │    6. Return protected data   │
     │<──────────────────────────────┤
     └───────────────────────────────┘
```

### 2. Token Types Issued by Cognito

| Token Type | Purpose | Typical Use | Expiration |
|------------|---------|-------------|------------|
| **ID Token** | User identity & attributes | Get user info (email, name, groups) | 1 hour |
| **Access Token** | API authorization | Access protected API endpoints | 1 hour |
| **Refresh Token** | Renew expired tokens | Get new ID & Access tokens | 30 days |

## JWT Token Structure

### Example ID Token from Cognito

When decoded, a Cognito JWT ID Token looks like:

```json
{
  "header": {
    "kid": "abcd1234...",
    "alg": "RS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "email_verified": true,
    "iss": "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_Abc123",
    "cognito:username": "user@example.com",
    "cognito:groups": ["user-group"],
    "origin_jti": "xyz789...",
    "aud": "1234567890abcdefghijklmn",
    "event_id": "event-123",
    "token_use": "id",
    "auth_time": 1699876543,
    "exp": 1699880143,
    "iat": 1699876543,
    "email": "user@example.com",
    "name": "John Doe",
    "custom:role": "user"
  },
  "signature": "..."
}
```

### Key Claims Explained

| Claim | Description | Set By |
|-------|-------------|--------|
| `sub` | Unique user identifier (never changes) | Cognito |
| `iss` | Token issuer (User Pool URL) | Cognito |
| `aud` | Audience (App Client ID) | Terraform `aws_cognito_user_pool_client` |
| `exp` | Expiration timestamp | Cognito (default: 1 hour) |
| `iat` | Issued at timestamp | Cognito |
| `email` | User email | Terraform User Pool schema |
| `cognito:groups` | User groups | Terraform `aws_cognito_user_group` |
| `cognito:username` | Username | Cognito |
| `custom:role` | Custom role attribute | Terraform User Pool schema |

## Code Examples

### Example 1: Authenticate and Get JWT Token (Go)

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
    "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// CognitoConfig loads from cognito-local-config/config.json
type CognitoConfig struct {
    UserPoolID string `json:"userPoolId"`
    ClientID   string `json:"clientId"`
    Endpoint   string `json:"endpoint"`
    Region     string `json:"region"`
}

func main() {
    ctx := context.Background()
    
    // Load config from Terraform output
    cognitoConfig, err := loadCognitoConfig("infra-localstack/cognito-local-config/config.json")
    if err != nil {
        log.Fatalf("Error loading config: %v", err)
    }
    
    // Create Cognito client
    client, err := createCognitoClient(ctx, cognitoConfig)
    if err != nil {
        log.Fatalf("Error creating client: %v", err)
    }
    
    // Authenticate and get JWT tokens
    username := "user@example.com"
    password := "PassTemp123!"
    
    result, err := client.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
        AuthFlow: types.AuthFlowTypeUserPasswordAuth,
        ClientId: aws.String(cognitoConfig.ClientID),
        AuthParameters: map[string]string{
            "USERNAME": username,
            "PASSWORD": password,
        },
    })
    
    if err != nil {
        log.Fatalf("Authentication failed: %v", err)
    }
    
    // JWT Tokens returned
    fmt.Println("✅ Authentication successful!")
    fmt.Println("\n📝 JWT Tokens:")
    fmt.Printf("ID Token (first 50 chars): %s...\n", 
        (*result.AuthenticationResult.IdToken)[:50])
    fmt.Printf("Access Token (first 50 chars): %s...\n", 
        (*result.AuthenticationResult.AccessToken)[:50])
    fmt.Printf("Token Type: %s\n", 
        *result.AuthenticationResult.TokenType)
    fmt.Printf("Expires In: %d seconds\n", 
        result.AuthenticationResult.ExpiresIn)
    
    // Save tokens for later use
    saveTokens(result.AuthenticationResult)
}

func loadCognitoConfig(path string) (*CognitoConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var cfg CognitoConfig
    err = json.Unmarshal(data, &cfg)
    return &cfg, err
}

func createCognitoClient(ctx context.Context, cognitoConfig *CognitoConfig) (*cognitoidentityprovider.Client, error) {
    cfg, err := config.LoadDefaultConfig(ctx,
        config.WithRegion(cognitoConfig.Region),
        config.WithEndpointResolverWithOptions(
            aws.EndpointResolverWithOptionsFunc(
                func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                    if service == cognitoidentityprovider.ServiceID {
                        return aws.Endpoint{
                            URL:           cognitoConfig.Endpoint,
                            SigningRegion: cognitoConfig.Region,
                        }, nil
                    }
                    return aws.Endpoint{}, fmt.Errorf("unknown endpoint")
                },
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    return cognitoidentityprovider.NewFromConfig(cfg), nil
}

func saveTokens(auth *types.AuthenticationResultType) {
    tokens := map[string]string{
        "id_token":      *auth.IdToken,
        "access_token":  *auth.AccessToken,
        "refresh_token": *auth.RefreshToken,
    }
    
    data, _ := json.MarshalIndent(tokens, "", "  ")
    os.WriteFile("tokens.json", data, 0600)
}
```

### Example 2: Validate JWT Token (Go)

```go
package main

import (
    "crypto/rsa"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "math/big"
    "net/http"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// JWKS represents the JSON Web Key Set from Cognito
type JWKS struct {
    Keys []JWK `json:"keys"`
}

type JWK struct {
    Kid string `json:"kid"`
    Alg string `json:"alg"`
    Kty string `json:"kty"`
    Use string `json:"use"`
    N   string `json:"n"`
    E   string `json:"e"`
}

// JWTValidator validates Cognito JWT tokens
type JWTValidator struct {
    userPoolID string
    region     string
    clientID   string
    jwks       *JWKS
}

func NewJWTValidator(userPoolID, region, clientID string) (*JWTValidator, error) {
    validator := &JWTValidator{
        userPoolID: userPoolID,
        region:     region,
        clientID:   clientID,
    }
    
    // Download JWKS (public keys) from Cognito
    err := validator.downloadJWKS()
    if err != nil {
        return nil, err
    }
    
    return validator, nil
}

func (v *JWTValidator) downloadJWKS() error {
    // Cognito JWKS endpoint format
    jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
        v.region, v.userPoolID)
    
    resp, err := http.Get(jwksURL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    
    return json.Unmarshal(body, &v.jwks)
}

func (v *JWTValidator) ValidateToken(tokenString string) (*jwt.Token, error) {
    // Parse token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Verify signing algorithm
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        
        // Get key ID from token header
        kid, ok := token.Header["kid"].(string)
        if !ok {
            return nil, fmt.Errorf("kid header not found")
        }
        
        // Find matching public key
        key := v.findKey(kid)
        if key == nil {
            return nil, fmt.Errorf("public key not found for kid: %s", kid)
        }
        
        // Convert JWK to RSA public key
        return v.jwkToPublicKey(key)
    })
    
    if err != nil {
        return nil, err
    }
    
    // Validate claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token claims")
    }
    
    // Verify issuer (iss)
    expectedIssuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s",
        v.region, v.userPoolID)
    if claims["iss"] != expectedIssuer {
        return nil, fmt.Errorf("invalid issuer")
    }
    
    // Verify audience (aud) - should be client ID for ID tokens
    tokenUse, _ := claims["token_use"].(string)
    if tokenUse == "id" {
        if claims["aud"] != v.clientID {
            return nil, fmt.Errorf("invalid audience")
        }
    }
    
    // Verify expiration
    exp, ok := claims["exp"].(float64)
    if !ok || time.Now().Unix() > int64(exp) {
        return nil, fmt.Errorf("token expired")
    }
    
    return token, nil
}

func (v *JWTValidator) findKey(kid string) *JWK {
    for _, key := range v.jwks.Keys {
        if key.Kid == kid {
            return &key
        }
    }
    return nil
}

func (v *JWTValidator) jwkToPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
    // Decode n (modulus)
    nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
    if err != nil {
        return nil, err
    }
    
    // Decode e (exponent)
    eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
    if err != nil {
        return nil, err
    }
    
    // Convert to big.Int
    n := new(big.Int).SetBytes(nBytes)
    e := new(big.Int).SetBytes(eBytes)
    
    return &rsa.PublicKey{
        N: n,
        E: int(e.Int64()),
    }, nil
}

// Example usage
func main() {
    userPoolID := "us-east-1_Abc123"
    region := "us-east-1"
    clientID := "1234567890abcdefghijklmn"
    
    validator, err := NewJWTValidator(userPoolID, region, clientID)
    if err != nil {
        log.Fatalf("Failed to create validator: %v", err)
    }
    
    // Load token from file or get from request header
    tokenString := "eyJraWQiOiJhYmNk..." // Your JWT token here
    
    token, err := validator.ValidateToken(tokenString)
    if err != nil {
        log.Fatalf("Token validation failed: %v", err)
    }
    
    // Extract user information
    claims := token.Claims.(jwt.MapClaims)
    fmt.Println("✅ Token is valid!")
    fmt.Printf("User: %s\n", claims["email"])
    fmt.Printf("Groups: %v\n", claims["cognito:groups"])
    fmt.Printf("Subject: %s\n", claims["sub"])
}
```

### Example 3: Middleware for Protected Routes (Go Gin)

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware validates JWT tokens from Cognito
func JWTAuthMiddleware(validator *JWTValidator) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Authorization header required",
            })
            c.Abort()
            return
        }
        
        // Remove "Bearer " prefix
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid authorization format",
            })
            c.Abort()
            return
        }
        
        // Validate token
        token, err := validator.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid or expired token",
            })
            c.Abort()
            return
        }
        
        // Extract claims and set in context
        claims := token.Claims.(jwt.MapClaims)
        c.Set("user_id", claims["sub"])
        c.Set("email", claims["email"])
        c.Set("groups", claims["cognito:groups"])
        
        c.Next()
    }
}

// RequireGroup middleware checks if user belongs to specific group
func RequireGroup(requiredGroup string) gin.HandlerFunc {
    return func(c *gin.Context) {
        groupsInterface, exists := c.Get("groups")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "No groups found",
            })
            c.Abort()
            return
        }
        
        groups, ok := groupsInterface.([]interface{})
        if !ok {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Invalid groups format",
            })
            c.Abort()
            return
        }
        
        // Check if required group exists
        hasGroup := false
        for _, group := range groups {
            if groupStr, ok := group.(string); ok && groupStr == requiredGroup {
                hasGroup = true
                break
            }
        }
        
        if !hasGroup {
            c.JSON(http.StatusForbidden, gin.H{
                "error": fmt.Sprintf("Requires %s group", requiredGroup),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Example usage in main.go
func setupRoutes(router *gin.Engine, validator *JWTValidator) {
    // Public routes (no JWT required)
    router.POST("/auth/login", handleLogin)
    
    // Protected routes (JWT required)
    protected := router.Group("/api")
    protected.Use(JWTAuthMiddleware(validator))
    {
        // Any authenticated user
        protected.GET("/profile", getProfile)
        
        // Only admin group
        adminRoutes := protected.Group("/admin")
        adminRoutes.Use(RequireGroup("admin-group"))
        {
            adminRoutes.GET("/users", listAllUsers)
            adminRoutes.DELETE("/users/:id", deleteUser)
        }
        
        // Only reviewer group
        reviewerRoutes := protected.Group("/reviews")
        reviewerRoutes.Use(RequireGroup("reviewers-group"))
        {
            reviewerRoutes.POST("/", createReview)
            reviewerRoutes.PUT("/:id", updateReview)
        }
    }
}
```

## Testing JWT Locally

### Option 1: Using cognito-local

```bash
# Start cognito-local
make cognito-local-start

# Setup infrastructure (creates User Pool, App Client, Groups, Users)
make cognito-local-setup

# Test authentication and JWT generation
make cognito-local-test
```

### Option 2: Manual Testing with curl

```bash
# 1. Get JWT tokens
curl -X POST http://localhost:9229/ \
  -H "Content-Type: application/x-amz-json-1.1" \
  -H "X-Amz-Target: AWSCognitoIdentityProviderService.InitiateAuth" \
  -d '{
    "AuthFlow": "USER_PASSWORD_AUTH",
    "ClientId": "<CLIENT_ID>",
    "AuthParameters": {
      "USERNAME": "user@example.com",
      "PASSWORD": "PassTemp123!"
    }
  }'

# Response includes:
# {
#   "AuthenticationResult": {
#     "IdToken": "eyJraWQiOiJ...",
#     "AccessToken": "eyJraWQiOiJ...",
#     "RefreshToken": "eyJjdHki...",
#     "ExpiresIn": 3600,
#     "TokenType": "Bearer"
#   }
# }

# 2. Use JWT token to access protected endpoint
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer <ID_TOKEN>"
```

### Option 3: Using jwt.io to Decode

1. Copy your JWT token
2. Go to https://jwt.io/
3. Paste the token in the "Encoded" section
4. View decoded header, payload, and verify signature

## Validation and Security

### Security Best Practices

1. **Always Validate JWT Signature**
   ```go
   // DON'T: Parse without validation
   token, _ := jwt.Parse(tokenString, nil) // ❌ Insecure!
   
   // DO: Validate with public key
   token, err := validator.ValidateToken(tokenString) // ✅ Secure
   ```

2. **Check Token Expiration**
   ```go
   claims := token.Claims.(jwt.MapClaims)
   exp, _ := claims["exp"].(float64)
   if time.Now().Unix() > int64(exp) {
       return errors.New("token expired")
   }
   ```

3. **Verify Issuer and Audience**
   ```go
   // Verify token came from your User Pool
   if claims["iss"] != expectedIssuer {
       return errors.New("invalid issuer")
   }
   
   // Verify token is for your app
   if claims["aud"] != clientID {
       return errors.New("invalid audience")
   }
   ```

4. **Use HTTPS in Production**
   - Never send JWT tokens over HTTP
   - Always use HTTPS to prevent token interception

5. **Store Tokens Securely**
   - Use httpOnly cookies for web apps
   - Use secure storage for mobile apps
   - Never store tokens in localStorage (XSS vulnerable)

6. **Implement Token Refresh**
   ```go
   func refreshToken(refreshToken string) (*AuthResult, error) {
       result, err := client.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
           AuthFlow: types.AuthFlowTypeRefreshTokenAuth,
           ClientId: aws.String(clientID),
           AuthParameters: map[string]string{
               "REFRESH_TOKEN": refreshToken,
           },
       })
       return result.AuthenticationResult, err
   }
   ```

### JWT Validation Checklist

- [ ] Verify signature using JWKS public keys
- [ ] Check token expiration (`exp` claim)
- [ ] Validate issuer (`iss` claim)
- [ ] Validate audience (`aud` claim)
- [ ] Check token type (`token_use` claim)
- [ ] Verify token has not been revoked (optional - requires additional logic)
- [ ] Extract and validate required claims (email, groups, etc.)

## Terraform Outputs for JWT Integration

Add these outputs to your `cognito.tf` to make JWT integration easier:

```hcl
# Add at the end of cognito.tf
output "user_pool_id" {
  description = "Cognito User Pool ID for JWT validation"
  value       = aws_cognito_user_pool.cognito_pool.id
}

output "user_pool_endpoint" {
  description = "Cognito User Pool endpoint for JWKS"
  value       = aws_cognito_user_pool.cognito_pool.endpoint
}

output "app_client_id" {
  description = "App Client ID for JWT audience validation"
  value       = aws_cognito_user_pool_client.client.id
}

output "jwks_uri" {
  description = "JWKS URI for JWT signature validation"
  value       = "https://${aws_cognito_user_pool.cognito_pool.endpoint}/.well-known/jwks.json"
}

output "identity_pool_id" {
  description = "Identity Pool ID for AWS credential exchange"
  value       = aws_cognito_identity_pool.main.id
}
```

After `terraform apply`, get these values:
```bash
cd infra-localstack
terraform output user_pool_id
terraform output app_client_id
terraform output jwks_uri
```

## Summary

This Terraform configuration provides a complete JWT authentication infrastructure:

1. **User Pool** issues JWT tokens after successful authentication
2. **App Client** enables authentication flows and sets JWT audience
3. **User Groups** add authorization claims to JWT tokens
4. **Identity Pool** exchanges JWT tokens for AWS credentials

The JWT tokens can be used to:
- Authenticate API requests
- Implement role-based access control (RBAC)
- Access AWS services with temporary credentials
- Build secure single sign-on (SSO) solutions

All token signing, validation, and security are handled automatically by AWS Cognito, making it a robust and production-ready JWT solution.
