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
variable "load_balancer_subnetwork" { type = string }
variable "load_balancer_network_mask" { type = string }
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
variable "tier_1" {
  type = bool
  default = false
}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "service" {
  name         = var.service_name
  region       = var.region
  subnetwork   = var.default_subnetwork
  address_type = "INTERNAL"
  purpose      = "SHARED_LOADBALANCER_VIP"
}

output "address" {
  description = "The IP address of the load balancer"
  value = google_compute_address.service.address
}

resource "google_compute_forwarding_rule" "service" {
  name                  = var.service_name
  project               = var.project
  region                = var.region
  depends_on            = [var.load_balancer_subnetwork]
  ip_protocol           = "TCP"
  ip_address            = google_compute_address.service.id
  load_balancing_scheme = "INTERNAL_MANAGED"
  port_range            = "80"
  target                = google_compute_region_target_http_proxy.service.id
  network               = var.default_network
  subnetwork            = var.default_subnetwork
  network_tier          = "PREMIUM"
}

resource "google_compute_region_target_http_proxy" "service" {
  name     = var.service_name
  project  = var.project
  region   = var.region
  url_map  = google_compute_region_url_map.service.id
}

resource "google_compute_region_url_map" "service" {
  name            = var.service_name
  project         = var.project
  region          = var.region
  default_service = google_compute_region_backend_service.service.id
}

resource "google_compute_region_backend_service" "service" {
  name                  = var.service_name
  project               = var.project
  region                = var.region
  protocol              = "HTTP"
  load_balancing_scheme = "INTERNAL_MANAGED"
  timeout_sec           = 10
  health_checks         = [google_compute_region_health_check.service_lb.id]
  backend {
    group           = google_compute_region_instance_group_manager.service.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
  connection_draining_timeout_sec = 60
}

resource "google_compute_instance_template" "service" {

  provider     = google-beta

  name         = "${var.service_name}-${var.tag}${var.extra}"
  project      = var.project
  machine_type = var.machine_type

  network_interface {
    network    = var.default_network
    subnetwork = var.default_subnetwork
  }

  network_performance_config {
    total_egress_bandwidth_tier = var.tier_1 ? "TIER_1" : "DEFAULT"
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

resource "google_compute_region_health_check" "service_lb" {
  name                = "${var.service_name}-lb"
  timeout_sec         = 1
  check_interval_sec  = 1
  healthy_threshold   = 5
  unhealthy_threshold = 2
  project             = var.project
  region              = var.region
  http_health_check {
    request_path = "/lb_health"
    port = "80"
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
  project  = var.project
  region   = var.region
  distribution_policy_zones = var.zones
  version {
    instance_template = google_compute_instance_template.service.id
    name              = "primary"
  }
  base_instance_name = var.service_name
  target_size        = var.target_size
  named_port {
    name = "http"
    port = 80
  }
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
