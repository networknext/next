# ========================================================================================
#                                     GOOGLE CLOUD
# ========================================================================================

/*
    Here is the full set of amazon cloud datacenters, as of March 13, 2023:

      amazon.capetown.1
      amazon.capetown.2
      amazon.capetown.3
      amazon.hongkong.1
      amazon.hongkong.2
      amazon.hongkong.3
      amazon.tokyo.1
      amazon.tokyo.2
      amazon.tokyo.3
      amazon.seoul.1
      amazon.seoul.2
      amazon.seoul.3
      amazon.seoul.4
      amazon.osaka.1
      amazon.osaka.2
      amazon.osaka.3
      amazon.mumbai.1
      amazon.mumbai.2
      amazon.mumbai.3
      amazon.hyderabad.1
      amazon.hyderabad.2
      amazon.hyderabad.3
      amazon.singapore.1
      amazon.singapore.2
      amazon.singapore.3
      amazon.sydney.1
      amazon.sydney.2
      amazon.sydney.3
      amazon.jakarta.1
      amazon.jakarta.2
      amazon.jakarta.3
      amazon.melbourne.1
      amazon.melbourne.2
      amazon.melbourne.3
      amazon.montreal.1
      amazon.montreal.2
      amazon.montreal.3
      amazon.frankfurt.1
      amazon.frankfurt.2
      amazon.frankfurt.3
      amazon.zurich.1
      amazon.zurich.2
      amazon.zurich.3
      amazon.stockholm.1
      amazon.stockholm.2
      amazon.stockholm.3
      amazon.milan.1
      amazon.milan.2
      amazon.milan.3
      amazon.spain.1
      amazon.spain.2
      amazon.spain.3
      amazon.ireland.1
      amazon.ireland.2
      amazon.ireland.3
      amazon.london.1
      amazon.london.2
      amazon.london.3
      amazon.paris.1
      amazon.paris.2
      amazon.paris.3
      amazon.uae.1
      amazon.uae.2
      amazon.uae.3
      amazon.bahrain.1
      amazon.bahrain.2
      amazon.bahrain.3
      amazon.saopaulo.1
      amazon.saopaulo.2
      amazon.saopaulo.3
      amazon.virginia.1
      amazon.virginia.2
      amazon.virginia.3
      amazon.virginia.4
      amazon.virginia.5
      amazon.virginia.6
      amazon.ohio.1
      amazon.ohio.2
      amazon.ohio.3
      amazon.sanjose.1
      amazon.sanjose.2
      amazon.oregon.1
      amazon.oregon.2
      amazon.oregon.3
      amazon.oregon.4
      amazon.atlanta.1
      amazon.boston.1
      amazon.buenosaires.1
      amazon.chicago.1
      amazon.dallas.1
      amazon.houston.1
      amazon.lima.1
      amazon.kansas.1
      amazon.miami.1
      amazon.minneapolis.1
      amazon.newyork.1
      amazon.philadelphia.1
      amazon.queretaro.1
      amazon.santiago.1

    Some of these regions need to be manually added to your account before you can use them.

    For more details on enabling regions in your account, see: 

      https://docs.aws.amazon.com/general/latest/gr/rande-manage.html

    One all zones and regions are enabled, a manual configuration step is required 
    so terraform can map network next amazon datacenter names to the correct physical 
    amazon datacenters.

    Before deploying relays in AWS, run the network next AWS configuration tool:

      run amazon-config

    This is necessary because zone ids in AWS are different per-account.

    See this document for more details:

      https://docs.aws.amazon.com/ram/latest/userguide/working-with-az-ids.html
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
