# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Unity"

  seller_code = "unity"

  ssh_user = "root"

  datacenter_map = {

    "unity.saopaulo.1" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = "i3d.saopaulo"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "unity.saopaulo.2" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = "latitude.saopaulo"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "unity.saopaulo.3" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = "gcore.saopaulo"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for unity"
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
      ], 
      [
        k,
        v.datacenter_name,
        local.seller_name,
        local.seller_code, 
        v.public_address,
        40000,
        "0.0.0.0",
        0,
        "", 
        v.public_address,
        22,
        local.ssh_user,
      ]
    )
  }
}

output "datacenters" {
  description = "All datacenters for unity"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
