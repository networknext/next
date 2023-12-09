# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "servers.com"

  seller_code = "serversdotcom"

  ssh_user = "root"

  datacenter_map = {

    "serversdotcom.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "serversdotcom.sanfran" = {
      latitude    = 37.7749
      longitude   = -122.4194
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "serversdotcom.washingtondc" = {
      latitude    = 38.9072
      longitude   = -77.0369
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "serversdotcom.hongkong" = {
      latitude    = 22.3193
      longitude   = 114.1694
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "serversdotcom.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "serversdotcom.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "serversdotcom.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "serversdotcom.luxembourg" = {
      latitude    = 49.8153
      longitude   = 6.1296
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "serversdotcom.lagos" = {
      latitude    = 6.5244
      longitude   = 3.3792
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "serversdotcom.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for serversdotcom"
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
  description = "All datacenters for serversdotcom"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
