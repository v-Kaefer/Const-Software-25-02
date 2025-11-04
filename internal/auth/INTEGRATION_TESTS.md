# Cognito-Local Integration Tests

This directory contains integration tests for AWS Cognito authentication using Localstack.

## Overview

The integration tests verify that the JWT authentication system can work with Cognito by:
1. Setting up a Localstack container with Cognito service
2. Creating user pools, app clients, and user groups
3. Creating test users and authenticating them
4. Obtaining JWT tokens from Cognito
5. Validating the JWT flow with the application's authentication middleware

## Running Integration Tests

### Prerequisites

- Docker must be installed and running
- Go 1.22 or later

### Run All Integration Tests

```bash
# Run integration tests
go test -tags=integration ./internal/auth/... -v

# Run with timeout
go test -tags=integration ./internal/auth/... -v -timeout 5m
```

### Run Specific Integration Tests

```bash
# Run only Cognito integration test
go test -tags=integration ./internal/auth/... -v -run TestCognitoIntegration

# Run only Docker Compose integration test
go test -tags=integration ./internal/auth/... -v -run TestCognitoLocalDockerCompose
```

## Test Descriptions

### TestCognitoIntegration

**Purpose**: Validates the complete Cognito integration flow with an ephemeral Localstack container.

**What it tests**:
- Starts a Localstack container with Cognito service
- Creates a Cognito user pool with proper configuration
- Creates app client with authentication flows
- Creates admin and user groups
- Creates a test user and assigns to admin group
- Authenticates the user and obtains JWT token
- Validates JWT token structure

**Note**: Full Cognito support requires Localstack Pro. The test gracefully handles this limitation and validates that the integration setup is correct.

**Expected behavior**:
- With Localstack free tier: Test skips with informative message about Pro requirement
- With Localstack Pro: Test runs full Cognito flow and validates JWT tokens

### TestCognitoLocalDockerCompose

**Purpose**: Validates integration with a manually running Localstack instance (via docker-compose or standalone).

**What it tests**:
- Checks if Localstack is accessible at `http://localhost:4566`
- Verifies Localstack health endpoint responds
- Validates Cognito service is available
- Lists existing user pools (if any)

**Expected behavior**:
- If Localstack is not running: Test skips gracefully
- If Localstack is running: Test validates connectivity and service availability

**To use with docker-compose**:
```bash
# Start Localstack
localstack start

# In another terminal, run the test
go test -tags=integration ./internal/auth/... -v -run TestCognitoLocalDockerCompose
```

## CI/CD Integration

Integration tests are **not** run automatically in CI/CD pipelines by default because:
1. They require Docker to be available
2. Full Cognito support requires Localstack Pro (paid)
3. They add significant time to test runs

To run integration tests in CI, add the `-tags=integration` flag:

```yaml
- name: Run integration tests
  run: go test -tags=integration ./... -v -timeout 10m
```

## Localstack Pro vs Free

| Feature | Localstack Free | Localstack Pro |
|---------|----------------|----------------|
| Cognito User Pool | ❌ Not implemented | ✅ Full support |
| JWT Token Generation | ❌ Not implemented | ✅ Supported |
| User Authentication | ❌ Not implemented | ✅ Supported |
| JWKS Endpoint | ❌ Not implemented | ✅ Supported |

### Using Localstack Pro

If you have access to Localstack Pro:

1. Set your auth token:
   ```bash
   export LOCALSTACK_AUTH_TOKEN=your-token-here
   ```

2. Start Localstack Pro:
   ```bash
   localstack start
   ```

3. Run integration tests:
   ```bash
   go test -tags=integration ./internal/auth/... -v
   ```

## Manual Testing with Terraform

For manual integration testing with the Terraform infrastructure:

1. Start Localstack:
   ```bash
   localstack start
   ```

2. Deploy infrastructure:
   ```bash
   cd infra-localstack
   terraform init
   terraform apply
   ```

3. Get the outputs:
   ```bash
   terraform output cognito_user_pool_id
   terraform output cognito_client_id
   ```

4. Run manual authentication tests:
   ```bash
   # Authenticate a user
   aws cognito-idp initiate-auth \
     --auth-flow USER_PASSWORD_AUTH \
     --client-id $(terraform output -raw cognito_client_id) \
     --auth-parameters USERNAME=user@example.com,PASSWORD=PassTemp123! \
     --endpoint-url http://localhost:4566
   ```

## Troubleshooting

### "Cannot connect to Docker daemon"

Ensure Docker is running:
```bash
docker ps
```

### "Localstack not available"

Check if Localstack is running:
```bash
curl http://localhost:4566/_localstack/health
```

### "API not yet implemented or pro feature"

This is expected with Localstack free tier. The integration test will skip gracefully.
To run full tests, upgrade to Localstack Pro.

### Test timeout

Increase the timeout if container startup takes longer:
```bash
go test -tags=integration ./internal/auth/... -v -timeout 10m
```

## Additional Resources

- [Localstack Documentation](https://docs.localstack.cloud/)
- [Localstack Cognito Coverage](https://docs.localstack.cloud/references/coverage/coverage_cognito-idp/)
- [AWS Cognito Documentation](https://docs.aws.amazon.com/cognito/)
- [Testcontainers Documentation](https://golang.testcontainers.org/)
