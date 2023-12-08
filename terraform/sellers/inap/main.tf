# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "INAP"

  seller_code = "inap"

  ssh_user = "root"

  datacenter_map = {

    "inap.secaucus" = {
      latitude    = 40.7895
      longitude   = -74.0565
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "inap.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "inap.santaclara" = {
      latitude    = 37.3541
      longitude   = -121.9552
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "inap.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "inap.phoenix" = {
      latitude    = 33.4484
      longitude   = -112.0740
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "inap.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "inap.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "inap.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for inap"
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
  description = "All datacenters for inap"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
