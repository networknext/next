# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Azure"

  seller_code = "azure"

  ssh_user = "ubuntu"

  datacenter_map = {

    "azure.saopaulo.1" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "azure.saopaulo.2" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "azure.saopaulo.3" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for azure"
  value = {
    for k, v in var.relays : k => zipmap( 
      [
        "relay_name", 
        "datacenter_name",
        "seller_name",
        "seller_code",
        "public_ip",
        "public_port",
        "internal_ip",
        "internal_port",
        "internal_group",
        "ssh_ip",
        "ssh_port",
        "ssh_user",
        "bandwidth_price",
      ], 
      [
        k,
        v.datacenter_name,
        local.seller_name,
        local.seller_code, 
        v.public_address,
        40000,
        v.internal_address,
        40000,
        "", 
        v.public_address,
        22,
        local.ssh_user,
        2,
      ]
    )
  }
}

output "datacenters" {
  description = "All datacenters for azure"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
