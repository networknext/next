# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "<Your seller>"

  seller_code = "<seller>"

  ssh_user = "root"

  datacenter_map = {

    "seller.cityname" = {
      latitude    = 10.00
      longitude   = 20.00
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    }

  }

}

output "relays" {
  description = "All relays for <seller>"
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
  description = "All datacenters for <seller>"
  value = locals.datacenter_map
}

# --------------------------------------------------------------------------
