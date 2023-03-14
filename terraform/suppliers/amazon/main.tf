# ==========================================================================
#                               AMAZON CLOUD
# ==========================================================================

/*
    Before deploying relays in AWS, run the amazon configuration tool:

      run amazon-config

    It generates config/amazon.txt, schemas/sql/sellers/amazon.sql and
    terraform/amazon.tfvars which is used by this module.

    Some AWS regions need to be manually enabled on your account before you can use them.

    For more details on enabling regions in your account, see: 

      https://docs.aws.amazon.com/general/latest/gr/rande-manage.html
*/

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
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  region                   = var.region
}

# --------------------------------------------------------------------------

variable "config" { type = list(string) }
variable "credentials" { type = list(string) }
variable "profile" { type = string }
variable "relays" { type = map(map(string)) }
variable "region" { type = string }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# ----------------------------------------------------------------------------------------

locals {

  /*
      IMPORTANT: This map translates from the network next amazon datacenter names to 
      AWS regions, zones and AZID.
      
      If want you add more datacenters from amazon, then you must:

          1. Enable the datacenters in your account

          2. Run the amazon tool: "run amazon-tool"

          3. Update schemas/sql/sellers/amazon.sql then update your postgres database

          4. Update config/amazon.txt then "Deploy Config" config to your environment via semaphore

      Please be extremely careful making these changes!
  */

  datacenter_map = {

    "amazon.virginia.1" = {
      azid   = "use1-az1"
      zone   = "us-east-1a"
      region = "us-east-1"
    },

  }
}

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
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
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
  for_each = var.relays
  most_recent = true
  filter {
    name   = "name"
    values = [each.value.ami]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "relay" {
  for_each               = var.relays
  availability_zone      = each.value.zone
  instance_type          = each.value.type
  ami                    = data.aws_ami.ubuntu[each.key].id
  key_name               = "ssh-key"
  vpc_security_group_ids = [aws_security_group.allow_ssh_and_udp.id]
  tags = {
    Name = each.key
  }
  lifecycle {
    create_before_destroy = true
  }
  user_data = file("./setup_relay.sh")
}

resource "aws_eip" "address" {
  for_each = var.relays
  instance = aws_instance.relay[each.key].id
  vpc      = true
}

# --------------------------------------------------------------------------

output "relays" {

  description = "Data for each bare metal relay setup by Terraform"

  value = {
    for k, v in var.relays : k => zipmap( 
      [
        "relay_name",
        "native_name",
        "datacenter_name",
        "supplier_name", 
        "public_address", 
        "internal_address", 
        "internal_group", 
        "ssh_address", 
        "ssh_user",
      ], 
      [
        k,
        local.datacenter_map[v.datacenter_name].azid,
        v.datacenter_name,
        "amazon", 
        "${aws_eip.address[k].public_ip}:40000",
        "${aws_eip.address[k].private_ip}:40000",
        local.datacenter_map[v.datacenter_name].region,
        "${aws_eip.address[k].public_ip}:22",
        "ubuntu",
      ]
    )
  }
}

# --------------------------------------------------------------------------
