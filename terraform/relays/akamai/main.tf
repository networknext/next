# --------------------------------------------------------------------------

terraform {
  required_providers {
    linode = {
      source = "linode/linode"
      version = "1.27.1"
    }
  }
}

provider "linode" {
  token = file("~/Documents/terraform-akamai.txt")
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# --------------------------------------------------------------------------

resource "linode_instance" "relay" {
  count           = length(var.relays)
  image           = "linode/ubuntu22.04"
  label           = "relay"
  group           = "relays"
  region          = "us-east"
  type            = "g6-dedicated-2"
  private_ip      = true
  authorized_keys = [replace(file(var.ssh_public_key_file), "\n", "")]
}

output "relays" {
  description = "Data for each akamai relay setup by Terraform"
  value = [for i, v in var.relays : zipmap(["relay_name", "region", "public_address", "internal_address", "type", "image"], [var.relays[i].name, var.relays[i].region, linode_instance.relay[i].ip_address, linode_instance.relay[i].private_ip_address, var.relays[i].type, var.relays[i].image])]
}

# --------------------------------------------------------------------------
