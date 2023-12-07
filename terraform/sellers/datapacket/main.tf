# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "Datapacket"

  seller_code = "datapacket"

  ssh_user = "root"


  stockholm
  copenhagen
  dublin
  london
  amsterdam
  warsaw
  kyiv
  london
  frankfurt
  brussels
  zurich
  prague
  bratislava
  zurich
  paris
  vienna
  bucharest
  milan
  zagreb
  sofia
  lisbon
  madrid
  marseille
  palermo
  athens
  istanbul
  telaviv
  johanesburg
  sydney
  singapore
  hongkong
  tokyo


  datacenter_map = {

    "datapacket.losangeles" = {
      latitude    = 34.0522
      longitude   = -118.2437
    },

    "datapacket.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
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
