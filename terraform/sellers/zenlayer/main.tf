# ----------------------------------------------------------------------------------------

// Zenlayer now supports a terraform provider for their cloud
// https://registry.terraform.io/providers/zenlayer/zenlayercloud/latest/docs

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Zenlayer"

  seller_code = "zenlayer"

  ssh_user = "root"

  datacenter_map = {

    "zenlayer.tokyo" = {
      latitude    = 35.6764
      longitude   = 139.6500
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.taipei" = {
      latitude    = 25.0330
      longitude   = 121.5654
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.hongkong" = {
      latitude    = 22.3193
      longitude   = 114.1694
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.hanoi" = {
      latitude    = 21.0278
      longitude   = 105.8342
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.manilla" = {
      latitude    = 14.5995
      longitude   = 120.9842
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.jakarta" = {
      latitude    = -6.1944
      longitude   = 106.8229
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.moscow" = {
      latitude    = 55.7558
      longitude   = 37.6173
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.paris" = {
      latitude    = 48.8566
      longitude   = 2.3522
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.zurich" = {
      latitude    = 47.3769
      longitude   = 8.5417
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.madrid" = {
      latitude    = 40.4168
      longitude   = -3.7038
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.sofia" = {
      latitude    = 42.6977
      longitude   = 23.3219
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.istanbul" = {
      latitude    = 41.0082
      longitude   = 28.9784
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.dubai" = {
      latitude    = 25.2048
      longitude   = 55.2708
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.telaviv" = {
      latitude    = 32.0853
      longitude   = 34.7818
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.riyadh" = {
      latitude    = 24.7136
      longitude   = 46.6753
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.jeddah" = {
      latitude    = 21.5292
      longitude   = 39.1611
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.nairobi" = {
      latitude    = -1.2921
      longitude   = 36.8219
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.lima" = {
      latitude    = -12.0464
      longitude   = -77.0428
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.bogota" = {
      latitude    = 4.7110
      longitude   = -74.0721
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.mexico" = {
      latitude    = 23.6345
      longitude   = -102.5528
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.losangeles" = {
      latitude    = 34.0549
      longitude   = -118.2426
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.ashburn" = {
      latitude    = 39.0438
      longitude   = -77.4874
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.miami" = {
      latitude    = 25.7617
      longitude   = -80.1918
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "zenlayer.buenosaires" = {
      latitude    = -34.6037
      longitude   = -58.3816
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for zenlayer"
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
        1,
      ]
    )
  }
}

output "datacenters" {
  description = "All datacenters for zenlayer"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
