# --------------------------------------------------------------------------

terraform {
  required_providers {
    vultr = {
      source = "vultr/vultr"
      version = "2.12.1"
    }
  }
}

provider "vultr" {
  api_key = file("~/Documents/terraform-vultr.txt")
  rate_limit = 100
  retry_limit = 3
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# --------------------------------------------------------------------------

resource "vultr_ssh_key" "relay" {
  name    = "relay"
  ssh_key = replace(file(var.ssh_public_key_file), "\n", "")
}

resource "vultr_startup_script" "setup_relay" {
    name   = "setup-relay"
    script = base64encode(file("./setup_relay.sh"))
}

data "vultr_plan" "relay" {
  filter {
    name   = "id"
    values = ["vc2-1c-1gb"]
  }
}

data "vultr_os" "relay" {
  filter {
    name   = "name"
    values = ["Ubuntu 22.04 LTS x64"]
  }
}

resource "vultr_instance" "relay" {
  plan        = data.vultr_plan.relay.id
  region      = "sea"
  os_id       = 167
  label       = "relay"
  ssh_key_ids = [vultr_ssh_key.relay.id]
  script_id   = vultr_startup_script.setup_relay.id
}

/*
output "relays" {
  description = "Data for each akamai relay setup by Terraform"
  value = [for i, v in var.relays : zipmap(["relay_name", "region", "public_address", "internal_address", "type", "image"], [var.relays[i].name, var.relays[i].region, linode_instance.relay[i].ip_address, "0.0.0.0", var.relays[i].type, var.relays[i].image])]
}
*/

# --------------------------------------------------------------------------
