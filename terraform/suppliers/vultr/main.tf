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
  count    = length(var.relays)
  filter {
    name   = "id"
    values = [var.relays[count.index].plan]
  }
}

data "vultr_os" "relay" {
  count    = length(var.relays)
  filter {
    name   = "name"
    values = [var.relays[count.index].os]
  }
}

resource "vultr_instance" "relay" {
  count       = length(var.relays)
  label       = "#{var.relays[count.index].name"
  region      = var.relays[count.index].region
  plan        = data.vultr_plan.relay[count.index].id
  os_id       = data.vultr_os.relay[count.index].id
  ssh_key_ids = [vultr_ssh_key.relay.id]
  script_id   = vultr_startup_script.setup_relay.id
}

resource "vultr_reserved_ip" "relay" {
  count       = length(var.relays)
  label       = var.relays[count.index].name
  region      = var.relays[count.index].region
  ip_type     = "v4"
  instance_id = vultr_instance.relay[count.index].id
}

output "relays" {
  description = "Data for each vultr relay setup by Terraform"
  value = [for i, v in var.relays : zipmap(["relay_name", "region", "public_address", "internal_address", "plan", "os"], [var.relays[i].name, var.relays[i].region, vultr_instance.relay[i].main_ip, "0.0.0.0", var.relays[i].plan, var.relays[i].os])]
}

# --------------------------------------------------------------------------
