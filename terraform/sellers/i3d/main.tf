# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "i3D"

  seller_code = "i3d"

  ssh_user = "root"

  datacenter_map = {

    "i3d.losangeles" = {
      latitude    = 34.0522
      longitude   = -118.2437
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.buenosaires" = {
      latitude    = -34.6037
      longitude   = -58.3816
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.santiago" = {
      latitude    = -33.4489
      longitude   = -70.6693
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.montreal" = {
      latitude    = 45.5019
      longitude   = -73.5674
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.stlaurent" = {
      latitude    = 45.5023
      longitude   = -73.7061
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.toronto" = {
      latitude    = 43.6532
      longitude   = -79.3832
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.ashburn" = {
      latitude    = 39.0438
      longitude   = -77.4874
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.atlanta" = {
      latitude    = 33.7488
      longitude   = -84.3877
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.newark" = {
      latitude    = 40.7357
      longitude   = -74.1724
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.santaclara" = {
      latitude    = 37.3541
      longitude   = -121.9552
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.seattle" = {
      latitude    = 47.6061
      longitude   = -122.3328
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.vienna" = {
      latitude    = 48.2082
      longitude   = 16.3719
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.oslo" = {
      latitude    = 59.9139
      longitude   = 10.7522
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.copenhagen" = {
      latitude    = 55.6761
      longitude   = 12.5683
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.warsaw" = {
      latitude    = 52.2297
      longitude   = 21.0122
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.paris" = {
      latitude    = 48.8566
      longitude   = 2.3522
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.stdenis" = {
      latitude    = 48.9362
      longitude   = 2.3574
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.bucharest" = {
      latitude    = 44.4268
      longitude   = 26.1025
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.budapest" = {
      latitude    = 47.4979
      longitude   = 19.0402
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.johannesburg" = {
      latitude    = -26.2041
      longitude   = 28.0473
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.madrid" = {
      latitude    = 40.4168
      longitude   = -3.7038
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.milan" = {
      latitude    = 45.4642
      longitude   = 9.1900
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.stockholm" = {
      latitude    = 59.3293
      longitude   = 18.0686
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.zurich" = {
      latitude    = 47.3769
      longitude   = 8.5417
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.eindhoven" = {
      latitude    = 51.4231
      longitude   = 5.4623
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.heerlen" = {
      latitude    = 50.8860
      longitude   = 5.9804
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.rotterdam" = {
      latitude    = 51.9244
      longitude   = 4.4777
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.kyiv" = {
      latitude    = 50.4504
      longitude   = 30.5245
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.dubai" = {
      latitude    = 25.2048
      longitude   = 55.2708
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.fujairah" = {
      latitude    = 25.1288
      longitude   = 56.3265
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.sydney" = {
      latitude    = -33.8688
      longitude   = 151.2093
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.mumbai" = {
      latitude    = 19.0760
      longitude   = 72.8777
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.hongkong" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.tokyo" = {
      latitude    = 35.6764
      longitude   = 139.6500
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "i3d.seoul" = {
      latitude    = 37.5519
      longitude   = 126.9918
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for i3d"
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
  description = "All datacenters for i3d"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
