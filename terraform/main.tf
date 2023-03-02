terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

provider "google" {
  credentials = file("~/Documents/terraform.json")
  project = "heroic-grove-379322"
  region  = "us-central1"
  zone    = "us-central1-c"
}

resource "google_compute_network" "default" {
  name                    = "network-next"
  auto_create_subnetworks = false
  mtu                     = 1460
}

resource "google_compute_subnetwork" "default" {
  name          = "network-next"
  ip_cidr_range = "10.0.1.0/24"
  region        = "us-central1"
  network       = google_compute_network.default.id
}

resource "google_compute_instance_template" "magic-backend" {

  name        = "magic-backend"

  description = "Serves up magic values for other services in the Network Next Backend"

  tags = ["http-server", "https-server"]

  labels = {
    environment = "dev"
  }

  machine_type         = "f1-micro"

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
  }

  disk {
    source_image      = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    auto_delete       = true
    boot              = true
  }

  network_interface {
    network = "default"
  }
}

resource "google_compute_health_check" "traffic" {
  name                = "traffic"
  check_interval_sec  = 1
  timeout_sec         = 1
  healthy_threshold   = 10
  unhealthy_threshold = 1
  http_health_check {
    request_path = "/health"
    port = 80
  }
}

resource "google_compute_http_health_check" "vm" {
  name         = "vm"
  request_path = "/health_check"
  timeout_sec        = 1
  check_interval_sec = 1
}

resource "google_compute_target_pool" "magic-backend" {
  name = "magic-backend"
  health_checks = [
    google_compute_http_health_check.vm.name,
  ]
}

resource "google_compute_region_instance_group_manager" "magic-backend" {

  name = "magic-backend"

  base_instance_name         = "magic-backend"
  region                     = "us-central1"
  distribution_policy_zones  = ["us-central1-a", "us-central1-f"]

  version {
    instance_template = google_compute_instance_template.magic-backend.id
  }

  target_pools = [google_compute_target_pool.magic-backend.id]
  target_size  = 2

  named_port {
    name = "http"
    port = 80
  }

  auto_healing_policies {
    health_check      = google_compute_health_check.traffic.id
    initial_delay_sec = 300
  }
}

resource "google_storage_bucket" "dev-artifacts" {
  name = "network_next_dev_artifacts"
  storage_class = "MULTI_REGIONAL"
  location = "US"
}

resource "google_storage_bucket" "relay-artifacts" {
  name = "network_next_relay_artifacts"
  storage_class = "MULTI_REGIONAL"
  location = "US"
}
