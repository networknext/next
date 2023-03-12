# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    hivelocity = {
      version  = "0.5.0"
      source   = "hivelocity/hivelocity"
    }
  }
}

provider hivelocity {
  api_key  = file("~/Documents/terraform-hivelocity.txt")
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# ----------------------------------------------------------------------------------------

resource "hivelocity_ssh_key" "relay" {
  name       = "ssh key for relays"
  public_key = file(var.ssh_public_key_file)
}

data "hivelocity_product" "relay" {
  count = length(var.relays)
  first = true
  filter {
    name   = "product_memory"
    values = ["16GB"]
  }
  filter {
    name   = "data_center"
    values = [var.relays[count.index].zone]
  }
  filter {
    name   = "stock"
    values = ["limited", "available"]
  }
}

resource "hivelocity_bare_metal_device" "relay" {
  count             = length(var.relays)
  hostname          = "test.relay.fuckyou"  # what is up with this field... =p
  tags              = [var.relays[count.index].name]
  os_name           = var.relays[count.index].os
  period            = "Hourly"
  product_id        = data.hivelocity_product.relay[count.index].product_id
  location_name     = data.hivelocity_product.relay[count.index].data_center
  public_ssh_key_id = hivelocity_ssh_key.relay.ssh_key_id
  script            = file("./setup_relay.sh")
}

/*
resource "hivelocity_vlan" "private_vlan" {
  count         = length(var.relays)
  device_ids    = [
    hivelocity_bare_metal_device.relay[*].device_id
  ]
}
*/

output "relays" {
  description = "Data for each hivelocity relay setup by Terraform"
  value = [for i, v in var.relays : zipmap(["relay_name", "zone", "public_address", "internal_address", "os"], [var.relays[i].name, var.relays[i].zone, hivelocity_bare_metal_device.relay[i].primary_ip, "0.0.0.0", var.relays[i].os])]
}

# ----------------------------------------------------------------------------------------
