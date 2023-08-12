
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

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    networknext = {
      source = "networknext/networknext"
      version = "~> 5.0.2"
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
  source              = "../../suppliers/google"
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
  source              = "../../suppliers/akamai"
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
  source              = "../../suppliers/vultr"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

# ----------------------------------------------------------------------------------------

# =======================
# INITIALIZE DEV DATABASE
# =======================

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

  seller_names = distinct([for k, relay in local.relays : relay.supplier_name])

  sellers = {
    for seller_name in local.seller_names: 
      seller_name => true
  }

  datacenter_names = distinct([for k, relay in local.relays : relay.datacenter_name])
}

resource "networknext_seller" "sellers" {
  for_each = local.sellers
  name     = each.key
}

data "networknext_sellers" "test" {
  depends_on = [
    networknext_seller.sellers,
  ]
}

locals {
  seller_map = {
    for seller in networknext_seller.sellers: 
      seller.name => seller
  }
}

output "sellers" {
  value = data.networknext_sellers.test
}

output "seller_map" {
  value = local.seller_map
}

resource "networknext_datacenter" "datacenters" {
  for_each = local.datacenters
  name = each.key
  seller_id = local.seller_map[each.value.seller_name]
  latitude = each.value.latitude
  longitude = each.value.longitude
  native_name = each.value.native_name
}

data "networknext_datacenters" "test" {
  depends_on = [
    networknext_datacenter.datacenters,
  ]
}

output "datacenters" {
  value = data.networknext_datacenters.test
}





/*
resource "networknext_customer" "test" {
  name = "Test Customer"
  code = "test"
  debug = true
}

resource "networknext_seller" "test" {
  name = "test"
}

resource "networknext_datacenter" "test" {
  name = "test"
  seller_id = networknext_seller.test.id
  latitude = 100
  longitude = 50
}

resource "networknext_relay_keypair" "test" {}

resource "networknext_relay" "test" {
  name = "test.relay"
  datacenter_id = networknext_datacenter.test.id
  public_ip = "127.0.0.1"
  public_key_base64=networknext_relay_keypair.test.public_key_base64
  private_key_base64=networknext_relay_keypair.test.private_key_base64
}

resource "networknext_route_shader" test {
  name = "test"
}

resource "networknext_buyer_keypair" "test" {}

resource "networknext_buyer" "test" {
  name = "Test Buyer"
  customer_id = networknext_customer.test.id
  route_shader_id = networknext_route_shader.test.id
  public_key_base64 = networknext_buyer_keypair.test.public_key_base64
}

resource "networknext_buyer_datacenter_settings" "test" {
  buyer_id = networknext_buyer.test.id
  datacenter_id = networknext_datacenter.test.id
  enable_acceleration = true
}

output "relay_names" {
  description = "Relay names"
  value = local.relay_names
}

output "seller_names" {
  description = "Seller names"
  value = local.seller_names
}

output "datacenter_names" {
  description = "Datacenter names"
  value = local.datacenter_names
}
*/

# ----------------------------------------------------------------------------------------
