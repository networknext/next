# ----------------------------------------------------------------------------------------

/*
    See https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs
*/

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = file("~/secrets/terraform-digitalocean.txt")
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# ----------------------------------------------------------------------------------------

resource "digitalocean_firewall" "relay" {

  name = "relay"

  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["${var.vpn_address}/32"]
  }

  inbound_rule {
    protocol         = "udp"
    port_range       = "40000"
    source_addresses = ["0.0.0.0/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0"]
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0"]
  }
}

resource "digitalocean_ssh_key" "relay" {
  name       = "relay"
  public_key = file(var.ssh_public_key_file)
}

resource "digitalocean_droplet" "relay" {
  count     = length(var.relays)
  name      = var.relays[count.index].relay_name
  image     = var.relays[count.index].image
  region    = var.relays[count.index].region
  size      = var.relays[count.index].size
  ssh_keys  = [digitalocean_ssh_key.relay.id]
  user_data = replace(file("./setup_relay.sh"), "$VPN_ADDRESS", var.vpn_address)
}

# ----------------------------------------------------------------------------------------

output "relays" {
  description = "Data for each digitalocean relay setup by Terraform"
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
        "digitalocean",
        "${digitalocean_droplet.relay[i].ipv4_address}:40000",
        "0.0.0.0", 
        "", 
        "${digitalocean_droplet.relay[i].ipv4_address}:22",
        "root",
      ]
    )
  ]
}

# --------------------------------------------------------------------------
