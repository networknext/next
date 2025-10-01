# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "OVH"

  seller_code = "ovh"

  ssh_user = "ubuntu"

  datacenter_map = {

    "ovh.ashburn" = {
      latitude    = 39.0438
      longitude   = -77.4874
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

    "ovh.frankfurt" = {
      latitude    = 50.1109
      longitude   = 8.6821
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

  }
}

output "relays" {
  description = "All relays for ovh"
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
  description = "All datacenters for ovh"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
