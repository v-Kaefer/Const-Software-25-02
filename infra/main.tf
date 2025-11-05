provider "aws" {
  region                   = "us-east-1"
  shared_credentials_files = ["./.aws/credentials"]
}

resource "aws_key_pair" "this" {
  key_name   = "grupo-l-key"
  public_key = file(var.public_key_path)
}

# Lookup real Ubuntu AMI for production deployments
# This data source is only used when not running with LocalStack
data "aws_ami" "ubuntu" {
  count       = var.use_localstack ? 0 : 1
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "grupo_l_terraform" {
  # Use mock AMI for LocalStack, real AMI lookup for production
  # LocalStack accepts any AMI ID in the format ami-xxxxxxxx
  ami           = var.use_localstack ? "ami-ff0fea8310f3" : data.aws_ami.ubuntu[0].id
  instance_type = "t2.micro"

  security_groups = ["allow-http"]

  tags = {
    Name = "grupo-l-sprint1"
  }
}
