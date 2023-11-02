
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

resource "aws_default_vpc" "default" {
  tags = {
    Name = "dev-default"
  }
}

resource "aws_key_pair" "ssh_key" {
  key_name   = "dev-region-ssh-key"
  public_key = file(var.ssh_public_key_file)
}

resource "aws_security_group" "allow_ssh_and_udp" {

  name = "dev-region-security-group"

  vpc_id = aws_default_vpc.default.id

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["${var.vpn_address}/32"]
  }

  ingress {
    protocol    = "udp"
    from_port   = 40000
    to_port     = 40000
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

output "security_group_id" {
  description = "The id of the security group for this region"
  value = aws_security_group.allow_ssh_and_udp.id
}
