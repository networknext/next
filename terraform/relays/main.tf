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

locals {
  context = {
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
}

module "google_iowa_1" {
  relay_name        = "google.iowa.1"
  zone              = "us-central1-a"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

module "google_iowa_2" {
  relay_name        = "google.iowa.2"
  zone              = "us-central1-b"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

module "google_iowa_3" {
  relay_name        = "google.iowa.3"
  zone              = "us-central1-c"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

module "google_iowa_4" {
  relay_name        = "google.iowa.4"
  zone              = "us-central1-f"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

# ----------------------------------------------------------------------------------------

module "google_losangeles_1" {
  relay_name        = "google.losangeles.1"
  zone              = "us-west2-a"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

module "google_losangeles_2" {
  relay_name        = "google.losangeles.2"
  zone              = "us-west2-b"
  machine_type      = "n1-standard-2"
  source            = "./google_relay"
  context           = local.context
}

# ----------------------------------------------------------------------------------------
