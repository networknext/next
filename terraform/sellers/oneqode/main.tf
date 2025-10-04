# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Oneqode"

  seller_code = "oneqode"

  ssh_user = "root"

  datacenter_map = {

    "oneqode.losangeles" = {
      latitude    = 34.0522
      longitude   = -118.2437
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "oneqode.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "oneqode.hongkong" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "oneqode.tokyo" = {
      latitude    = 35.6764
      longitude   = 139.6500
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "oneqode.sydney" = {
      latitude    = -33.8688
      longitude   = 151.2093
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "oneqode.melbourne" = {
      latitude    = -37.8136
      longitude   = 144.9631
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "oneqode.brisbane" = {
      latitude    = -27.4705
      longitude   = 153.0260
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "oneqode.sunshinecoast" = {
      latitude    = -26.6500
      longitude   = 153.0667
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "oneqode.manilla" = {
      latitude    = 14.5995
      longitude   = 120.9842
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "oneqode.guam" = {
      latitude    = 13.4443
      longitude   = 144.7937
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for oneqode"
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
        "0.0.0.0",
        0,
        "", 
        v.public_address,
        22,
        local.ssh_user,
        0,
      ]
    )
  }
}

output "datacenters" {
  description = "All datacenters for oneqode"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
