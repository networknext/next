# --------------------------------------------------------------------------

terraform {
  required_providers {
    equinix = {
      source = "equinix/equinix"
    }
  }
}

provider "equinix" {
  auth_token = file("~/secrets/terraform-equinix.txt")
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }
variable "project" { type = string }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# ----------------------------------------------------------------------------------------

resource "equinix_metal_project_ssh_key" "relay" {
  name       = "relay"
  project_id = var.project
  public_key = file(var.ssh_public_key_file)
}

resource "equinix_metal_device" "relay" {
  count               = length(var.relays)
  hostname            = var.relays[count.index].name
  plan                = var.relays[count.index].plan
  facilities          = [var.relays[count.index].zone]
  operating_system    = var.relays[count.index].os
  billing_cycle       = "hourly"
  project_id          = var.project
  project_ssh_key_ids = [equinix_metal_project_ssh_key.relay.id]
  user_data           = replace(file("./setup_relay.sh"), "$VPN_ADDRESS", var.vpn_address)
}

output "relays" {
  description = "Data for each equinix relay setup by Terraform"
  value = [for i, v in var.relays : zipmap(["relay_name", "zone", "public_address", "internal_address", "plan", "os"], [var.relays[i].name, var.relays[i].zone, equinix_metal_device.relay[i].network.0.address, equinix_metal_device.relay[i].network.2.address, var.relays[i].plan, var.relays[i].os])]
}

# ----------------------------------------------------------------------------------------
