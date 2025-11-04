variable "aws_region" {
  description = "AWS region para deploy dos recursos"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Ambiente de deployment (development, staging, production)"
  type        = string
  default     = "development"
}

variable "cognito_domain_prefix" {
  description = "Prefixo para o domínio do Cognito Hosted UI"
  type        = string
  default     = "user-service"
}

variable "project_name" {
  description = "Nome do projeto para tagging"
  type        = string
  default     = "user-service"
}
