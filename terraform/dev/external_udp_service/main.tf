# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

# ----------------------------------------------------------------------------------------

variable "service_name" { type = string }
variable "startup_script" { type = string }
variable "machine_type" { type = string }
variable "git_hash" { type = string }
variable "project" { type = string }
variable "zone" { type = string }
variable "default_network" { type = string }
variable "default_subnetwork" { type = string }
variable "service_account" { type = string }
variable "port" { type = string }

# ----------------------------------------------------------------------------------------

resource "google_compute_forwarding_rule" "service" {
  name                  = var.service_name
  project               = var.project
  target                = google_compute_target_pool.service.self_link
  load_balancing_scheme = "EXTERNAL"
  ip_protocol           = "UDP"
  port_range            = var.port
}

resource "google_compute_target_pool" "default" {
  name             = var.service_name
  project          = var.project
  region           = var.region

  instances = var.instances

  health_checks = [
    google_compute_http_health_check.service_lb.name,
  ]
}

resource "google_compute_instance_template" "service" {
  name         = "${var.service_name}-${var.git_hash}"
  machine_type = var.machine_type
  tags         = ["allow-health-check", "http-server", "udp-server"]

  network_interface {
    network    = var.default_network
    subnetwork = var.default_subnetwork
  }

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    auto_delete  = true
    boot         = true
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

resource "google_compute_health_check" "service_lb" {
  name                = "${var.service_name}-lb"
  check_interval_sec  = 1
  timeout_sec         = 1
  healthy_threshold   = 5
  unhealthy_threshold = 2
  http_health_check {
    request_path = "/lb_health"
    port         = "80"
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

resource "google_compute_instance_group_manager" "service" {
  name     = var.service_name
  zone     = var.zone
  named_port {
    name = "udp"
    port = 45000
  }
  version {
    instance_template = google_compute_instance_template.service.id
    name              = "primary"
  }
  base_instance_name = var.service_name
  target_size        = 2
  auto_healing_policies {
    health_check      = google_compute_health_check.service_vm.id
    initial_delay_sec = 120
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

output "address" {
  description = "The IP address of the external udp load balancer"
  value = google_compute_global_address.service.address
}

# ----------------------------------------------------------------------------------------
