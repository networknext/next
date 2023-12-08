# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "phoenixNAP"

  seller_code = "phoenixnap"

  ssh_user = "root"

  datacenter_map = {

    "phoenixnap.losangeles" = {
      latitude    = 34.0522
      longitude   = -118.2437
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.seattle" = {
      latitude    = 47.6061
      longitude   = -122.3328
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.phoenix" = {
      latitude    = 33.4484
      longitude   = -112.0740
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.austin" = {
      latitude    = 30.2672
      longitude   = -97.7431
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.ashburn" = {
      latitude    = 39.0438
      longitude   = -77.4874
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.atlanta" = {
      latitude    = 33.7488
      longitude   = -84.3877
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.madrid" = {
      latitude    = 40.4168
      longitude   = -3.7038
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.milan" = {
      latitude    = 45.4642
      longitude   = 9.1900
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.sofia" = {
      latitude    = 42.6977
      longitude   = 23.3219
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.belgrade" = {
      latitude    = 44.8125
      longitude   = 20.4612
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.helsinki" = {
      latitude    = 60.1699
      longitude   = 24.9384
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.warsaw" = {
      latitude    = 52.2297
      longitude   = 21.0122
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "phoenixnap.sydney" = {
      latitude    = -33.8688
      longitude   = 151.2093
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },
  }
}

output "relays" {
  description = "All relays for phoenixnap"
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
  description = "All datacenters for phoenixnap"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
