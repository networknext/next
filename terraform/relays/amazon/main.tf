# --------------------------------------------------------------------------

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  shared_config_files      = ["~/.aws/config"]
  shared_credentials_files = ["~/.aws/credentials"]
  profile                  = "default"
  region                   = local.regions[count.index]
}

# --------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }
variable "region" { type = string }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# --------------------------------------------------------------------------

resource "aws_default_vpc" "default" {
  tags = {
    Name = "default"
  }
}

resource "aws_security_group" "allow_ssh_and_udp" {

  name = "allow-ssh-and-udp"

  vpc_id = aws_default_vpc.default.id

  ingress {
    protocol    = "tcp"
    from_port   = 22
      to_port     = 22
    cidr_blocks = ["${var.vpn_address}/32"]
  }

  ingress {
    protocol    = "udp"
    from_port   = 45000
    to_port     = 45000
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    cidr_blocks     = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_key_pair" "ssh_key" {
  key_name   = "ssh-key"
  public_key = file(var.ssh_public_key_file)
}

data "aws_ami" "ubuntu" {
  count = length(var.relays)
  most_recent = true
  filter {
    name   = "name"
    values = [var.relays[count.index].ami]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "relay" {
  count                  = length(var.relays)
  availability_zone      = var.relays[count.index].zone
  instance_type          = var.relays[count.index].type
  ami                    = data.aws_ami.ubuntu[count.index].id
  key_name               = "ssh-key"
  vpc_security_group_ids = [aws_security_group.allow_ssh_and_udp.id]
  tags = {
    Name = "var.relays[count.index].name}"
  }
  lifecycle {
    create_before_destroy = true
  }
  user_data = file("./setup_relay.sh")
}

resource "aws_eip" "address" {
  count    = length(var.relays)
  instance = aws_instance.relay[count.index].id
  vpc      = true
}

output "relays" {
  description = "Data for each amazon relay setup by Terraform"
  value = [for i, v in var.relays : zipmap(["relay_name", "zone", "region", "public_address", "internal_address", "instance_type", "ami"], [var.relays[i].name, var.relays[i].zone, var.region, aws_eip.address[i].public_ip, aws_eip.address[i].private_ip, var.relays[i].type, data.aws_ami.ubuntu[i].id])]
}

# --------------------------------------------------------------------------
