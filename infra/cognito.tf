# Cognito User Pool para autenticação JWT
# Este é o provedor de identidade (IdP) que gerará tokens JWT

resource "aws_cognito_user_pool" "main" {
  name = "user-service-pool"

  # Configurações de MFA
  mfa_configuration = "OPTIONAL"
  software_token_mfa_configuration {
    enabled = true
  }

  # Previne enumeração de usuários
  username_configuration {
    case_sensitive = false
  }

  # Política de senha forte
  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_symbols   = false
    require_uppercase = true
  }

  # Atributos obrigatórios
  schema {
    name                = "email"
    attribute_data_type = "String"
    required            = true
    mutable             = true

    string_attribute_constraints {
      min_length = 1
      max_length = 256
    }
  }

  schema {
    name                = "name"
    attribute_data_type = "String"
    required            = false
    mutable             = true

    string_attribute_constraints {
      min_length = 1
      max_length = 256
    }
  }

  # Atributo customizado para role/grupo
  schema {
    name                     = "role"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false

    string_attribute_constraints {
      min_length = 1
      max_length = 20
    }
  }

  # Recuperação de conta via email
  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  # Auto-verificação de email
  auto_verified_attributes = ["email"]

  # Template de verificação
  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }

  tags = {
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "user-service"
  }
}

# App Client - permite que a aplicação se autentique
resource "aws_cognito_user_pool_client" "main" {
  name         = "user-service-client"
  user_pool_id = aws_cognito_user_pool.main.id

  # Fluxos de autenticação permitidos
  explicit_auth_flows = [
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_SRP_AUTH"
  ]

  # Configuração de tokens
  id_token_validity      = 60  # minutos
  access_token_validity  = 60  # minutos
  refresh_token_validity = 30  # dias

  token_validity_units {
    id_token      = "minutes"
    access_token  = "minutes"
    refresh_token = "days"
  }

  # Previne secret client (para apps públicos como SPA)
  generate_secret = false

  # Callbacks e logout URLs (ajustar conforme necessário)
  # callback_urls = ["http://localhost:3000/callback"]
  # logout_urls   = ["http://localhost:3000/logout"]
}

# Grupos de usuários para RBAC
resource "aws_cognito_user_group" "admin" {
  name         = "admin-group"
  user_pool_id = aws_cognito_user_pool.main.id
  description  = "Administrators with full access"
  precedence   = 1
}

resource "aws_cognito_user_group" "user" {
  name         = "user-group"
  user_pool_id = aws_cognito_user_pool.main.id
  description  = "Regular users with limited access"
  precedence   = 10
}

# Identity Pool (opcional - para acesso a recursos AWS)
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "user-service-identity-pool"
  allow_unauthenticated_identities = false

  cognito_identity_providers {
    client_id               = aws_cognito_user_pool_client.main.id
    provider_name           = aws_cognito_user_pool.main.endpoint
    server_side_token_check = false
  }

  tags = {
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "user-service"
  }
}

# Outputs importantes para configuração da API
output "cognito_user_pool_id" {
  description = "ID do Cognito User Pool"
  value       = aws_cognito_user_pool.main.id
}

output "cognito_user_pool_arn" {
  description = "ARN do Cognito User Pool"
  value       = aws_cognito_user_pool.main.arn
}

output "cognito_user_pool_endpoint" {
  description = "Endpoint do Cognito User Pool"
  value       = aws_cognito_user_pool.main.endpoint
}

output "cognito_client_id" {
  description = "ID do App Client - use como JWT_AUDIENCE"
  value       = aws_cognito_user_pool_client.main.id
}

output "jwt_issuer" {
  description = "JWT Issuer - use como JWT_ISSUER"
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.main.id}"
}

output "jwks_uri" {
  description = "JWKS URI para validação de tokens - use como JWKS_URI"
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.main.id}/.well-known/jwks.json"
}

output "cognito_domain" {
  description = "Domínio do Cognito para Hosted UI (se configurado)"
  value       = "https://${var.cognito_domain_prefix}.auth.${var.aws_region}.amazoncognito.com"
}
