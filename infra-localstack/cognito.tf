# User user pool -> Autenticação
resource "aws_cognito_user_pool" "example_cog_pool" {
  name = "MyExamplePool"

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  mfa_configuration = "ON"

  software_token_mfa_configuration {
    enabled = true
  }

  #User attributes, definidos aqui
  schema {
    name                     = "terraform"
    attribute_data_type      = "Boolean"
    mutable                  = false
    required                 = false
    developer_only_attribute = false
  }

  schema {
    name                     = "foo"
    attribute_data_type      = "String"
    mutable                  = false
    required                 = false
    developer_only_attribute = false
    string_attribute_constraints {}
  }
}

resource "aws_cognito_user" "example_cog_user" {
  # The user pool ID for the user pool where the user will be created.
  user_pool_id = aws_cognito_user_pool.example_cog_pool.id
  # The username for the user. Must be unique within the user pool. Must be a UTF-8 string between 1 and 128 characters. After the user is created, the username cannot be changed.
  username     = "example"

  attributes = {
    terraform      = true
    foo            = "bar"
    email          = "no-reply@hashicorp.com"
    email_verified = true
  }
}

resource "aws_cognito_user" "grupo_l_membros" {
  user_pool_id = aws_cognito_user_pool.example_cog_pool.id
  username     = "example"

  attributes = {
    terraform      = true
    foo            = "bar"
    email          = "no-reply@hashicorp.com"
    email_verified = true
  }
}


# Usuário Admin
resource "aws_cognito_user" "admin" {
  user_pool_id = aws_cognito_user_pool.example_cog_pool.id
  username     = var.admin_cognito["email"]
  
  attributes = var.admin_cognito  # Chama variável diretamente!

  temporary_password = "AdminTemp123!"
  message_action     = "SUPPRESS"  # Não envia email
  enabled            = true
}

resource "aws_cognito_user" "user" {
  user_pool_id = aws_cognito_user_pool.example_cog_pool.id
  username     = var.user_cognito["email"]
  
  attributes = var.user_cognito

  temporary_password = "PassTemp123!"
  message_action     = "SUPPRESS"  # Não envia email
  enabled            = true
}