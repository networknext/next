# ========================================================================================
#                                       DEV RELAYS
# ========================================================================================

variable "vpn_address" { type = string }
variable "ssh_public_key_file" { type = string }
variable "ssh_private_key_file" { type = string }
variable "env" { type = string }
variable "relay_version" { type = string }
variable "relay_artifacts_bucket" { type = string }
variable "relay_backend_url" { type = string }
variable "relay_backend_public_key" { type = string }
variable "sellers" { type = map(string) }
variable "raspberry_buyer_public_key" { type = string }
variable "raspberry_datacenters" { type = list(string) }
variable "test_buyer_public_key" { type = string }
variable "test_datacenters" { type = list(string) }

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    networknext = {
      source = "networknext/networknext"
      version = "~> 5.0.6"
    }
  }
  backend "gcs" {
    bucket  = "newyork_network_next_terraform"
    prefix  = "dev_relays"
  }
}

provider "networknext" {
  hostname = "https://api-dev.virtualgo.net"
  api_key  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwicG9ydGFsIjp0cnVlLCJpc3MiOiJuZXh0IGtleWdlbiIsImlhdCI6MTcwMTcxMzk2N30.wExR2URUfD5fIArq-QntcmlC36K1M5iOaoTUEtrACJQ"
}

# ----------------------------------------------------------------------------------------

# =============
# GOOGLE RELAYS
# =============

