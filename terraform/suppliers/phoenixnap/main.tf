# --------------------------------------------------------------------------

terraform {
  required_providers {
    pnap = {
      source = "phoenixnap/pnap"
      version = "0.18.2"
    }
  }
}

provider "pnap" {
  client_id = var.client_id
  client_secret = file("~/Documents/terraform-phoenixnap.txt")
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }
variable "client_id" { type = string }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# ----------------------------------------------------------------------------------------

resource "pnap_server" "relay" {
    hostname      = "relay"
    os            = "ubuntu/jammy"
    type          = "s1.c1.medium"
    location      = "PHX"
    pricing_model = "HOURLY"
    ssh_keys      = [file(var.ssh_public_key_file)]
}

# ----------------------------------------------------------------------------------------
