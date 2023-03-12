# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

# ----------------------------------------------------------------------------------------

# todo: these two are actually specific to google cloud only...
variable "credentials" { type = string }
variable "project" { type = string }

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

variable "google_relays" { type = list(map(string)) }

module "google_relays" {
  relays              = var.google_relays
  source              = "./google"
  project             = var.project
  credentials         = var.credentials
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "google_relays" {
  description = "Data for each google relay"
  value = module.google_relays.relays
}

# ----------------------------------------------------------------------------------------

variable "amazon_relays" { type = list(map(string)) }

module "amazon_relays" {
  relays              = var.amazon_relays
  region              = "us-east-1"
  source              = "./amazon"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "amazon_relays" {
  description = "Data for each amazon relay"
  value = module.amazon_relays.relays
}

# ----------------------------------------------------------------------------------------

variable "akamai_relays" { type = list(map(string)) }

module "akamai_relays" {
  relays              = var.akamai_relays
  source              = "./akamai"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "akamai_relays" {
  description = "Data for each akamai relay"
  value = module.akamai_relays.relays
}

# ----------------------------------------------------------------------------------------

variable "vultr_relays" { type = list(map(string)) }

module "vultr_relays" {
  relays              = var.vultr_relays
  source              = "./vultr"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "vultr_relays" {
  description = "Data for each vultr relay"
  value = module.vultr_relays.relays
}

# ----------------------------------------------------------------------------------------

// todo: not ready yet

variable "latitude_relays" { type = list(map(string)) }

/*
module "latitude_relays" {
  relays              = var.latitude_relays
  source              = "./latitude"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "latitude_relays" {
  description = "Data for each latitude relay"
  value = module.latitude_relays.relays
}
*/

# ----------------------------------------------------------------------------------------

variable "equinix_project_id" { type = string }
variable "equinix_relays" { type = list(map(string)) }

module "equinix_relays" {
  relays              = var.equinix_relays
  project             = var.equinix_project_id
  source              = "./equinix"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "equinix_relays" {
  description = "Data for each equinix relay"
  value = module.equinix_relays.relays
}

# ----------------------------------------------------------------------------------------

module "hivelocity_relays" {
  source              = "./hivelocity"
}

output "hivelocity_relays" {
  description = "Data for each hivelocity relay"
  value = module.hivelocity_relays.relays
}

# ----------------------------------------------------------------------------------------
