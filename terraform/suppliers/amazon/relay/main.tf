# --------------------------------------------------------------------------

variable "name" { type = string }
variable "zone" { type = string }
variable "region" { type = string }
variable "type" { type = string }
variable "ami" { type = string }
variable "security_group_id" { type = string }

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
  ami                    = var.ami
  key_name               = "ssh-key"
  vpc_security_group_ids = [var.security_group_id]
  tags = {
    Name = var.name
  }
  lifecycle {
    create_before_destroy = true
  }
  user_data = file("./setup_relay.sh")
}

resource "aws_eip" "relay" {
  instance = aws_instance.relay.id
  vpc      = true
}

# --------------------------------------------------------------------------

output "relay" {
  description = "Relay data"
  public_address = aws_eip.relay.public_ip
  public_address = aws_eip.relay.private_ip
}

# --------------------------------------------------------------------------
