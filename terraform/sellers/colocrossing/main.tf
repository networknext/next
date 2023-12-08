# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "ColoCrossing"

  seller_code = "colocrossing"

  ssh_user = "root"

  datacenter_map = {

    "colocrossing.buffalo" = {
      latitude    = 42.8864
      longitude   = -78.8784
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "colocrossing.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "colocrossing.newyork" = {
      latitude    = 40.7128
      longitude   = -74.0060
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "colocrossing.atlanta" = {
      latitude    = 33.7488
      longitude   = -84.3877
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "colocrossing.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "colocrossing.seattle" = {
      latitude    = 47.6061
      longitude   = -122.3328
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "colocrossing.losangeles" = {
      latitude    = 34.0549
      longitude   = -118.2426
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "colocrossing.sanjose" = {
      latitude    = 37.3387
      longitude   = -121.8853
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for colocrossing"
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
  description = "All datacenters for colocrossing"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
