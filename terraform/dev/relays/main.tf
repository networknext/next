
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

# ==========
# BARE METAL
# ==========

locals {

  bare_metal_relays = {

    "datapacket.losangeles.a" = {
      datacenter_name  = "datapacket.losangeles"
      supplier_name    = "datapacket"
      public_address   = "127.0.0.1:40000"
      internal_address = "0.0.0.0"
      internal_group   = ""
      ssh_address      = "127.0.0.1:22"
      ssh_user         = "ubuntu"
    }
  }
}

module "bare_metal_relays" {
  relays = local.bare_metal_relays
  source = "../../suppliers/bare_metal"
}

output "bare_metal_relays" {
  description = "Data for each bare metal relay"
  value = module.bare_metal_relays.relays
}

# ----------------------------------------------------------------------------------------

# ============
# GOOGLE CLOUD
# ============

locals {

  google_credentials = "~/secrets/terraform-relays.json"
  google_project     = "relays-380114"
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

output "google_relays" {
  description = "Data for each google relay"
  value = module.google_relays.relays
}

# ----------------------------------------------------------------------------------------

# ============
# AMAZON CLOUD
# ============

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

output "amazon_relays" {
  description = "Data for each amazon relay"
  value = module.amazon_relays.relays
}

# ----------------------------------------------------------------------------------------

# ===================
# AKAMAI (linode.com)
# ===================

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

output "akamai_relays" {
  description = "Data for each akamai relay"
  value = module.akamai_relays.relays
}

# ----------------------------------------------------------------------------------------

# =====
# VULTR
# =====

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

output "vultr_relays" {
  description = "Data for each vultr relay"
  value = module.vultr_relays.relays
}

# ----------------------------------------------------------------------------------------