locals {

  google_credentials = "~/secrets/terraform-dev-relays.json"
  google_project     = file("../../projects/dev-relays-project-id.txt")
  google_relays = {

    "google.iowa.1" = {
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

    "google.ohio.1" = {
      datacenter_name = "google.ohio.1"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.ohio.2" = {
      datacenter_name = "google.ohio.2"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.ohio.3" = {
      datacenter_name = "google.ohio.3"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.virginia.1" = {
      datacenter_name = "google.virginia.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.virginia.2" = {
      datacenter_name = "google.virginia.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.virginia.3" = {
      datacenter_name = "google.virginia.3"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.frankfurt.1" = {
      datacenter_name = "google.frankfurt.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.frankfurt.2" = {
      datacenter_name = "google.frankfurt.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.frankfurt.3" = {
      datacenter_name = "google.frankfurt.3"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.finland.1" = {
      datacenter_name = "google.finland.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.finland.2" = {
      datacenter_name = "google.finland.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.finland.3" = {
      datacenter_name = "google.finland.3"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.netherlands.1" = {
      datacenter_name = "google.netherlands.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.netherlands.2" = {
      datacenter_name = "google.netherlands.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.netherlands.3" = {
      datacenter_name = "google.netherlands.3"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.london.1" = {
      datacenter_name = "google.london.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.london.2" = {
      datacenter_name = "google.london.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.london.3" = {
      datacenter_name = "google.london.3"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

/*
    # IOWA

    "google.iowa.1" = {
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

    # LOS ANGELES

    "google.losangeles.1" = {
      datacenter_name = "google.losangeles.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    # SALT LAKE CITY

    "google.saltlakecity.1" = {
      datacenter_name = "google.saltlakecity.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    # SOUTH CAROLINA

    "google.southcarolina.2" = {
      datacenter_name = "google.southcarolina.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    # VIRGINIA

    "google.virginia.1" = {
      datacenter_name = "google.virginia.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    # DALLAS

    "google.dallas.1" = {
      datacenter_name = "google.dallas.1"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    # OHIO

    "google.ohio.1" = {
      datacenter_name = "google.ohio.1"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
*/

  }
}

module "google_relays" {
  relays              = local.google_relays
  project             = local.google_project
  credentials         = local.google_credentials
  source              = "../../sellers/google"
  vpn_address         = var.vpn_address
  ssh_public_key_file = "~/secrets/next_ssh.pub"
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
  # Please edit sellers/amazon.go for the set of dev relays, then run "next config" to generate amazon/generated.tf

  config              = local.amazon_config
  credentials         = local.amazon_credentials
  profile             = local.amazon_profile
  source              = "./amazon"
  vpn_address         = var.vpn_address
  ssh_public_key_file = "~/secrets/next_ssh.pub"
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

    "akamai.frankfurt" = {
      datacenter_name = "akamai.frankfurt"
      type            = "g6-dedicated-2"
      image           = "linode/ubuntu22.04"
    }
    
    "akamai.london" = {
      datacenter_name = "akamai.london"
      type            = "g6-dedicated-2"
      image           = "linode/ubuntu22.04"
    }

    /*
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
    
    "akamai.dallas" = {
      datacenter_name = "akamai.dallas"
      type            = "g6-dedicated-2"
      image           = "linode/ubuntu22.04"
    }
    */

  }
}

module "akamai_relays" {
  env                 = "dev"
  relays              = local.akamai_relays
  source              = "../../sellers/akamai"
  vpn_address         = var.vpn_address
  ssh_public_key_file = "~/secrets/next_ssh.pub"
}

# ----------------------------------------------------------------------------------------

# =================
# DATAPACKET RELAYS
# =================

locals {

  datapacket_relays = {

    /*
    "datapacket.losangeles" = {
      datacenter_name = "datapacket.losangeles"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "datapacket_relays" {
  relays = local.datapacket_relays
  source = "../../sellers/datapacket"
}

# ----------------------------------------------------------------------------------------

# ==========
# i3D RELAYS
# ==========

locals {

  i3d_relays = {

    /*
    "i3d.losangeles" = {
      datacenter_name = "i3d.losangeles"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "i3d_relays" {
  relays = local.i3d_relays
  source = "../../sellers/i3d"
}

# ----------------------------------------------------------------------------------------

# ==============
# ONEQODE RELAYS
# ==============

locals {

  oneqode_relays = {

    /*
    "oneqode.singapore" = {
      datacenter_name = "oneqode.singapore"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "oneqode_relays" {
  relays = local.oneqode_relays
  source = "../../sellers/oneqode"
}

# ----------------------------------------------------------------------------------------

# ============
# GCORE RELAYS
# ============

locals {

  gcore_relays = {

    /*
    "gcore.frankfurt" = {
      datacenter_name = "gcore.frankfurt"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "gcore_relays" {
  relays = local.gcore_relays
  source = "../../sellers/gcore"
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
      keys(module.datapacket_relays.relays),
      keys(module.i3d_relays.relays),
      keys(module.oneqode_relays.relays),
      keys(module.gcore_relays.relays),
    )
  )

  relays = merge(
    module.google_relays.relays,
    module.amazon_relays.relays,
    module.akamai_relays.relays,
    module.datapacket_relays.relays,
    module.i3d_relays.relays,
    module.oneqode_relays.relays,
    module.gcore_relays.relays,
  )

  datacenters = merge(
    module.google_relays.datacenters,
    module.amazon_relays.datacenters,
    module.akamai_relays.datacenters,
    module.datapacket_relays.datacenters,
    module.i3d_relays.datacenters,
    module.oneqode_relays.datacenters,
    module.gcore_relays.datacenters,
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

# ===============
# RASPBERRY BUYER
# ===============

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
  public_key_base64 = var.raspberry_buyer_public_key
}

resource "networknext_buyer_datacenter_settings" raspberry {
  count = length(var.raspberry_datacenters)
  buyer_id = networknext_buyer.raspberry.id
  datacenter_id = networknext_datacenter.datacenters[var.raspberry_datacenters[count.index]].id
  enable_acceleration = true
}

# ==========
# TEST BUYER
# ==========

resource "networknext_route_shader" test {
  name = "test"
  force_next = true
  acceptable_latency = 50
  latency_reduction_threshold = 10
  route_select_threshold = 0
  route_switch_threshold = 5
  acceptable_packet_loss_instant = 0.25
  acceptable_packet_loss_sustained = 0.1
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
}

resource "networknext_buyer" test {
  name = "Test"
  code = "test"
  debug = true
  live = true
  route_shader_id = networknext_route_shader.test.id
  public_key_base64 = var.test_buyer_public_key
}

resource "networknext_buyer_datacenter_settings" test {
  count = length(var.test_datacenters)
  buyer_id = networknext_buyer.test.id
  datacenter_id = networknext_datacenter.datacenters[var.raspberry_datacenters[count.index]].id
  enable_acceleration = true
}

# ----------------------------------------------------------------------------------------
