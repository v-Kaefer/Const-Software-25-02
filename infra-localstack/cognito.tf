# User user pool -> Pense nele como seu banco de dados de autenticação
# Armazena usuários (emails, senhas, atributos personalizados)
# Gerencia registro, login, recuperação de senha
# Emite tokens JWT após autenticação bem-sucedida
resource "aws_cognito_user_pool" "cognito_pool" {
  name = "CognitoUserPool"

  mfa_configuration = "OPTIONAL"
  software_token_mfa_configuration {
    enabled = true
  }

  # Previne enumeração de usuários
  username_configuration {
    case_sensitive = false
  }

  # Política de senha
  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_symbols   = false
    require_uppercase = true
  }


  # Schema padrão para email
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

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  auto_verified_attributes = ["email"]

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }

  tags = {
    Environment = "test"
    ManagedBy   = "terraform"
  }
}

# Usuário Admin
resource "aws_cognito_user" "admin" {
  for_each = { for user in var.admin_users : user.email => user }
  # The user pool ID for the user pool where the user will be created.
  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  username     = each.value.email

  attributes = {
    email         = each.value.email
    name          = each.value.name
    "custom:role" = "admin" # APENAS INFORMATIVO
  }

  temporary_password = "AdminTemp123!"
  message_action     = "SUPPRESS" # Não envia email
  enabled            = true
}

# Usuário avaliador
resource "aws_cognito_user" "reviewers" {
  for_each = { for user in var.reviewer_users : user.email => user }

  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  username     = each.value.email

  attributes = {
    email         = each.value.email
    name          = each.value.name
    "custom:role" = "reviewer" # APENAS INFORMATIVO
  }

  temporary_password = "PassTemp123!"
  message_action     = "SUPPRESS" # Não envia email
  enabled            = true
}

# Usuário padrão
resource "aws_cognito_user" "main" {
  #for_each = { for user in var.main_users : user.email => user }

  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  username     = var.user_cognito["email"]

  attributes = var.user_cognito

  temporary_password = "PassTemp123!"
  message_action     = "SUPPRESS" # Não envia email
  enabled            = true
}

# IAM Roles
resource "aws_iam_role" "cognito_admin_group_role" {
  name               = "admin-group-role"
  assume_role_policy = data.aws_iam_policy_document.cognito_assume_admin_role.json
}

resource "aws_iam_role" "cognito_reviewer_group_role" {
  name               = "reviewer-group-role"
  assume_role_policy = data.aws_iam_policy_document.cognito_assume_reviewer_role.json
}

resource "aws_iam_role" "cognito_main_group_role" {
  name               = "user-group-role"
  assume_role_policy = data.aws_iam_policy_document.cognito_assume_main_role.json
}

# Cognito User Groups
resource "aws_cognito_user_group" "admin" {
  name         = "admin-group"
  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  description  = "Managed by Terraform"
  precedence   = 1
  role_arn     = aws_iam_role.cognito_admin_group_role.arn
}

resource "aws_cognito_user_group" "reviewer" {
  name         = "reviewers-group"
  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  description  = "Managed by Terraform"
  precedence   = 2
  role_arn     = aws_iam_role.cognito_reviewer_group_role.arn
}

resource "aws_cognito_user_group" "main" {
  name         = "user-group"
  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  description  = "Managed by Terraform"
  precedence   = 3
  role_arn     = aws_iam_role.cognito_main_group_role.arn
}

# Adição dos usuários aos grupos
resource "aws_cognito_user_in_group" "admin_in_admin_group" {
  for_each = aws_cognito_user.admin #Para "each.value" funcionar

  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  group_name   = aws_cognito_user_group.admin.name
  username     = each.value.username
}

resource "aws_cognito_user_in_group" "reviewer_in_reviewer_group" {
  for_each = aws_cognito_user.reviewers #Para "each.value" funcionar

  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  group_name   = aws_cognito_user_group.reviewer.name
  username     = each.value.username
}

resource "aws_cognito_user_in_group" "main_in_main_group" {
  user_pool_id = aws_cognito_user_pool.cognito_pool.id
  group_name   = aws_cognito_user_group.main.name
  username     = aws_cognito_user.main.username
}

# IAM Roles Policies
resource "aws_iam_role_policy" "admin_role_policy" {
  name = "admin-permissions"
  role = aws_iam_role.cognito_admin_group_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:*",
          "dynamodb:*",
          "cognito-idp:*"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy" "reviewer_role_policy" {
  name = "reviewer-permissions"
  role = aws_iam_role.cognito_reviewer_group_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject"]
        Resource = "arn:aws:s3:::meu-bucket/*"
      }
    ]
  })
}

