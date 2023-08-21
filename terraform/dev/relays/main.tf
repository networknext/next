
# ========================================================================================
#                                       DEV RELAYS
# ========================================================================================

variable "vpn_address" { type = string }
variable "ssh_public_key_file" { type = string }
variable "ssh_private_key_file" { type = string }
variable "env" { type = string }
variable "relay_version" { type = string }
variable "relay_artifacts_bucket" { type = string }
variable "relay_public_key" { type = string }
variable "relay_private_key" { type = string }
variable "relay_backend_hostname" { type = string }
variable "relay_backend_public_key" { type = string }
variable "raspberry_datacenters" { type = list(string) }
variable "sellers" { type = map(string) }

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    networknext = {
      source = "networknext/networknext"
      version = "~> 5.0.6"
    }
  }
}

provider "networknext" {
  hostname = "http://dev.virtualgo.net"
  api_key  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZGF0YWJhc2UiOnRydWUsInBvcnRhbCI6dHJ1ZX0.QFPdb-RcP8wyoaOIBYeB_X6uA7jefGPVxm2VevJvpwU"
}

# ----------------------------------------------------------------------------------------

# =============
# GOOGLE RELAYS
# =============

locals {

  google_credentials = "~/secrets/terraform-development-relays.json"
  google_project     = "development-relays"
  google_relays = {

    # IOWA

    "google.iowa.1.a" = {
      datacenter_name = "google.iowa.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.1.b" = {
      datacenter_name = "google.iowa.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.2" = {
      datacenter_name = "google.iowa.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.3" = {
      datacenter_name = "google.iowa.3"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.6" = {
      datacenter_name = "google.iowa.6"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    # OREGON

    "google.oregon.1" = {
      datacenter_name = "google.oregon.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    # SALT LAKE CITY

    "google.saltlakecity.1" = {
      datacenter_name = "google.saltlakecity.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.saltlakecity.2" = {
      datacenter_name = "google.saltlakecity.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
  }
}

module "google_relays" {
  relays              = local.google_relays
  project             = local.google_project
  credentials         = local.google_credentials
  source              = "../../sellers/google"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

# ----------------------------------------------------------------------------------------

# =============
# AMAZON RELAYS
# =============

locals {
  amazon_config      = ["~/.aws/config"]
  amazon_credentials = ["~/.aws/credentials"]
  amazon_profile     = "default"
}

module "amazon_relays" {
  # IMPORTANT: It is LITERALLY IMPOSSIBLE to work with multiple AWS regions programmatically in Terraform
  # So for AWS, see tools/amazon_config/amazon_config.go for the set of dev relays -> amazon/generated.tf
  config              = local.amazon_config
  credentials         = local.amazon_credentials
  profile             = local.amazon_profile
  source              = "./amazon"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

# ----------------------------------------------------------------------------------------

# =============
# AKAMAI RELAYS
# =============

locals {

  akamai_relays = {

    "akamai.newyork" = {
      datacenter_name = "akamai.newyork"
      type            = "g6-dedicated-2"
      image           = "linode/ubuntu22.04"
    },

    "akamai.atlanta" = {
      datacenter_name = "akamai.atlanta"
      type            = "g6-dedicated-2"
      image           = "linode/ubuntu22.04"
    },

    "akamai.fremont" = {
      datacenter_name = "akamai.fremont"
      type            = "g6-dedicated-2"
      image           = "linode/ubuntu22.04"
    }
    
    "akamai.frankfurt" = {
      datacenter_name = "akamai.frankfurt"
      type            = "g6-dedicated-2"
      image           = "linode/ubuntu22.04"
    },
  }
}

module "akamai_relays" {
  relays              = local.akamai_relays
  source              = "../../sellers/akamai"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

# ----------------------------------------------------------------------------------------

# ============
# VULTR RELAYS
# ============

locals {

  vultr_relays = {

    "vultr.seattle" = {
      datacenter_name = "vultr.seattle"
      plan            = "vc2-1c-1gb"
      os              = "Ubuntu 22.04 LTS x64"
    },

    "vultr.chicago" = {
      datacenter_name = "vultr.chicago"
      plan            = "vc2-1c-1gb"
      os              = "Ubuntu 22.04 LTS x64"
    },

  }
}

module "vultr_relays" {
  relays              = local.vultr_relays
  source              = "../../sellers/vultr"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

# ----------------------------------------------------------------------------------------

# =======================
# INITIALIZE DEV DATABASE
# =======================

# Setup sellers, datacenters and relays in dev

locals {
  
  relay_names = sort(
    concat(
      keys(module.google_relays.relays),
      keys(module.amazon_relays.relays),
      keys(module.akamai_relays.relays),
      keys(module.vultr_relays.relays),
    )
  )

  relays = merge(
    module.google_relays.relays,
    module.amazon_relays.relays,
    module.akamai_relays.relays,
    module.vultr_relays.relays,
  )

  datacenters = merge(
    module.google_relays.datacenters,
    module.amazon_relays.datacenters,
    module.akamai_relays.datacenters,
    module.vultr_relays.datacenters,
  )

  datacenter_names = distinct([for k, relay in local.relays : relay.datacenter_name])
}

resource "networknext_seller" sellers {
  for_each = var.sellers
  name     = each.key
  code     = each.value
}

locals {
  seller_map = {
    for seller in networknext_seller.sellers: 
      seller.code => seller
  }
}

resource "networknext_datacenter" datacenters {
  for_each = local.datacenters
  name = each.key
  seller_id = local.seller_map[each.value.seller_code].id
  latitude = each.value.latitude
  longitude = each.value.longitude
  native_name = each.value.native_name
}

locals {
  datacenter_map = {
    for datacenter in networknext_datacenter.datacenters:
      datacenter.name => datacenter
  }
}

resource "networknext_relay_keypair" relay_keypairs {
  for_each = local.relays
}

resource "networknext_relay" relays {
  for_each = local.relays
  name = each.key
  datacenter_id = local.datacenter_map[each.value.datacenter_name].id
  public_key_base64=networknext_relay_keypair.relay_keypairs[each.key].public_key_base64
  private_key_base64=networknext_relay_keypair.relay_keypairs[each.key].private_key_base64
  public_ip = each.value.public_ip
  public_port = each.value.public_port
  internal_ip = each.value.internal_ip
  internal_port = each.value.internal_port
  internal_group = each.value.internal_group
  ssh_ip = each.value.ssh_ip
  ssh_port = each.value.ssh_port
  ssh_user = each.value.ssh_user
  version = var.relay_version
}

# ----------------------------------------------------------------------------------------

# Print out set of relays in the database

data "networknext_relays" relays {
  depends_on = [
    networknext_relay.relays,
  ]
}

locals {
  database_relays = {
    for k,v in networknext_relay.relays: k => v.public_ip
  }
}

output "database_relays" {
  value = local.database_relays
}

output "all_relays" {
  value = networknext_relay.relays
}

# ----------------------------------------------------------------------------------------

# Setup the raspberry buyer

resource "networknext_route_shader" raspberry {
  name = "raspberry"
  force_next = true
  route_select_threshold = 300
  route_switch_threshold = 300
}

resource "networknext_buyer" raspberry {
  name = "Raspberry"
  code = "raspberry"
  debug = true
  live = true
  route_shader_id = networknext_route_shader.raspberry.id
  public_key_base64 = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
}

resource "networknext_buyer_datacenter_settings" raspberry {
  count = length(var.raspberry_datacenters)
  buyer_id = networknext_buyer.raspberry.id
  datacenter_id = networknext_datacenter.datacenters[var.raspberry_datacenters[count.index]].id
  enable_acceleration = true
}

# ----------------------------------------------------------------------------------------
