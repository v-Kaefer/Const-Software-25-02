resource "aws_security_group" "grupo_l_terraform" {
  name   = "allow-http"
  vpc_id = data.aws_vpc.default.id

  tags = {
    Name = "grupo-l-sprint1"
  }
}

#Inbound SSH
resource "aws_vpc_security_group_ingress_rule" "ssh_connection" {
  security_group_id = aws_security_group.grupo_l_terraform.id

  cidr_ipv4   = "0.0.0.0/0"
  from_port   = 22
  ip_protocol = "tcp"
  to_port     = 22
}

#Inbound Conexão ao servidor
resource "aws_vpc_security_group_ingress_rule" "www_connection" {
  security_group_id = aws_security_group.grupo_l_terraform.id

  cidr_ipv4   = "0.0.0.0/0"
  from_port   = 8080
  ip_protocol = "tcp"
  to_port     = 8080
}

#Inbound Geral
resource "aws_vpc_security_group_ingress_rule" "inbound" {
  security_group_id = aws_security_group.grupo_l_terraform.id

  cidr_ipv4   = "0.0.0.0/0" #Opcional
  ip_protocol = "icmp"      #Opcional
  from_port   = "-1"        #Obrigatória com ICMP
  to_port     = "-1"        #Obrigatória com ICMP
}

#Outbound Geral
resource "aws_vpc_security_group_egress_rule" "outbound" {
  security_group_id = aws_security_group.grupo_l_terraform.id

  cidr_ipv4   = "0.0.0.0/0" #Opcional
  ip_protocol = "icmp"      #Opcional
  from_port   = "-1"        #Obrigatória com ICMP
  to_port     = "-1"        #Obrigatória com ICMP
}