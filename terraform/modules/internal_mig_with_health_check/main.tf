# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "~> 6.0.0"
    }
  }
}

# ----------------------------------------------------------------------------------------

variable "service_name" { type = string }
variable "startup_script" { type = string }
variable "machine_type" { type = string }
variable "tag" { type = string }
variable "extra" { type = string }
variable "project" { type = string }
variable "region" { type = string }
variable "zones" { type = list(string) }
variable "default_network" { type = string }
variable "default_subnetwork" { type = string }
variable "service_account" { type = string }
variable "tags" { type = list }
variable "target_size" { 
  type = number
  default = 2
}
variable "initial_delay" {
  type = number
  default = 60
}

# ----------------------------------------------------------------------------------------

resource "google_compute_instance_template" "service" {
  name         = "${var.service_name}-${var.tag}${var.extra}"
  machine_type = var.machine_type

  network_interface {
    network    = var.default_network
    subnetwork = var.default_subnetwork
  }

  tags = var.tags

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    auto_delete  = true
    boot         = true
    disk_type    = "pd-ssd"
  }

  metadata = {
    startup-script = var.startup_script
  }

  service_account {
    email  = var.service_account
    scopes = ["cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_health_check" "service_vm" {
  name                = "${var.service_name}-vm"
  check_interval_sec  = 5
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 10
  http_health_check {
    request_path = "/vm_health"
    port         = "80"
  }
}

resource "google_compute_region_instance_group_manager" "service" {
  name     = var.service_name
  region   = var.region
  distribution_policy_zones = var.zones
  version {
    instance_template = google_compute_instance_template.service.id
    name              = "primary"
  }
  base_instance_name = var.service_name
  target_size        = var.target_size
  auto_healing_policies {
    health_check      = google_compute_health_check.service_vm.id
    initial_delay_sec = var.initial_delay
  }
  update_policy {
    type                           = "PROACTIVE"
    minimal_action                 = "REPLACE"
    most_disruptive_allowed_action = "REPLACE"
    max_surge_fixed                = 10
    max_unavailable_fixed          = 0
    replacement_method             = "SUBSTITUTE"
  }
}

# ----------------------------------------------------------------------------------------
