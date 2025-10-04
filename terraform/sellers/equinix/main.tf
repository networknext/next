# ----------------------------------------------------------------------------------------

// Equinix metal has a terraform provider
// https://registry.terraform.io/providers/equinix/equinix/latest/docs

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Equinix"

  seller_code = "equinix"

  ssh_user = "root"

  datacenter_map = {

    "equinix.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = "AM"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.atlanta" = {
      latitude    = 33.7488
      longitude   = -84.3877
      native_name = "AT"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = "CH"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = "DA"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = "FR"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.helsinki" = {
      latitude    = 60.1699
      longitude   = 24.9384
      native_name = "HE"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.hongkong" = {
      latitude    = 22.3193
      longitude   = 114.1694
      native_name = "HK"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.losangeles" = {
      latitude    = 34.0549
      longitude   = -118.2426
      native_name = "LA"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
      native_name = "LD"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.madrid" = {
      latitude    = 40.4168
      longitude   = -3.7038
      native_name = "MD"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.manchester" = {
      latitude    = 53.4808
      longitude   = -2.2426
      native_name = "MA"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.melbourne" = {
      latitude    = -37.8136 
      longitude   = -144.9631
      native_name = "ME"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.miami" = {
      latitude    = 25.7617
      longitude   = -80.1918
      native_name = "MI"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.montreal" = {
      latitude    = 45.5019
      longitude   = -73.5674
      native_name = "MT"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.newyork" = {
      latitude    = 40.7128
      longitude   = -74.0060
      native_name = "NY"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.osaka" = {
      latitude    = 34.6937
      longitude   = 135.5023
      native_name = "OS"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.paris" = {
      latitude    = 48.8566
      longitude   = 2.3522
      native_name = "PA"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = "SP"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.seoul" = {
      latitude    = 37.5519
      longitude   = 126.9918
      native_name = "SL"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.seattle" = {
      latitude    = 47.6061
      longitude   = -122.3328
      native_name = "SE"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.siliconvalley" = {
      latitude    = 37.3387
      longitude   = -121.8853
      native_name = "SV"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = "SG"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.stockholm" = {
      latitude    = 59.3293
      longitude   = 18.0686
      native_name = "SK"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.sydney" = {
      latitude    = -33.8688
      longitude   = 151.2093
      native_name = "SY"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.toronto" = {
      latitude    = 43.6532
      longitude   = -79.3832
      native_name = "TR"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.tokyo" = {
      latitude    = 35.6764
      longitude   = 139.6500
      native_name = "TY"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "equinix.washingtondc" = {
      latitude    = 38.9072
      longitude   = -77.0369
      native_name = "DC"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for equinix"
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
  description = "All datacenters for equinix"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
