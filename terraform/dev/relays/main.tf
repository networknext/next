# ========================================================================================
#                                       DEV RELAYS
# ========================================================================================

locals {
  
  env                         = "dev"
  vpn_address                 = "45.79.157.168"
  ssh_public_key_file         = "~/secrets/next_ssh.pub"
  ssh_private_key_file        = "~/secrets/next_ssh"
  relay_version               = "relay-144"
  relay_artifacts_bucket      = "sloclap_network_next_relay_artifacts"
  relay_backend_public_key    = "Z+9puZkCkV03nm4yO49ySF+H181jAlWVy7JPGMlk10I="
  relay_backend_url           = "relay-dev.virtualgo.net"

  raspberry_buyer_public_key  = "gtdzp3hCfJ9Y+6OOpsWoMChMXhXGDRnY7vkFdHwNqVW0bdp6jjTx6Q=="

  raspberry_datacenters = [
    "google.iowa.1",
    "google.iowa.2",
    "google.iowa.3",
    "google.iowa.6"
  ]

  test_buyer_public_key       = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ=="

  test_datacenters = [
    "google.saopaulo.1",
    "google.saopaulo.2",
    "google.saopaulo.3",
  ]

  sellers = {
    "Akamai" = "akamai"
    "Amazon" = "amazon"
    "Google" = "google"
    "Datapacket" = "datapacket"
    "i3D" = "i3d"
    "Oneqode" = "oneqode"
    "GCore" = "gcore"
    "Hivelocity" = "hivelocity"
    "ColoCrossing" = "colocrossing"
    "phoenixNAP" = "phoenixnap"
    "servers.com" = "serversdotcom"
    "Velia" = "velia"
    "Zenlayer" = "zenlayer"
    "Latitude" = "latitude"
    "Equinix" = "equinix"
  }  
}

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    networknext = {
      source = "networknext/networknext"
      version = "~> 5.0.13"
    }
  }
  backend "gcs" {
    bucket  = "sloclap_network_next_terraform"
    prefix  = "dev_relays"
  }
}

provider "networknext" {
  hostname = "https://api-dev.virtualgo.net"
  api_key  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwicG9ydGFsIjp0cnVlLCJpc3MiOiJuZXh0IGtleWdlbiIsImlhdCI6MTc0OTczODE4OX0.c2mPjSWM_CSw8Z7SqvW7CRRFIY5kxh5DMHkQxgo5TQE"
}

# ----------------------------------------------------------------------------------------

# =============
# GOOGLE RELAYS
# =============

locals {

  google_credentials = "~/secrets/terraform-dev-relays.json"
  google_project     = file("~/secrets/dev-relays-project-id.txt")
  google_relays = {

    "google.iowa.1" = {
      datacenter_name = "google.iowa.1"
      type            = "c3-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.2" = {
      datacenter_name = "google.iowa.2"
      type            = "c3-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.3" = {
      datacenter_name = "google.iowa.3"
      type            = "c3-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.6" = {
      datacenter_name = "google.iowa.6"
      type            = "c3-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.ohio.1" = {
      datacenter_name = "google.ohio.1"
      type            = "c3-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.ohio.2" = {
      datacenter_name = "google.ohio.2"
      type            = "c3-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.ohio.3" = {
      datacenter_name = "google.ohio.3"
      type            = "c3-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

  }
}

module "google_relays" {
  env                 = "dev"
  relays              = local.google_relays
  project             = local.google_project
  credentials         = local.google_credentials
  source              = "../../sellers/google"
  vpn_address         = local.vpn_address
  ssh_public_key_file         = "~/secrets/next_ssh.pub"
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

  # Please edit sellers/amazon.go for the set of dev relays, then run "next config" to update amazon/generated.tf

  config              = local.amazon_config
  credentials         = local.amazon_credentials
  profile             = local.amazon_profile
  source              = "./amazon"
  vpn_address         = local.vpn_address
  ssh_public_key_file         = "~/secrets/next_ssh.pub"
}

# ----------------------------------------------------------------------------------------

# =============
# AKAMAI RELAYS
# =============

locals {

  akamai_relays = {

    "akamai.newyork" = {
      datacenter_name = "akamai.newyork"
      type            = "g6-dedicated-4"
      image           = "linode/ubuntu22.04"
    },

  }
}

module "akamai_relays" {
  env                 = "dev"
  relays              = local.akamai_relays
  source              = "../../sellers/akamai"
  vpn_address         = local.vpn_address
  ssh_public_key_file         = "~/secrets/next_ssh.pub"
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
      keys(module.hivelocity_relays.relays),
      keys(module.colocrossing_relays.relays),
      keys(module.phoenixnap_relays.relays),
      keys(module.serversdotcom_relays.relays),
      keys(module.velia_relays.relays),
      keys(module.zenlayer_relays.relays),
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
    module.serversdotcom_relays.relays,
    module.velia_relays.relays,
    module.zenlayer_relays.relays,
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
    module.serversdotcom_relays.datacenters,
    module.velia_relays.datacenters,
    module.zenlayer_relays.datacenters,
    module.latitude_relays.datacenters,
    module.equinix_relays.datacenters,
  )

  datacenter_names = distinct([for k, relay in local.relays : relay.datacenter_name])
}

resource "networknext_seller" sellers {
  for_each = local.sellers
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
  version = local.relay_version
  bandwidth_price = each.value.bandwidth_price
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
  public_key_base64 = local.raspberry_buyer_public_key
}

resource "networknext_buyer_datacenter_settings" raspberry {
  for_each = toset(local.raspberry_datacenters)
  buyer_id = networknext_buyer.raspberry.id
  datacenter_id = networknext_datacenter.datacenters[each.value].id
  enable_acceleration = true
}

# ==========
# TEST BUYER
# ==========

resource "networknext_route_shader" test {
  name = "test"
  force_next = true
  latency_reduction_threshold = 1
  acceptable_latency = 0
  acceptable_packet_loss_instant = 100
  acceptable_packet_loss_sustained = 100
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
  route_select_threshold = 1
  route_switch_threshold = 10
}

resource "networknext_buyer" test {
  name = "Test"
  code = "test"
  debug = true
  live = true
  route_shader_id = networknext_route_shader.test.id
  public_key_base64 = local.test_buyer_public_key
}

resource "networknext_buyer_datacenter_settings" test {
  for_each = toset(local.test_datacenters)
  buyer_id = networknext_buyer.test.id
  datacenter_id = networknext_datacenter.datacenters[each.value].id
  enable_acceleration = true
}

# ----------------------------------------------------------------------------------------
