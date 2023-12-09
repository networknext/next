# ----------------------------------------------------------------------------------------

// Stackpath now supports a terraform provider
// https://registry.terraform.io/providers/stackpath/stackpath/latest/docs

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Stackpath"

  seller_code = "stackpath"

  ssh_user = "root"

  datacenter_map = {

    "stackpath." = {
      latitude    = 
      longitude   = 
      native_name = ""
      seller_name = local.seller_name
      seller_code = local.seller_code
    },

/*
ashburn
chicago
newyork
atlanta
miami
dallas
denver
detroit
boston
stlouis
minneapolis
losangeles
seattle
sanjose
dublin
manchester
amsterdam
frankfurt
london
madrid
warsaw
stockholm
hongkong
singapore
tokyo
dubai
mumbai
saopaulo
melbourne
*/

  }
}

output "relays" {
  description = "All relays for zenlayer"
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
  description = "All datacenters for zenlayer"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
