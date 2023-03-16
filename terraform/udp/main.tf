# ----------------------------------------------------------------------------------------

variable "credentials" { type = string }
variable "project" { type = string }
variable "location" { type = string }
variable "region" { type = string }
variable "zone" { type = string }
variable "service_account" { type = string }
variable "artifacts_bucket" { type = string }
variable "machine_type" { type = string }
variable "git_hash" { type = string }
variable "vpn_address" { type = string }

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
  region      = var.region
  zone        = var.zone
}

# ----------------------------------------------------------------------------------------

resource "google_compute_network" "development" {
  name                    = "development"
  project                 = var.project
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "development" {
  name                     = "development"
  project                  = var.project
  ip_cidr_range            = "10.0.0.0/16"
  region                   = var.region
  network                  = google_compute_network.development.id
  private_ip_google_access = true
}

# ----------------------------------------------------------------------------------------

resource "google_compute_firewall" "allow_ssh" {
  name          = "allow-ssh"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "35.235.240.0/20"]
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
  target_tags = ["allow-ssh"]
}

resource "google_compute_firewall" "allow_health_checks" {
  name          = "allow-health-checks"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["0.0.0.0/0"]

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  target_tags = ["allow-health-checks"]
}

resource "google_compute_firewall" "allow_udp" {
  name          = "allow-udp"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
    ports    = ["40000"]
  }
  target_tags = ["allow-udp"]
}

# ----------------------------------------------------------------------------------------

resource "google_compute_instance" "test" {
  name         = "test"
  project      = var.project
  zone         = var.zone
  machine_type = var.machine_type
  network_interface {
    network    = google_compute_network.development.id
    subnetwork = google_compute_subnetwork.development.id
    access_config {}
  }
  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }
  tags = ["allow-ssh"]
}

# ----------------------------------------------------------------------------------------

module "udp" {

  source = "./external_udp_service"

  service_name = "udp"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a udp.tar.gz
    sudo touch /app/app.env
    sudo systemctl start app.service
  EOF1

  machine_type       = var.machine_type
  git_hash           = var.git_hash
  project            = var.project
  region             = var.region
  port               = 40000
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-udp"]
}

output "udp_address" {
  description = "The IP address of the udp load balancer"
  value       = module.udp.address
}

# ----------------------------------------------------------------------------------------
