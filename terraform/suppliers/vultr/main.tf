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
  api_key = file("~/secrets/terraform-vultr.txt")
  rate_limit = 100
  retry_limit = 3
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }
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
  for_each = var.relays
  filter {
    name   = "id"
    values = [each.value.plan]
  }
}

data "vultr_os" "relay" {
  for_each = var.relays
  filter {
    name   = "name"
    values = [each.value.os]
  }
}

resource "vultr_instance" "relay" {
  for_each    = var.relays
  label       = each.key
  region      = local.datacenter_map[each.value.datacenter_name].zone
  plan        = data.vultr_plan.relay[each.key].id
  os_id       = data.vultr_os.relay[each.key].id
  ssh_key_ids = [vultr_ssh_key.relay.id]
  script_id   = vultr_startup_script.setup_relay.id
}

resource "vultr_reserved_ip" "relay" {
  for_each    = var.relays
  label       = each.key
  region      = local.datacenter_map[each.value.datacenter_name].zone
  ip_type     = "v4"
  instance_id = vultr_instance.relay[each.key].id
}

# ----------------------------------------------------------------------------------------

output "relays" {

  description = "Data for each vultr relay setup by Terraform"

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
        "vultr", 
        "${vultr_instance.relay[k].main_ip}:40000",
        "0.0.0.0",
        "", 
        "${vultr_instance.relay[k].main_ip}:22",
        "root",
      ]
    )
  }
}

output "datacenters" {
  description = "Data for each vultr datacenter"
  value = local.datacenter_map
}

# ----------------------------------------------------------------------------------------
