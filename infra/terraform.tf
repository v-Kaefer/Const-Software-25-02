terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.92"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
  required_version = ">= 1.2"
}

data "aws_vpc" "default" { default = true }

# Importa sua chave pública local (~/.ssh/id_rsa.pub, por ex.)
variable "public_key_path" { default = "~/.ssh/id_rsa.pub" }

# Variable to indicate if running with LocalStack
# Set to true when using tflocal for local testing
variable "use_localstack" {
  description = "Whether to use LocalStack mock AMI instead of real AMI lookup"
  type        = bool
  default     = false
}
