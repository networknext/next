# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }

output "relays" {
  description = "Data for each bare metal relay setup by Terraform"
  value = [
    for i, v in var.relays : zipmap( 
      [
        "relay_name", 
        "datacenter_name",
        "supplier_name", 
        "public_address", 
        "internal_address", 
        "internal_group", 
        "ssh_address", 
        "ssh_user",
      ], 
      [
        var.relays[i].relay_name, 
        var.relays[i].datacenter_name,
        var.relays[i].supplier_name, 
        var.relays[i].public_address, 
        var.relays[i].internal_address, 
        var.relays[i].internal_group, 
        var.relays[i].ssh_address, 
        var.relays[i].ssh_user,
      ]
    )
  ]
}

# --------------------------------------------------------------------------

// note: good bare metal include datapacket.com, colocrossing.com, servers.com, serversaustralia.com.au, oneqode.com
