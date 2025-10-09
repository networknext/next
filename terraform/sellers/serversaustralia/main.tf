# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "ServersAustralia"

  seller_code = "serversaustralia"

  ssh_user = "root"

  datacenter_map = {

    "serversaustralia.sydney" = {
      latitude    = -33.8727
      longitude   = 151.2057
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for serversaustralia"
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
  description = "All datacenters for serversaustralia"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
