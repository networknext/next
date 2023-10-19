# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "~> 5.0.0"
    }
    google-beta = {
      source = "hashicorp/google-beta"
      version = "~> 5.0.0"
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
variable "port" { type = string }
variable "tags" { type = list }
variable "min_size" { type = number }
variable "max_size" { type = number }
variable "target_cpu" { type = number }

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "service" {
  name         = var.service_name
  network_tier = "STANDARD"
}

resource "google_compute_forwarding_rule" "service" {
  name                  = var.service_name
  region                = var.region
  project               = var.project
  ip_address            = google_compute_address.service.id
  port_range            = var.port
  network_tier          = "STANDARD"
  load_balancing_scheme = "EXTERNAL"
  ip_protocol           = "UDP"
  backend_service       = google_compute_region_backend_service.service.id
}

resource "google_compute_region_backend_service" "service" {
  name                  = var.service_name
  project               = var.project
  region                = var.region
  protocol              = "UDP"
  port_name             = "udp"
  load_balancing_scheme = "EXTERNAL"
  health_checks         = [google_compute_region_health_check.service_lb.id]
  backend {
    group           = google_compute_region_instance_group_manager.service.instance_group
    balancing_mode  = "CONNECTION"
  }
}

resource "google_compute_instance_template" "service" {

  provider     = google-beta

  name         = "${var.service_name}-${var.tag}${var.extra}"
  machine_type = var.machine_type

  network_interface {
    network    = var.default_network
    subnetwork = var.default_subnetwork
    access_config {}
  }

  network_performance_config {
    total_egress_bandwidth_tier = "TIER_1"
  }

  tags = var.tags

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    auto_delete  = true
    boot         = true
    disk_type    = "pd-ssd"
  }

  metadata = {
    startup-script = replace(var.startup_script, "##########", google_compute_address.service.address)
    shutdown-script = <<-EOF
      #!/bin/bash
      sleep 90
    EOF
  }

  service_account {
    email  = var.service_account
    scopes = ["cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_region_health_check" "service_lb" {
  name                = "${var.service_name}-lb"
  timeout_sec         = 1
  check_interval_sec  = 1
  healthy_threshold   = 10
  unhealthy_threshold = 1
  project             = var.project
  region              = var.region
  http_health_check {
    request_path = "/lb_health"
    port = "80"
  }
}

resource "google_compute_health_check" "service_vm" {
  name                = "${var.service_name}-vm"
  check_interval_sec  = 10
  timeout_sec         = 5
  healthy_threshold   = 3
  unhealthy_threshold = 6
  http_health_check {
    request_path = "/vm_health"
    port         = "80"
  }
}

resource "google_compute_region_instance_group_manager" "service" {
  name     = var.service_name
  region   = var.region
  distribution_policy_zones = var.zones
  named_port {
    name = "udp"
    port = 45000
  }
  version {
    instance_template = google_compute_instance_template.service.id
    name              = "primary"
  }
  base_instance_name = var.service_name
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
  value = google_compute_address.service.address
}

# ----------------------------------------------------------------------------------------

resource "google_compute_region_autoscaler" "default" {
  name   = var.service_name
  region = var.region
  target = google_compute_region_instance_group_manager.service.id
  autoscaling_policy {
    max_replicas    = var.max_size
    min_replicas    = var.min_size
    cooldown_period = 60
    cpu_utilization {
      target = var.target_cpu / 100.0
    }    
  }
}

# ----------------------------------------------------------------------------------------