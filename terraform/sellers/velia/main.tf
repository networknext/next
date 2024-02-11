# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Velia"

  seller_code = "velia"

  ssh_user = "root"

  datacenter_map = {

    "velia.stlouis" = {
      latitude    = 38.6270
      longitude   = -90.1994
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "velia.phoenix" = {
      latitude    = 33.4484
      longitude   = -112.0740
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "velia.miami" = {
      latitude    = 25.7617
      longitude   = -80.1918
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "velia.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "velia.strasbourg" = {
      latitude    = 48.5734
      longitude   = 7.7521
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "velia.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for velia"
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
  description = "All datacenters for velia"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