resource "aws_iam_role_policy" "main_role_policy" {
  name = "main-permissions"
  role = aws_iam_role.cognito_main_group_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject"]
        Resource = "arn:aws:s3:::meu-bucket/*"
      }
    ]
  })
}

# IAM Policy Documents
data "aws_iam_policy_document" "cognito_assume_admin_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Federated"
      identifiers = ["cognito-identity.amazonaws.com"]
    }

    actions = ["sts:AssumeRoleWithWebIdentity"]

    condition {
      test     = "StringEquals"
      variable = "cognito-identity.amazonaws.com:aud"
      values   = [aws_cognito_identity_pool.main.id] # ID do Identity Pool
    }
  }
}


data "aws_iam_policy_document" "cognito_assume_reviewer_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Federated"
      identifiers = ["cognito-identity.amazonaws.com"]
    }

    actions = ["sts:AssumeRoleWithWebIdentity"]

    condition {
      test     = "StringEquals"
      variable = "cognito-identity.amazonaws.com:aud"
      values   = [aws_cognito_identity_pool.main.id] # ID do Identity Pool
    }
  }
}

data "aws_iam_policy_document" "cognito_assume_main_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Federated"
      identifiers = ["cognito-identity.amazonaws.com"]
    }

    actions = ["sts:AssumeRoleWithWebIdentity"]

    condition {
      test     = "StringEquals"
      variable = "cognito-identity.amazonaws.com:aud"
      values   = [aws_cognito_identity_pool.main.id] # ID do Identity Pool
    }
  }
}

#Identity Pool
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "MyIdentityPool"
  allow_unauthenticated_identities = false

  cognito_identity_providers {
    client_id               = aws_cognito_user_pool_client.client.id
    provider_name           = aws_cognito_user_pool.cognito_pool.endpoint
    server_side_token_check = false
  }
}

#User Pool Client (App Client)
resource "aws_cognito_user_pool_client" "client" {
  name         = "my-app-client"
  user_pool_id = aws_cognito_user_pool.cognito_pool.id

  explicit_auth_flows = [
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH"
  ]
}

#Mapeamento Identity Pool
resource "aws_cognito_identity_pool_roles_attachment" "main" {
  identity_pool_id = aws_cognito_identity_pool.main.id

  roles = {
    "authenticated" = aws_iam_role.cognito_main_group_role.arn # Role padrão
  }

  role_mapping {
    identity_provider         = "${aws_cognito_user_pool.cognito_pool.endpoint}:${aws_cognito_user_pool_client.client.id}"
    ambiguous_role_resolution = "AuthenticatedRole"
    type                      = "Token"

    mapping_rule {
      claim      = "cognito:groups"
      match_type = "Contains"
      value      = "admin-group"
      role_arn   = aws_iam_role.cognito_admin_group_role.arn
    }

    mapping_rule {
      claim      = "cognito:groups"
      match_type = "Contains"
      value      = "reviewers-group"
      role_arn   = aws_iam_role.cognito_reviewer_group_role.arn
    }
  }
}

# Outputs for JWT integration
output "user_pool_id" {
  description = "Cognito User Pool ID - used for JWT validation"
  value       = aws_cognito_user_pool.cognito_pool.id
}

output "user_pool_arn" {
  description = "Cognito User Pool ARN"
  value       = aws_cognito_user_pool.cognito_pool.arn
}

output "user_pool_endpoint" {
  description = "Cognito User Pool endpoint - used for JWKS URL construction"
  value       = aws_cognito_user_pool.cognito_pool.endpoint
}

output "app_client_id" {
  description = "App Client ID - used as audience (aud) claim in JWT validation"
  value       = aws_cognito_user_pool_client.client.id
}

output "jwks_uri" {
  description = "JWKS URI for JWT signature validation - contains public keys"
  value       = "https://${aws_cognito_user_pool.cognito_pool.endpoint}/.well-known/jwks.json"
}

output "identity_pool_id" {
  description = "Identity Pool ID - used for exchanging JWT tokens for AWS credentials"
  value       = aws_cognito_identity_pool.main.id
}

output "jwt_issuer" {
  description = "JWT issuer (iss claim) - used for JWT validation"
  value       = "https://${aws_cognito_user_pool.cognito_pool.endpoint}"
}