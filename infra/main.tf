provider "aws" {
  region                   = "us-east-1"
  shared_credentials_files = ["./.aws/credentials"]
}

resource "aws_key_pair" "this" {
  key_name   = "grupo-l-key"
  public_key = file(var.public_key_path)
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "grupo_l_terraform" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t2.micro"

  security_groups = ["allow-http"]

  tags = {
    Name = "grupo-l-sprint1"
  }
}
