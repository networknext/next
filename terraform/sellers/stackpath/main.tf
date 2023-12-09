# ----------------------------------------------------------------------------------------

// Stackpath now supports a terraform provider
// https://registry.terraform.io/providers/stackpath/stackpath/latest/docs

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Stackpath"

  seller_code = "stackpath"

  ssh_user = "root"

  datacenter_map = {

    "stackpath.ashburn" = {
      latitude    = 39.0438
      longitude   = -77.4874
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.newyork" = {
      latitude    = 40.7128
      longitude   = -74.0060
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.atlanta" = {
      latitude    = 33.7488
      longitude   = -84.3877
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.miami" = {
      latitude    = 25.7617
      longitude   = -80.1918
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.denver" = {
      latitude    = 39.7392
      longitude   = -104.9903
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.detroit" = {
      latitude    = 42.3314
      longitude   = -83.0458
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.boston" = {
      latitude    = 42.3601
      longitude   = -71.0589
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.stlouis" = {
      latitude    = 38.6270
      longitude   = -90.1994
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.minneapolis" = {
      latitude    = 44.9778
      longitude   = -93.2650
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.losangeles" = {
      latitude    = 34.0549
      longitude   = -118.2426
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.seattle" = {
      latitude    = 47.6061
      longitude   = -122.3328
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.sanjose" = {
      latitude    = 37.3387
      longitude   = -121.8853
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.dublin" = {
      latitude    = 53.3498
      longitude   = -6.2603
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.manchester" = {
      latitude    = 53.4808
      longitude   = -2.2426
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.madrid" = {
      latitude    = 40.4168
      longitude   = -3.7038
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.warsaw" = {
      latitude    = 52.2297
      longitude   = 21.0122
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.stockholm" = {
      latitude    = 59.3293
      longitude   = 18.0686
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.hongkong" = {
      latitude    = 22.3193
      longitude   = 114.1694
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.tokyo" = {
      latitude    = 35.6764
      longitude   = 139.6500
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.dubai" = {
      latitude    = 25.2048
      longitude   = 55.2708
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.mumbai" = {
      latitude    = 19.0760
      longitude   = 72.8777
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "stackpath.melbourne" = {
      latitude    = -37.8136
      longitude   = 144.9631
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for stackpath"
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
  description = "All datacenters for stackpath"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
