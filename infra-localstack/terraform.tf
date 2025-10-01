terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.92"
    }
  }
  required_version = ">= 1.2"
}

data "aws_vpc" "default" { default = true }

# Importa sua chave pública local (~/.ssh/id_rsa.pub, por ex.)
variable "public_key_path" { default = "~/.ssh/id_rsa.pub" }
