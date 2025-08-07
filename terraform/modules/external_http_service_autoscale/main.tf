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
variable "min_size" { type = number }
variable "max_size" { type = number }
variable "target_cpu" { type = number }
variable "domain" { type = string }
variable "certificate" { type = string }
variable "initial_delay" {
  type = number
  default = 60
}
variable "tier_1" {
  type = bool
  default = false
}

# ----------------------------------------------------------------------------------------

resource "google_compute_target_https_proxy" "service" {
  name             = var.service_name
  url_map          = google_compute_url_map.service.id
  ssl_certificates = [var.certificate]
}

resource "google_compute_global_address" "service" {
  name = var.service_name
}

resource "google_compute_url_map" "service" {

  name  = var.service_name

  default_service = google_compute_backend_service.service.id

  host_rule {
    hosts        = [var.domain]
    path_matcher = "allpaths"
  }

  path_matcher {
    name            = "allpaths"
    default_service = google_compute_backend_service.service.id
    path_rule {
      paths   = ["/*"]
      service = google_compute_backend_service.service.id
    }
  }
}

resource "google_compute_global_forwarding_rule" "service" {
  name                  = var.service_name
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL"
  port_range            = 443
  target                = google_compute_target_https_proxy.service.id
  ip_address            = google_compute_global_address.service.id
}

resource "google_compute_backend_service" "service" {
  name                    = var.service_name
  protocol                = "HTTP"
  port_name               = "http"
  load_balancing_scheme   = "EXTERNAL"
  timeout_sec             = 10
  health_checks           = [google_compute_health_check.service_lb.id]
  backend {
    group           = google_compute_region_instance_group_manager.service.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
  connection_draining_timeout_sec = 360
}

resource "google_compute_instance_template" "service" {

  provider     = google-beta

  name         = "${var.service_name}-${var.tag}${var.extra}"
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

resource "google_compute_region_instance_group_manager" "service" {
  provider = google-beta
  name     = var.service_name
  region   = var.region
  distribution_policy_zones = var.zones
  named_port {
    name = "http"
    port = 80
  }
  version {
    instance_template = google_compute_instance_template.service.id
    name              = "primary"
  }
  base_instance_name = var.service_name
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
    min_ready_sec                  = var.initial_delay
  }
}

output "address" {
  description = "The IP address of the external http load balancer"
  value = google_compute_global_address.service.address
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
  depends_on = [google_compute_region_instance_group_manager.service]
}

# ----------------------------------------------------------------------------------------
