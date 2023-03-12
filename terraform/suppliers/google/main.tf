# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

provider "google" {
  credentials = file(var.credentials)
  project     = var.project
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }
variable "project" { type = string }
variable "credentials" { type = string }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# ----------------------------------------------------------------------------------------

data "google_compute_network" "default" {
  name = "default"
}

resource "google_compute_firewall" "google_allow_ssh" {
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

resource "google_compute_firewall" "google_allow_udp" {
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

resource "google_compute_address" "public" {
  count        = length(var.relays)
  name         = "${replace(var.relays[count.index].name, ".", "-")}-public"
  region       = var.relays[count.index].region
  address_type = "EXTERNAL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_address" "internal" {
  count        = length(var.relays)
  name         = "${replace(var.relays[count.index].name, ".", "-")}-internal"
  region       = var.relays[count.index].region
  address_type = "INTERNAL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_instance" "relay" {
  count        = length(var.relays)
  name         = "${replace(var.relays[count.index].name, ".", "-")}"
  zone         = var.relays[count.index].zone
  machine_type = var.relays[count.index].type
  network_interface {
    network_ip = google_compute_address.internal[count.index].address
    network    = "default"
    subnetwork = "default"
    access_config {
      nat_ip = google_compute_address.public[count.index].address
    }
  }
  boot_disk {
    initialize_params {
      image = var.relays[count.index].image
    }
  }
  metadata = {
    ssh-keys = "ubuntu:${file(var.ssh_public_key_file)}"
  }
  lifecycle {
    create_before_destroy = true
  }
  metadata_startup_script = file("./setup_relay.sh")
}

output "relays" {
  description = "Data for each google relay setup by Terraform"
  value = [for i, v in var.relays : zipmap(["relay_name", "zone", "region", "public_address", "internal_address", "machine_type", "image"], [var.relays[i].name, var.relays[i].zone, var.relays[i].region, google_compute_address.public[i].address, google_compute_address.internal[i].address, var.relays[i].type, var.relays[i].image])]
}

# ----------------------------------------------------------------------------------------
