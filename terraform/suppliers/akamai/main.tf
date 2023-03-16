# --------------------------------------------------------------------------

terraform {
  required_providers {
    linode = {
      source = "linode/linode"
      version = "1.27.1"
    }
  }
}

provider "linode" {
  token = file("~/secrets/terraform-akamai.txt")
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# --------------------------------------------------------------------------

resource "linode_firewall" "relays" {

  label      = "relays"

  inbound {
    label    = "allow-ssh"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "22"
    ipv4     = ["0.0.0.0/0"]
  }

  inbound {
    label    = "allow-udp"
    action   = "ACCEPT"
    protocol = "UDP"
    ports    = "40000"
    ipv4     = ["0.0.0.0/0"]
  }

  inbound_policy = "DROP"

  outbound_policy = "ACCEPT"

  tags = ["relay"]
}

resource "linode_stackscript" "setup_relay" {
  label = "setup-relay"
  description = "Set up relay"
  script = file("./setup_relay.sh")
  images = ["linode/ubuntu22.04"]
}

resource "linode_instance" "relay" {
  for_each        = var.relays
  image           = each.value.image
  label           = each.key
  region          = local.datacenter_map[each.value.datacenter_name].zone
  type            = each.value.type
  tags            = ["relay"]
  group           = "relays"
  authorized_keys = [replace(file(var.ssh_public_key_file), "\n", "")]
  stackscript_id  = linode_stackscript.setup_relay.id
  lifecycle {
    create_before_destroy = true
  }
}

# ----------------------------------------------------------------------------------------

output "relays" {

  description = "Data for each akamai relay setup by Terraform"

  value = {
    for k, v in var.relays : k => zipmap( 
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
        k,
        v.datacenter_name,
        "akamai", 
        "${linode_instance.relay[k].ip_address}:40000",
        "0.0.0.0",
        "", 
        "${linode_instance.relay[k].ip_address}:22",
        "root",
      ]
    )
  }
}

# ----------------------------------------------------------------------------------------
