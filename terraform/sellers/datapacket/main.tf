# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Datapacket"

  seller_code = "datapacket"

  ssh_user = "root"

  datacenter_map = {

    "datapacket.losangeles" = {
      latitude    = 34.0522
      longitude   = -118.2437
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
   },

    "datapacket.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.stockholm" = {
      latitude    = 59.3293
      longitude   = 18.0686
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.copenhagen" = {
      latitude    = 55.6761
      longitude   = 12.5683
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.dublin" = {
      latitude    = 53.3498
      longitude   = -6.2603
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.warsaw" = {
      latitude    = 52.2297
      longitude   = 21.0122
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.kyiv" = {
      latitude    = 50.4504
      longitude   = 30.5245
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.brussels" = {
      latitude    = 50.8476
      longitude   = 4.3572
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.zurich" = {
      latitude    = 47.3769
      longitude   = 8.5417
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.prague" = {
      latitude    = 50.0755
      longitude   = 14.4378
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.bratislava" = {
      latitude    = 48.1486
      longitude   = 17.1077
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.paris" = {
      latitude    = 48.8566
      longitude   = 2.3522
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.vienna" = {
      latitude    = 48.2082
      longitude   = 16.3719
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.bucharest" = {
      latitude    = 48.2082
      longitude   = 16.3719
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.milan" = {
      latitude    = 45.4642
      longitude   = 9.1900
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.zagreb" = {
      latitude    = 45.8150
      longitude   = 15.9819
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.sofia" = {
      latitude    = 42.6977
      longitude   = 23.3219
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.lisbon" = {
      latitude    = 38.7223
      longitude   = 9.1393
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.madrid" = {
      latitude    = 40.4168
      longitude   = 3.7038
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.madrid" = {
      latitude    = 43.2965
      longitude   = 5.3698
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.palermo" = {
      latitude    = 38.1157
      longitude   = 13.3615
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.athens" = {
      latitude    = 38.1157
      longitude   = 13.3615
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.istanbul" = {
      latitude    = 41.0082
      longitude   = 28.9784
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.telaviv" = {
      latitude    = 32.0853
      longitude   = 34.7818
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.johannesburg" = {
      latitude    = 32.0853
      longitude   = 34.7818
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.sydney" = {
      latitude    = -33.8688
      longitude   = 151.2093
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.hongkong" = {
      latitude    = 22.3193
      longitude   = 114.1694
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.tokyo" = {
      latitude    = 35.6764
      longitude   = 139.6500
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.queretaro" = {
      latitude    = 20.5888
      longitude   = -100.3899
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.bogota" = {
      latitude    = 4.7110
      longitude   = -74.0721
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.santiago" = {
      latitude    = -33.4489,
      longitude   = -70.6693,
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.vancouver" = {
      latitude    = 49.2827
      longitude   = -123.1207 
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.seattle" = {
      latitude    = 47.6061
      longitude   = -122.3328
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.denver" = {
      latitude    = 39.7392
      longitude   = -104.9903
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.sanjose" = {
      latitude    = 37.3387
      longitude   = -121.8853
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.toronto" = {
      latitude    = 43.6532
      longitude   = -79.3832
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.boston" = {
      latitude    = 42.3601
      longitude   = -71.0589
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.newyork" = {
      latitude    = 40.7128
      longitude   = -74.0060
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.losangeles" = {
      latitude    = 34.0549
      longitude   = -118.2426
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.houston" = {
      latitude    = 29.7604
      longitude   = -95.3698
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.mcallen" = {
      latitude    = 26.2034
      longitude   = -98.2300
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.miami" = {
      latitude    = 25.7617
      longitude   = -80.1918
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.ashburn" = {
      latitude    = 39.0438
      longitude   = -77.4874
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.atlanta" = {
      latitude    = 33.7488
      longitude   = -84.3877
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "datapacket.lima" = {
      latitude    = -12.0467,
      longitude   = -77.0431,
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }

}

output "relays" {
  description = "All relays for datapacket"
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
  description = "All datacenters for datapacket"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
