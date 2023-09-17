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
variable "tag" { type = string }
variable "extra" { type = string }
variable "project" { type = string }
variable "zone" { type = string }                   # todo: want to remove this and go regional
variable "zones" { type = list(string) }
variable "default_network" { type = string }
variable "default_subnetwork" { type = string }
variable "service_account" { type = string }
variable "tags" { type = list }
variable "min_size" { type = number }
variable "max_size" { type = number }
variable "target_cpu" { type = number }

# ----------------------------------------------------------------------------------------

resource "google_compute_global_address" "service" {
  name = var.service_name
}

resource "google_compute_global_forwarding_rule" "service" {
  name                  = var.service_name
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL"
  port_range            = "80"
  target                = google_compute_target_http_proxy.service.id
  ip_address            = google_compute_global_address.service.id
}

resource "google_compute_target_http_proxy" "service" {
  name     = var.service_name
  url_map  = google_compute_url_map.service.id
}

resource "google_compute_url_map" "service" {
  name            = var.service_name
  default_service = google_compute_backend_service.service.id
}

resource "google_compute_backend_service" "service" {
  name                    = var.service_name
  protocol                = "HTTP"
  port_name               = "http"
  load_balancing_scheme   = "EXTERNAL"
  timeout_sec             = 10
  health_checks           = [google_compute_health_check.service_lb.id]
  backend {
    group           = google_compute_instance_group_manager.service.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
}

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

# todo: this instance manager is zonal, but we want it regional

resource "google_compute_instance_group_manager" "service" {
  name     = var.service_name
  zone     = var.zone
  named_port {
    name = "http"
    port = 80
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
  description = "The IP address of the external http load balancer"
  value = google_compute_global_address.service.address
}

# ----------------------------------------------------------------------------------------

resource "google_compute_autoscaler" "default" {
  name   = var.service_name
  target = google_compute_instance_group_manager.service.id
  autoscaling_policy {
    max_replicas    = var.max_replicas
    min_replicas    = var.min_replicas
    cooldown_period = 60
    cpu_utilization {
      target = var.target_cpu / 100.0
    }    
  }
}

# ----------------------------------------------------------------------------------------