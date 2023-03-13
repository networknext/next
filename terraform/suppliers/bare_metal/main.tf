# ----------------------------------------------------------------------------------------

/*
  Good bare metal providers:

    1. datapacket.com
    2. colocrossing.com
    3. servers.com
    4. performive.com
    5. serversaustralia.com.au
    6. velia.net
    7. oneqode.com
*/

# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

output "relays" {
  description = "Data for each bare metal relay setup by Terraform"
  value = {
    for k, v in var.relays : k => zipmap( 
      [
        "relay_name", 
        "native_name",
        "datacenter_name",
        "supplier_name", 
        "public_address", 
        "internal_address", 
        "internal_group", 
        "ssh_address", 
        "ssh_user",
      ], 
      [
        k,
        v.native_name,
        v.datacenter_name,
        v.supplier_name, 
        v.public_address, 
        v.internal_address, 
        v.internal_group, 
        v.ssh_address, 
        v.ssh_user,
      ]
    )
  }
}

# --------------------------------------------------------------------------
