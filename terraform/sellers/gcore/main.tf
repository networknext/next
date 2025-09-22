# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "GCore"

  seller_code = "gcore"

  ssh_user = "root"

  datacenter_map = {

    "gcore.ashburn" = {
      latitude    = 39.0438
      longitude   = -77.4874
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.atlanta" = {
      latitude    = 33.7488
      longitude   = -84.3877
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.losangeles" = {
      latitude    = 34.0522
      longitude   = -118.2437
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.manassas" = {
      latitude    = 38.7509
      longitude   = -77.4753
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.miami" = {
      latitude    = 25.7617
      longitude   = -80.1918
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.newyork" = {
      latitude    = 40.7128
      longitude   = -74.0060
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.santaclara" = {
      latitude    = 37.3541
      longitude   = -121.9552
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.mcallen" = {
      latitude    = 26.2034
      longitude   = -98.2300
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.fortaleza" = {
      latitude    = -3.7327
      longitude   = -38.5270
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.lima" = {
      latitude    = -12.0464
      longitude   = -77.0428
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.queretaro" = {
      latitude    = 20.5888
      longitude   = -100.3899
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.buenosaires" = {
      latitude    = -34.6037
      longitude   = -58.3816
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.mexico" = {
      latitude    = 19.4326
      longitude   = -99.1332
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.santiago" = {
      latitude    = -33.4489
      longitude   = -70.6693
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.bucharest" = {
      latitude    = 44.4268
      longitude   = 26.1025
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.budapest" = {
      latitude    = 47.4979
      longitude   = 19.0402
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.darmstadt" = {
      latitude    = 49.8728
      longitude   = 8.6512
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.kishinev" = {
      latitude    = 47.0105
      longitude   = 28.8638
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.kyiv" = {
      latitude    = 50.4504
      longitude   = 30.5245
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.luxembourg" = {
      latitude    = 49.8153
      longitude   = 6.1296
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.madrid" = {
      latitude    = 40.4168
      longitude   = -3.7038
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.milan" = {
      latitude    = 45.4642
      longitude   = 9.1900
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.newport" = {
      latitude    = 51.5842
      longitude   = -2.9977
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.paris" = {
      latitude    = 48.8566
      longitude   = 2.3522
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.sofia" = {
      latitude    = 42.6977
      longitude   = 23.3219
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.warsaw" = {
      latitude    = 52.2297
      longitude   = 21.0122
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.prague" = {
      latitude    = 50.0755
      longitude   = 14.4378
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.vienna" = {
      latitude    = 48.2082
      longitude   = 16.3719
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.dubai" = {
      latitude    = 25.2048
      longitude   = 55.2708
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.istanbul" = {
      latitude    = 41.0082
      longitude   = 28.9784
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.telaviv" = {
      latitude    = 32.0853
      longitude   = 34.7818
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.yerevan" = {
      latitude    = 40.1872
      longitude   = 44.5152
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.johannesburg" = {
      latitude    = -26.2041
      longitude   = 28.0473
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.almaty" = {
      latitude    = 43.2380
      longitude   = 76.8829
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.hongkong" = {
      latitude    = 22.3193
      longitude   = 114.1694
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.mumbai" = {
      latitude    = 19.0760
      longitude   = 72.8777
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.seoul" = {
      latitude    = 37.5519
      longitude   = 126.9918
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.tokyo" = {
      latitude    = 35.6764
      longitude   = 139.6500
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.bangkok" = {
      latitude    = 13.7563
      longitude   = 100.5018
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.tashkent" = {
      latitude    = 41.2995
      longitude   = 69.2401
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.melbourne" = {
      latitude    = -37.8136
      longitude   = 144.9631
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "gcore.sydney" = {
      latitude    = -33.8688
      longitude   = 151.2093
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for gcore"
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
        v.ssh_user != "" ? v.ssh_user : local.ssh_user,
      ]
    )
  }
}

output "datacenters" {
  description = "All datacenters for gcore"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
