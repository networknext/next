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

variable "google_credentials" { type = string }
variable "google_project" { type = string }
variable "google_relays" { type = list(map(string)) }

module "google_relays" {
  relays              = var.google_relays
  project             = var.google_project
  credentials         = var.google_credentials
  source              = "../../suppliers/google"
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
  source              = "../../suppliers/amazon"
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
  source              = "../../suppliers/akamai"
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
  source              = "../../suppliers/vultr"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "vultr_relays" {
  description = "Data for each vultr relay"
  value = module.vultr_relays.relays
}

# ----------------------------------------------------------------------------------------

variable "latitude_relays" { type = list(map(string)) }

module "latitude_relays" {
  project_name        = "Test Relays"
  project_description = "Test Relays"
  project_environment = "Development"
  relays              = var.latitude_relays
  source              = "../../suppliers/latitude"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "latitude_relays" {
  description = "Data for each latitude relay"
  value = module.latitude_relays.relays
}

# ----------------------------------------------------------------------------------------

variable "equinix_project_id" { type = string }
variable "equinix_relays" { type = list(map(string)) }

module "equinix_relays" {
  relays              = var.equinix_relays
  project             = var.equinix_project_id
  source              = "../../suppliers/equinix"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "equinix_relays" {
  description = "Data for each equinix relay"
  value = module.equinix_relays.relays
}

# ----------------------------------------------------------------------------------------

variable "hivelocity_relays" { type = list(map(string)) }

module "hivelocity_relays" {
  relays              = var.hivelocity_relays
  source              = "../../suppliers/hivelocity"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "hivelocity_relays" {
  description = "Data for each hivelocity relay"
  value = module.hivelocity_relays.relays
}

# ----------------------------------------------------------------------------------------

variable "gcore_project_id" { type = string }
variable "gcore_relays" { type = list(map(string)) }

module "gcore_relays" {
  relays              = var.gcore_relays
  project             = var.gcore_project_id
  source              = "../../suppliers/gcore"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

/*
output "gcore_relays" {
  description = "Data for each gcore relay"
  value = module.gcore_relays.relays
}
*/

# ----------------------------------------------------------------------------------------

variable "phoenixnap_client_id" { type = string }
variable "phoenixnap_relays" { type = list(map(string)) }

/*
module "phoenixnap_relays" {
  relays              = var.phoenixnap_relays
  client_id           = var.phoenixnap_client_id
  source              = "../../suppliers/phoenixnap"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

output "phoenixnap_relays" {
  description = "Data for each phoenixnap relay"
  value = module.phoenixnap_relays.relays
}
*/

# ----------------------------------------------------------------------------------------

variable "bare_metal_relays" { type = list(map(string)) }

module "bare_metal_relays" {
  relays = var.bare_metal_relays
  source = "../../suppliers/bare_metal"
}

output "bare_metal_relays" {
  description = "Data for each bare metal relay"
  value = module.bare_metal_relays.relays
}

# ----------------------------------------------------------------------------------------
