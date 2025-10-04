# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "UK2Group"

  seller_code = "uk2group"

  ssh_user = "root"

  datacenter_map = {

    "uk2group.dallas" = {
      latitude    = 32.7767
      longitude   = -96.7970
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "uk2group.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for uk2group"
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
  description = "All datacenters for uk2group"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
