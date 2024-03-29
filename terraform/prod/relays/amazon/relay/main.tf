# --------------------------------------------------------------------------

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

# --------------------------------------------------------------------------

variable "name" { type = string }
variable "zone" { type = string }
variable "region" { type = string }
variable "type" { type = string }
variable "ami" { type = string }
variable "security_group_id" { type = string }
variable "vpn_address" { type = string }

# --------------------------------------------------------------------------

data "aws_ami" "ubuntu" {
  most_recent = true
  filter {
    name   = "name"
    values = [var.ami]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "relay" {
  availability_zone      = var.zone
  instance_type          = var.type
  ami                    = data.aws_ami.ubuntu.id
  key_name               = "prod-region-ssh-key"
  vpc_security_group_ids = [var.security_group_id]
  tags = {
    Name = "prod-${var.name}"
  }
  lifecycle {
    create_before_destroy = true
    ignore_changes = [ami]
  }
  user_data = replace(file("../../../scripts/init_relay.sh"), "$VPN_ADDRESS", var.vpn_address)
}

# --------------------------------------------------------------------------

output "public_address" {
  description = "Relay public address"
  value = aws_instance.relay.public_ip
}

output "internal_address" {
  description = "Relay internal address"
  value = aws_instance.relay.private_ip
}

# --------------------------------------------------------------------------
