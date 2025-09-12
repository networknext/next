# ----------------------------------------------------------------------------------------

// Latitude now supports a terraform provider
// https://registry.terraform.io/providers/latitudesh/latitudesh/latest/docs

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Latitude"

  seller_code = "latitude"

  ssh_user = "ubuntu"

  datacenter_map = {

    "latitude.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = "CHI"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.newyork" = {
      latitude    = 40.7128
      longitude   = -74.0060
      native_name = "NYC"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.ashburn" = {
      latitude    = 39.0438
      longitude   = -77.4874
      native_name = "ASH"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = "DAL2"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.miami" = {
      latitude    = 25.7617
      longitude   = -80.1918
      native_name = "MIA"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.losangeles" = {
      latitude    = 34.0549
      longitude   = -118.2426
      native_name = "LAX"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.mexico.1" = {
      latitude    = 23.6345
      longitude   = -102.5528
      native_name = "MEX"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.mexico.2" = {
      latitude    = 23.6345
      longitude   = -102.5528
      native_name = "MEX2"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.bogota" = {
      latitude    = 4.7110
      longitude   = -74.0721
      native_name = "BGT"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = "SAO"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.buenosaires" = {
      latitude    = -34.6037
      longitude   = -58.3816
      native_name = "BUE"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.santiago.1" = {
      latitude    = -33.4489
      longitude   = -70.6693
      native_name = "SAN"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.santiago.2" = {
      latitude    = -33.4489
      longitude   = -70.6693
      native_name = "SAN2"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
      native_name = "LON"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = "FRA"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.tokyo.1" = {
      latitude    = 35.6764
      longitude   = 139.6500
      native_name = "TYO"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.tokyo.2" = {
      latitude    = 35.6764
      longitude   = 139.6500
      native_name = "TYO2"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "latitude.sydney" = {
      latitude    = -33.8688
      longitude   = 151.2093
      native_name = "SYD"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for latitude"
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
  description = "All datacenters for latitude"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
