# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "~> 6.0.0"
    }
    google-beta = {
      source = "hashicorp/google-beta"
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
variable "port" { type = string }
variable "tags" { type = list }
variable "min_size" { type = number }
variable "max_size" { type = number }
variable "target_cpu" { type = number }
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
    startup-script = replace(var.startup_script, "##########", google_compute_address.service.address)
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
  provider = google-beta
  name     = var.service_name
  region   = var.region
  distribution_policy_zones = var.zones
  named_port {
    name = "udp"
    port = 45000
  }
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
  description = "The IP address of the external udp load balancer"
  value = google_compute_address.service.address
}

output "http_address" {
  description = "The IP address of the internal http load balancer"
  value = google_compute_address.dummy.address
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

# to get connection drain, we need a dummy internal http load balancer... =p

resource "google_compute_address" "dummy" {
  name         = "${var.service_name}-dummy"
  region       = var.region
  subnetwork   = var.default_subnetwork
  address_type = "INTERNAL"
  purpose      = "SHARED_LOADBALANCER_VIP"
}

resource "google_compute_forwarding_rule" "dummy" {
  name                  = "${var.service_name}-dummy"
  project               = var.project
  region                = var.region
  depends_on            = [var.load_balancer_subnetwork]
  ip_protocol           = "TCP"
  ip_address            = google_compute_address.dummy.id
  load_balancing_scheme = "INTERNAL_MANAGED"
  port_range            = "80"
  target                = google_compute_region_target_http_proxy.dummy.id
  network               = var.default_network
  subnetwork            = var.default_subnetwork
  network_tier          = "PREMIUM"
}

resource "google_compute_region_target_http_proxy" "dummy" {
  name     = "${var.service_name}-dummy"
  project  = var.project
  region   = var.region
  url_map  = google_compute_region_url_map.dummy.id
}

resource "google_compute_region_url_map" "dummy" {
  name            = "${var.service_name}-dummy"
  project         = var.project
  region          = var.region
  default_service = google_compute_region_backend_service.dummy.id
}

resource "google_compute_region_backend_service" "dummy" {
  name                  = "${var.service_name}-dummy"
  project               = var.project
  region                = var.region
  protocol              = "HTTP"
  port_name             = "http"
  load_balancing_scheme = "INTERNAL_MANAGED"
  timeout_sec           = 10
  health_checks         = [google_compute_region_health_check.dummy.id]
  backend {
    group                 = google_compute_region_instance_group_manager.service.instance_group
    balancing_mode        = "RATE"
    max_rate_per_instance = 1000
    capacity_scaler       = 1.0
  }
  connection_draining_timeout_sec = 300
}

resource "google_compute_region_health_check" "dummy" {
  name                = "${var.service_name}-dummy"
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

# ----------------------------------------------------------------------------------------
