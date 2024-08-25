# ========================================================================================
#                                      PROD RELAYS
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
    bucket  = "theodore_network_next_terraform"
    prefix  = "prod_relays"
  }
}

provider "networknext" {
  hostname = "https://api.virtualgo.net"
  api_key  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwicG9ydGFsIjp0cnVlLCJpc3MiOiJuZXh0IGtleWdlbiIsImlhdCI6MTcxMDM0NTUxNX0.xq33dxiVZ9V4sYAG9hX6k_ijXqED9_MCp9us7mycneo"
}

# ----------------------------------------------------------------------------------------

# =============
# GOOGLE RELAYS
# =============

locals {

  google_credentials = "~/secrets/terraform-prod-relays.json"
  google_project     = file("../../projects/prod-relays-project-id.txt")
  google_relays = {

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

  }
}

module "google_relays" {
  env                 = "prod"
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
  # So for AWS, see sellers/amazon.go for the set of prod relays -> amazon/generated.tf

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

  }
}

module "akamai_relays" {
  env                 = "prod"
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

# =================
# HIVELOCITY RELAYS
# =================

locals {

  hivelocity_relays = {

    /*
    "hivelocity.chicago" = {
      datacenter_name = "hivelocity.chicago"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "hivelocity_relays" {
  relays = local.hivelocity_relays
  source = "../../sellers/hivelocity"
}

# ----------------------------------------------------------------------------------------

# ===================
# COLOCROSSING RELAYS
# ===================

locals {

  colocrossing_relays = {

    /*
    "colocrossing.chicago" = {
      datacenter_name = "colocrossing.chicago"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "colocrossing_relays" {
  relays = local.colocrossing_relays
  source = "../../sellers/colocrossing"
}

# ----------------------------------------------------------------------------------------

# =================
# PHOENIXNAP RELAYS
# =================

locals {

  phoenixnap_relays = {

    /*
    "phoenixnap.phoenix" = {
      datacenter_name = "phoenixnap.phoenix"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "phoenixnap_relays" {
  relays = local.phoenixnap_relays
  source = "../../sellers/phoenixnap"
}

# ----------------------------------------------------------------------------------------

# ===========
# INAP RELAYS
# ===========

locals {

  inap_relays = {

    /*
    "inap.chicago" = {
      datacenter_name = "inap.chicago"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "inap_relays" {
  relays = local.inap_relays
  source = "../../sellers/inap"
}

# ----------------------------------------------------------------------------------------

# ==================
# SERVERS.COM RELAYS
# ==================

locals {

  serversdotcom_relays = {

    /*
    "serversdotcom.dallas" = {
      datacenter_name = "serversdotcom.dallas"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "serversdotcom_relays" {
  relays = local.serversdotcom_relays
  source = "../../sellers/serversdotcom"
}

# ----------------------------------------------------------------------------------------

# ============
# VELIA RELAYS
# ============

locals {

  velia_relays = {

    /*
    "velia.stlouis" = {
      datacenter_name = "velia.stlouis"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "velia_relays" {
  relays = local.velia_relays
  source = "../../sellers/velia"
}

# ----------------------------------------------------------------------------------------

# ===============
# ZENLAYER RELAYS
# ===============

locals {

  zenlayer_relays = {

    /*
    "zenlayer.singapore" = {
      datacenter_name = "zenlayer.singapore"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "zenlayer_relays" {
  relays = local.zenlayer_relays
  source = "../../sellers/zenlayer"
}

# ----------------------------------------------------------------------------------------

# ================
# STACKPATH RELAYS
# ================

locals {

  stackpath_relays = {

    /*
    "stackpath.singapore" = {
      datacenter_name = "stackpath.singapore"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "stackpath_relays" {
  relays = local.stackpath_relays
  source = "../../sellers/stackpath"
}

# ----------------------------------------------------------------------------------------

# ===============
# LATITUDE RELAYS
# ===============

locals {

  latitude_relays = {

    /*
    "latitude.buenosaires" = {
      datacenter_name = "latitude.buenosaires"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "latitude_relays" {
  relays = local.latitude_relays
  source = "../../sellers/latitude"
}

# ----------------------------------------------------------------------------------------

# ==============
# EQUINIX RELAYS
# ==============

locals {

  equinix_relays = {

    /*
    "equinix.miami" = {
      datacenter_name = "equinix.miami"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "equinix_relays" {
  relays = local.equinix_relays
  source = "../../sellers/equinix"
}

# ----------------------------------------------------------------------------------------

# ========================
# INITIALIZE PROD DATABASE
# ========================

# Setup sellers, datacenters and relays in prod

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
      keys(module.hivelocity_relays.relays),
      keys(module.colocrossing_relays.relays),
      keys(module.phoenixnap_relays.relays),
      keys(module.inap_relays.relays),
      keys(module.serversdotcom_relays.relays),
      keys(module.velia_relays.relays),
      keys(module.zenlayer_relays.relays),
      keys(module.stackpath_relays.relays),
      keys(module.latitude_relays.relays),
      keys(module.equinix_relays.relays),
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
    module.hivelocity_relays.relays,
    module.colocrossing_relays.relays,
    module.phoenixnap_relays.relays,
    module.inap_relays.relays,
    module.serversdotcom_relays.relays,
    module.velia_relays.relays,
    module.zenlayer_relays.relays,
    module.stackpath_relays.relays,
    module.latitude_relays.relays,
    module.equinix_relays.relays,
  )

  datacenters = merge(
    module.google_relays.datacenters,
    module.amazon_relays.datacenters,
    module.akamai_relays.datacenters,
    module.datapacket_relays.datacenters,
    module.i3d_relays.datacenters,
    module.oneqode_relays.datacenters,
    module.gcore_relays.datacenters,
    module.hivelocity_relays.datacenters,
    module.colocrossing_relays.datacenters,
    module.phoenixnap_relays.datacenters,
    module.inap_relays.datacenters,
    module.serversdotcom_relays.datacenters,
    module.velia_relays.datacenters,
    module.zenlayer_relays.datacenters,
    module.stackpath_relays.datacenters,
    module.latitude_relays.datacenters,
    module.equinix_relays.datacenters,
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
  acceptable_latency = 0
  latency_reduction_threshold = 1
  acceptable_packet_loss_instant = 0.1
  acceptable_packet_loss_sustained = 0.01
  bandwidth_envelope_up_kbps = 1024
  bandwidth_envelope_down_kbps = 1024
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
