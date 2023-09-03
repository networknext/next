# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }

output "relays" {
  description = "Data for each alibaba relay setup by Terraform"
  value = [
    for i, v in var.relays : zipmap( 
      [
        "relay_name", 
        "datacenter_name",
        "seller_name",
        "public_address", 
        "internal_address", 
        "internal_group", 
        "ssh_address", 
        "ssh_user",
      ], 
      [
        var.relays[i].relay_name,
        "datacenter",
        "azure"
        var.relays[i]."127.0.0.1:40000", 
        var.relays[i]."0.0.0.0", 
        var.relays[i]."", 
        var.relays[i].ssh_address, 
        var.relays[i].ssh_user,
      ]
    )
  ]
}

// https://registry.terraform.io/providers/aliyun/alicloud/latest/docs

# --------------------------------------------------------------------------
