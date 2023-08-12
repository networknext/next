# ========================================================================================
#                                     GOOGLE CLOUD
# ========================================================================================

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

variable "relays" { type = map(map(string)) }
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
  for_each     = var.relays
  name         = "${replace(each.key, ".", "-")}-public"
  region       = local.datacenter_map[each.value.datacenter_name].region
  address_type = "EXTERNAL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_address" "internal" {
  for_each     = var.relays
  name         = "${replace(each.key, ".", "-")}-internal"
  region       = local.datacenter_map[each.value.datacenter_name].region
  address_type = "INTERNAL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_instance" "relay" {
  for_each     = var.relays
  name         = "${replace(each.key, ".", "-")}"
  zone         = local.datacenter_map[each.value.datacenter_name].zone
  machine_type = each.value.type
  network_interface {
    network_ip = google_compute_address.internal[each.key].address
    network    = "default"
    subnetwork = "default"
    access_config {
      nat_ip = google_compute_address.public[each.key].address
    }
  }
  boot_disk {
    initialize_params {
      image = each.value.image
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

# ----------------------------------------------------------------------------------------

output "relays" {

  description = "Data for each google relay setup by Terraform"

  value = {
    for k, v in var.relays : k => zipmap( 
      [
        "relay_name", 
        "datacenter_name",
        "supplier_name", 
        "public_address", 
        "internal_address", 
        "internal_group", 
        "ssh_address", 
        "ssh_user",
      ], 
      [
        k,
        v.datacenter_name,
        "google", 
        "${google_compute_address.public[k].address}:40000",
        "${google_compute_address.internal[k].address}:40000",
        "", 
        "${google_compute_address.public[k].address}:22",
        "ubuntu",
      ]
    )
  }
}

output "datacenters" {
  description = "Data for each google datacenter"
  value = local.datacenter_map
}

# ----------------------------------------------------------------------------------------
