# ----------------------------------------------------------------------------------------

variable "credentials" { type = string }
variable "project" { type = string }
variable "vpn_address" { type = string }
variable "ssh_public_key_file" { type = string }
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
    google = {
      source  = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

provider "google" {
  credentials = file(var.credentials)
  project     = var.project
}

# ----------------------------------------------------------------------------------------

data "google_compute_network" "default" {
  name = "default"
}

resource "google_compute_firewall" "allow_ssh" {
  name          = "allow-ssh"
  project       = var.project
  direction     = "INGRESS"
  network       = "default"
  source_ranges = [var.vpn_address]
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}

resource "google_compute_firewall" "allow_udp" {
  name          = "allow-udp"
  project       = var.project
  direction     = "INGRESS"
  network       = "default"
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
    ports    = ["40000"]
  }
}

# ----------------------------------------------------------------------------------------

module "google_relay" {
  source                 = "./google_relay"
  relay_name               = "google.iowa.1"
  region                   = "us-central1"
  zone                     = "us-central1-a"
  machine_type             = "n1-standard-2"
  vpn_address              = var.vpn_address
  ssh_public_key_file      = var.ssh_public_key_file
  env                      = var.env
  relay_version            = var.relay_version
  relay_artifacts_bucket   = var.relay_artifacts_bucket
  relay_public_key         = var.relay_public_key
  relay_private_key        = var.relay_private_key
  relay_backend_hostname   = var.relay_backend_hostname
  relay_backend_public_key = var.relay_backend_public_key
}

output "relay_public_address" {
  description = "The public IP address of the google cloud relay"
  value       = module.google_relay.public_address
}

output "relay_internal_address" {
  description = "The internal IP address of the google cloud relay"
  value       = module.google_relay.internal_address
}

# ----------------------------------------------------------------------------------------
