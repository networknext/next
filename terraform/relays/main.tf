# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.51.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

# ----------------------------------------------------------------------------------------

variable "credentials" { type = string }
variable "project" { type = string }
variable "vpn_address" { type = string }
variable "ssh_public_key_file" { type = string }
variable "ssh_private_key_file" { type = string }
variable "env" { type = string }
variable "relay_version" { type = string }
variable "relay_artifacts_bucket" { type = string }
variable "relay_public_key" { type = string }
variable "relay_private_key" { type = string }
variable "relay_backend_hostname" { type = string }
variable "relay_backend_public_key" { type = string }

# ----------------------------------------------------------------------------------------

locals {
  context = {
    vpn_address              = var.vpn_address
    ssh_public_key_file      = var.ssh_public_key_file
    env                      = var.env
    relay_version            = var.relay_version
    relay_artifacts_bucket   = var.relay_artifacts_bucket
    relay_public_key         = var.relay_public_key
    relay_private_key        = var.relay_private_key
    relay_backend_hostname   = var.relay_backend_hostname
    relay_backend_public_key = var.relay_backend_public_key    
  }
}

# ----------------------------------------------------------------------------------------

provider "google" {
  credentials = file(var.credentials)
  project     = var.project
}

data "google_compute_network" "default" {
  name = "default"
}

resource "google_compute_firewall" "google_allow_ssh" {
  name          = "allow-ssh"
  project       = var.project
  direction     = "INGRESS"
  network       = "default"
  source_ranges = [var.vpn_address]
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}

resource "google_compute_firewall" "google_allow_udp" {
  name          = "allow-udp"
  project       = var.project
  direction     = "INGRESS"
  network       = "default"
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
    ports    = ["40000"]
  }
}

# us-central1: IOWA

module "google_iowa_1" {
  relay_name        = "google.iowa.1"
  zone              = "us-central1-a"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

module "google_iowa_2" {
  relay_name        = "google.iowa.2"
  zone              = "us-central1-b"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

module "google_iowa_3" {
  relay_name        = "google.iowa.3"
  zone              = "us-central1-c"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

module "google_iowa_4" {
  relay_name        = "google.iowa.4"
  zone              = "us-central1-f"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

# us-west2: LOS ANGELES

module "google_losangeles_1" {
  relay_name        = "google.losangeles.1"
  zone              = "us-west2-a"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

module "google_losangeles_2" {
  relay_name        = "google.losangeles.2"
  zone              = "us-west2-b"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

# ----------------------------------------------------------------------------------------

locals {
  relay_name = "relay"
}

provider "aws" {
  shared_config_files      = ["~/.aws/config"]
  shared_credentials_files = ["~/.aws/credentials"]
  profile                  = "default"
  region                   = "us-west-2"
}

resource "aws_default_vpc" "default" {
  tags = {
    Name = "default"
  }
}

resource "aws_security_group" "amazon_allow_ssh_and_udp" {

  name = "amazon-allow-ssh-and-udp"

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

data "aws_ami" "amazon_ubuntu" {
  most_recent = true
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
  owners = ["099720109477"] # Canonical
}

resource "aws_key_pair" "amazon_ssh_key" {
  key_name   = "amazon-ssh-key"
  public_key = file(var.ssh_public_key_file)
}

resource "aws_instance" "amazon_relay" {
  availability_zone      = "us-west-2a"
  instance_type          = "t2.micro"
  ami                    = data.aws_ami.amazon_ubuntu.id
  key_name               = "amazon-ssh-key"
  vpc_security_group_ids = [aws_security_group.amazon_allow_ssh_and_udp.id]
  tags = {
    Name = "relay"
  }
  lifecycle {
    create_before_destroy = true
  }
  connection {
    type        = "ssh"
    host        = self.public_ip
    user        = "ubuntu"
    private_key = file(var.ssh_private_key_file)
    timeout     = "10m"
  }
  provisioner "remote-exec" {
    inline ["echo ANUS"]
  }
}

resource "aws_eip" "amazon_address" {
  instance = aws_instance.amazon_relay.id
  vpc      = true
}

output "amazon_public_address" {
  description = "The public IP address of amazon relay"
  value = aws_eip.amazon_address.public_ip
}

output "amazon_internal_address" {
  description = "The internal IP address of amazon relay"
  value = aws_eip.amazon_address.private_ip
}

# ----------------------------------------------------------------------------------------
