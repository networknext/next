# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Hivelocity"

  seller_code = "hivelocity"

  ssh_user = "root"

  datacenter_map = {

    "hivelocity.amsterdam" = {
      latitude    = 52.3676
      longitude   = 4.9041
      native_name = "AMS1"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.atlanta" = {
      latitude    = 33.7488
      longitude   = -84.3877
      native_name = "ATL2"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = "ORD1"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = "DAL1"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = "FRA1"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.losangeles" = {
      latitude    = 34.0549
      longitude   = -118.2426
      native_name = "LA1 and LA2"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.miami" = {
      latitude    = 25.7617
      longitude   = -80.1918
      native_name = "MIA1"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.newyork" = {
      latitude    = 40.7128
      longitude   = -74.0060
      native_name = "NYC1"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.seattle" = {
      latitude    = 47.6061
      longitude   = -122.3328
      native_name = "SEA1"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.tampa.1" = {
      latitude    = 27.9517
      longitude   = -82.4588
      native_name = "TPA1"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "hivelocity.tampa.2" = {
      latitude    = 27.9517
      longitude   = -82.4588
      native_name = "TPA2"
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for hivelocity"
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
  description = "All datacenters for hivelocity"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
