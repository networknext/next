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
    },

    "datapacket.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
    },

    "datapacket.stockholm" = {
      latitude    = 59.3293
      longitude   = 18.0686
    },

    "datapacket.copenhagen" = {
      latitude    = 55.6761
      longitude   = 12.5683
    },

    "datapacket.dublin" = {
      latitude    = 53.3498
      longitude   = -6.2603
    },

    "datapacket.london" = {
      latitude    = 51.5072
      longitude   = -0.1276
    },

    "datapacket.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
    },

    "datapacket.warsaw" = {
      latitude    = 52.2297
      longitude   = 21.0122
    },

    "datapacket.kyiv" = {
      latitude    = 50.4504
      longitude   = 30.5245
    },

    "datapacket.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
    },

    "datapacket.brussels" = {
      latitude    = 50.8476
      longitude   = 4.3572
    },

    "datapacket.zurich" = {
      latitude    = 47.3769
      longitude   = 8.5417
    },

    "datapacket.prague" = {
      latitude    = 50.0755
      longitude   = 14.4378
    },

    "datapacket.bratislava" = {
      latitude    = 48.1486
      longitude   = 17.1077
    },

    "datapacket.paris" = {
      latitude    = 48.8566
      longitude   = 2.3522
    },

    "datapacket.vienna" = {
      latitude    = 48.2082
      longitude   = 16.3719
    },

    "datapacket.bucharest" = {
      latitude    = 48.2082
      longitude   = 16.3719
    },

    "datapacket.milan" = {
      latitude    = 45.4642
      longitude   = 9.1900
    },

    "datapacket.zagreb" = {
      latitude    = 45.8150
      longitude   = 15.9819
    },

    "datapacket.sofia" = {
      latitude    = 42.6977
      longitude   = 23.3219
    },

    "datapacket.lisbon" = {
      latitude    = 38.7223
      longitude   = 9.1393
    },

    "datapacket.madrid" = {
      latitude    = 40.4168
      longitude   = 3.7038
    },

    "datapacket.madrid" = {
      latitude    = 43.2965
      longitude   = 5.3698
    },

    "datapacket.palermo" = {
      latitude    = 38.1157
      longitude   = 13.3615
    },

    "datapacket.athens" = {
      latitude    = 38.1157
      longitude   = 13.3615
    },

    "datapacket.istanbul" = {
      latitude    = 41.0082
      longitude   = 28.9784
    },

    "datapacket.telaviv" = {
      latitude    = 32.0853
      longitude   = 34.7818
    },

    "datapacket.johannesburg" = {
      latitude    = 32.0853
      longitude   = 34.7818
    },

    "datapacket.sydney" = {
      latitude    = -33.8688
      longitude   = 151.2093
    },

    "datapacket.singapore" = {
      latitude    = 1.3521
      longitude   = 103.8198
    },

    "datapacket.hongkong" = {
      latitude    = 22.3193
      longitude   = 114.1694
    },

    "datapacket.tokyo" = {
      latitude    = 35.6764
      longitude   = 139.6500
    },

    "datapacket.queretaro" = {
      latitude    = 20.5888
      longitude   = -100.3899
    },

    "datapacket.bogota" = {
      latitude    = 4.7110
      longitude   = -74.0721
    },

    "datapacket.saopaulo" = {
      latitude    = -23.5558
      longitude   = -46.6396
    },

    "datapacket.santiago" = {
      latitude    = -33.4489,
      longitude   = -70.6693,
    },

    "datapacket.vancouver" = {
      latitude    = 49.2827
      longitude   = -123.1207 
    },

    "datapacket.seattle" = {
      latitude    = 47.6061
      longitude   = -122.3328
    },

    "datapacket.denver" = {
      latitude    = 39.7392
      longitude   = -104.9903
    },

    "datapacket.sanjose" = {
      latitude    = 37.3387
      longitude   = -121.8853
    },

    "datapacket.toronto" = {
      latitude    = 43.6532
      longitude   = -79.3832
    },

    "datapacket.boston" = {
      latitude    = 42.3601
      longitude   = -71.0589
    },

    "datapacket.newyork" = {
      latitude    = 40.7128
      longitude   = -74.0060
    },

    "datapacket.losangeles" = {
      latitude    = 34.0549
      longitude   = -118.2426
    },

    "datapacket.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
    },

    "datapacket.houston" = {
      latitude    = 29.7604
      longitude   = -95.3698
    },

    "datapacket.mcallen" = {
      latitude    = 26.2034
      longitude   = -98.2300
    },

    "datapacket.miami" = {
      latitude    = 25.7617
      longitude   = -80.1918
    },

    "datapacket.ashburn" = {
      latitude    = 39.0438
      longitude   = -77.4874
    },

    "datapacket.atlanta" = {
      latitude    = 33.7488
      longitude   = -84.3877
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
        "public_address", 
        "internal_address", 
        "internal_group", 
        "ssh_address", 
        "ssh_user",
      ], 
      [
        k,
        v.datacenter_name,
        local.seller_name,
        local.seller_code,
        v.public_address, 
        "", 
        0, 
        v.public_address, 
        local.ssh_user,
      ]
    )
  }
}

output "datacenters" {
  description = "All datacenters for datapacket"
  value = locals.datacenter_map
}

# --------------------------------------------------------------------------
