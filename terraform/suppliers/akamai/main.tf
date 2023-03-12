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

variable "relays" { type = list(map(string)) }
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
  count           = length(var.relays)
  image           = var.relays[count.index].image
  label           = var.relays[count.index].name
  region          = var.relays[count.index].region
  type            = var.relays[count.index].type
  tags            = ["relay"]
  group           = "relays"
  authorized_keys = [replace(file(var.ssh_public_key_file), "\n", "")]
  stackscript_id  = linode_stackscript.setup_relay.id
  lifecycle {
    create_before_destroy = true
  }
}

output "relays" {
  description = "Data for each akamai relay setup by Terraform"
  value = [for i, v in var.relays : zipmap(["relay_name", "region", "public_address", "internal_address", "type", "image"], [var.relays[i].name, var.relays[i].region, linode_instance.relay[i].ip_address, "0.0.0.0", var.relays[i].type, var.relays[i].image])]
}

# --------------------------------------------------------------------------
