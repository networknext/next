# ----------------------------------------------------------------------------------------

variable "credentials" { type = string }
variable "project" { type = string }
variable "location" { type = string }
variable "region" { type = string }
variable "zone" { type = string }
variable "service_account" { type = string }
variable "dev_artifacts_bucket" { type = string }
variable "machine_type" { type = string }
variable "git_hash" { type = string }

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

provider "google" {
  credentials = file(var.credentials)
  project     = var.project
  region      = var.region
  zone        = var.zone
}

# ----------------------------------------------------------------------------------------

resource "google_compute_network" "development" {
  name                    = "development"
  project                 = var.project
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "development" {
  name                     = "development"
  project                  = var.project
  ip_cidr_range            = "10.0.0.0/24"
  region                   = var.region
  network                  = google_compute_network.development.id
  private_ip_google_access = true
}

resource "google_compute_firewall" "development" {
  name          = "development"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "35.235.240.0/20"]
  allow {
    protocol = "tcp"
    ports    = ["22", "80"]
  }
}

# ----------------------------------------------------------------------------------------

resource "google_compute_instance" "test" {
  name         = "test"
  project      = var.project
  zone         = var.zone
  machine_type = var.machine_type
  network_interface {
    network    = google_compute_network.development.id
    subnetwork = google_compute_subnetwork.development.id
  }
  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }
}

# ----------------------------------------------------------------------------------------

module "magic_backend" {

  source  = "./internal_http_service"

  service_name = "magic-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.dev_artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.dev_artifacts_bucket} -a magic_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    EOF
    sudo systemctl start app.service
  EOF1

  machine_type = var.machine_type
  git_hash = var.git_hash
  project = var.project
  region = var.region
  network = google_compute_network.development.id
  subnetwork = google_compute_subnetwork.development.id
  service_network_mask = "10.0.1.0/24"
  service_account = var.service_account
}

# ----------------------------------------------------------------------------------------

























/*
# ----------------------------------------------------------------------------------------

resource "google_compute_network" "development" {
  name                    = "development"
  project                 = var.project
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "development" {
  name                     = "development"
  project                  = var.project
  ip_cidr_range            = "10.0.0.0/24"
  region                   = var.region
  network                  = google_compute_network.development.id
  private_ip_google_access = true
}

resource "google_compute_firewall" "development" {
  name          = "development"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "35.235.240.0/20"]
  allow {
    protocol = "tcp"
    ports    = ["22", "80"]
  }
}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "magic-backend" {
  name         = "magic-backend"
  subnetwork   = google_compute_subnetwork.development.id
  address_type = "INTERNAL"
  region       = "us-central1"
  purpose      = "SHARED_LOADBALANCER_VIP"
}

resource "google_compute_subnetwork" "magic-backend" {
  name          = "magic-backend"
  project       = var.project
  region        = var.region
  purpose       = "INTERNAL_HTTPS_LOAD_BALANCER"
  role          = "ACTIVE"
  network       = google_compute_network.development.id
  ip_cidr_range = "10.0.1.0/24"
}

resource "google_compute_forwarding_rule" "magic-backend" {
  name                  = "magic-backend"
  project               = var.project
  region                = var.region
  depends_on            = [google_compute_subnetwork.magic-backend]
  ip_protocol           = "TCP"
  ip_address            = google_compute_address.magic-backend.id
  load_balancing_scheme = "INTERNAL_MANAGED"
  port_range            = "80"
  target                = google_compute_region_target_http_proxy.magic-backend.id
  network               = google_compute_network.development.id
  subnetwork            = google_compute_subnetwork.development.id
  network_tier          = "PREMIUM"
}

resource "google_compute_region_target_http_proxy" "magic-backend" {
  name     = "magic-backend"
  project  = var.project
  region   = var.region
  url_map  = google_compute_region_url_map.magic-backend.id
}

resource "google_compute_region_url_map" "magic-backend" {
  name            = "magic-backend"
  project         = var.project
  region          = var.region
  default_service = google_compute_region_backend_service.magic-backend.id
}

resource "google_compute_region_backend_service" "magic-backend" {
  name                  = "magic-backend"
  project               = var.project
  region                = var.region
  protocol              = "HTTP"
  load_balancing_scheme = "INTERNAL_MANAGED"
  timeout_sec           = 10
  health_checks         = [google_compute_region_health_check.magic-backend-lb.id]
  backend {
    group           = google_compute_region_instance_group_manager.magic-backend.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
}

resource "google_compute_instance_template" "magic-backend" {
  name         = "magic-backend-${var.git_hash}"
  project      = var.project
  machine_type = var.machine_type
  tags         = ["http-server"]

  network_interface {
    network    = google_compute_network.development.id
    subnetwork = google_compute_subnetwork.development.id
  }

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    auto_delete  = true
    boot         = true
  }

  metadata = {
    startup-script = <<-EOF1
      #!/bin/bash
      gsutil cp ${var.dev_artifacts_bucket}/bootstrap.sh bootstrap.sh
      chmod +x bootstrap.sh
      sudo ./bootstrap.sh -b ${var.dev_artifacts_bucket} -a magic_backend.tar.gz
      cat <<EOF > /app/app.env
      ENV=dev
      DEBUG_LOGS=1
      EOF
      sudo systemctl start app.service
    EOF1
  }

  service_account {
    email  = var.service_account
    scopes = ["cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_region_health_check" "magic-backend-lb" {
  name                = "magic-backend"
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

resource "google_compute_health_check" "magic-backend-vm" {
  name                = "magic-backend-vm"
  check_interval_sec  = 5
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 10
  http_health_check {
    request_path = "/vm_health"
    port         = "80"
  }
}

resource "google_compute_region_instance_group_manager" "magic-backend" {
  name     = "magic-backend"
  project  = var.project
  region   = var.region
  version {
    instance_template = google_compute_instance_template.magic-backend.id
    name              = "primary"
  }
  base_instance_name = "magic-backend"
  target_size        = 2
  named_port {
    name = "http"
    port = 80
  }
  auto_healing_policies {
    health_check      = google_compute_health_check.magic-backend-vm.id
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

resource "google_compute_firewall" "magic-backend" {
  name          = "magic-backend"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["10.0.1.0/24"]
  target_tags   = ["http-server"]
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
}

# ----------------------------------------------------------------------------------------

resource "google_redis_instance" "redis" {
  name           = "redis"
  tier           = "BASIC"
  memory_size_gb = 1
  region         = "us-central1"
  redis_version  = "REDIS_6_X"
  authorized_network = google_compute_network.development.id
}

output "redis_host" {
  description = "The IP address of the redis instance"
  value = "${google_redis_instance.redis.host}"
}

# ----------------------------------------------------------------------------------------

resource "google_compute_global_address" "relay-gateway" {
  name = "relay-gateway"
}

resource "google_compute_subnetwork" "relay-gateway" {
  name          = "relay-gateway"
  ip_cidr_range = "10.0.2.0/24"
  region        = "us-central1"
  network       = google_compute_network.development.id
}

resource "google_compute_global_forwarding_rule" "relay-gateway" {
  name                  = "relay-gateway"
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL"
  port_range            = "80"
  target                = google_compute_target_http_proxy.relay-gateway.id
  ip_address            = google_compute_global_address.relay-gateway.id
}

resource "google_compute_target_http_proxy" "relay-gateway" {
  name     = "relay-gateway"
  url_map  = google_compute_url_map.relay-gateway.id
}

resource "google_compute_url_map" "relay-gateway" {
  name            = "relay-gateway"
  default_service = google_compute_backend_service.relay-gateway.id
}

resource "google_compute_backend_service" "relay-gateway" {
  name                    = "relay-gateway"
  protocol                = "HTTP"
  port_name               = "http"
  load_balancing_scheme   = "EXTERNAL"
  timeout_sec             = 10
  health_checks           = [google_compute_health_check.relay-gateway-lb.id]
  backend {
    group           = google_compute_instance_group_manager.relay-gateway.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
}

resource "google_compute_instance_template" "relay-gateway" {
  name         = "relay-gateway-${var.git_hash}"
  machine_type = var.machine_type
  tags         = ["allow-health-check", "http-server"]

  network_interface {
    network    = google_compute_network.development.id
    subnetwork = google_compute_subnetwork.development.id
  }

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    auto_delete  = true
    boot         = true
  }

  metadata = {
    startup-script = <<-EOF1
      #!/bin/bash
      gsutil cp ${var.dev_artifacts_bucket}/bootstrap.sh bootstrap.sh
      chmod +x bootstrap.sh
      sudo ./bootstrap.sh -b ${var.dev_artifacts_bucket} -a relay_gateway.tar.gz
      cat <<EOF > /app/app.env
      ENV=dev
      DEBUG_LOGS=1
      GOOGLE_PROJECT_ID=${var.project}
      REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
      MAGIC_URL="http://${google_compute_address.magic-backend.address}/magic"
      DATABASE_URL="${var.dev_artifacts_bucket}/database.bin"
      DATABASE_PATH="/app/database.bin"
      RELAY_BACKEND_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=
      RELAY_BACKEND_PRIVATE_KEY=ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=
      EOF
      sudo gsutil cp ${var.dev_artifacts_bucket}/database.bin /app/database.bin
      sudo systemctl start app.service
    EOF1
  }

  service_account {
    email  = var.service_account
    scopes = ["cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_health_check" "relay-gateway-lb" {
  name     = "relay-gateway-lb"
  check_interval_sec  = 5
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 10
  http_health_check {
    request_path = "/lb_health"
    port         = "80"
  }
}

resource "google_compute_health_check" "relay-gateway-vm" {
  name                = "relay-gateway-vm"
  check_interval_sec  = 5
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 10
  http_health_check {
    request_path = "/vm_health"
    port         = "80"
  }
}

resource "google_compute_instance_group_manager" "relay-gateway" {
  name     = "relay-gateway"
  zone     = "us-central1-a"
  named_port {
    name = "http"
    port = 80
  }
  version {
    instance_template = google_compute_instance_template.relay-gateway.id
    name              = "primary"
  }
  base_instance_name = "relay-gateway"
  target_size        = 2
  auto_healing_policies {
    health_check      = google_compute_health_check.relay-gateway-vm.id
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

resource "google_compute_firewall" "relay-gateway" {
  name          = "relay-gateway"
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16"]
  allow {
    protocol = "tcp"
  }
  target_tags = ["allow-health-check"]
}

output "relay-gateway-address" {
  description = "The IP address of the relay frontend load balancer"
  value = google_compute_global_address.relay-gateway.address
}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "relay-backend" {
  name         = "relay-backend"
  subnetwork   = google_compute_subnetwork.development.id
  address_type = "INTERNAL"
  region       = "us-central1"
  purpose      = "SHARED_LOADBALANCER_VIP"
}

resource "google_compute_subnetwork" "relay-backend" {
  name          = "relay-backend"
  project       = var.project
  region        = var.region
  purpose       = "INTERNAL_HTTPS_LOAD_BALANCER"
  role          = "ACTIVE"
  network       = google_compute_network.development.id
  ip_cidr_range = "10.0.3.0/24"
}

resource "google_compute_forwarding_rule" "relay-backend" {
  name                  = "relay-backend"
  project               = var.project
  region                = var.region
  depends_on            = [google_compute_subnetwork.relay-backend]
  ip_protocol           = "TCP"
  ip_address            = google_compute_address.relay-backend.id
  load_balancing_scheme = "INTERNAL_MANAGED"
  port_range            = "80"
  target                = google_compute_region_target_http_proxy.relay-backend.id
  network               = google_compute_network.development.id
  subnetwork            = google_compute_subnetwork.development.id
  network_tier          = "PREMIUM"
}

resource "google_compute_region_target_http_proxy" "relay-backend" {
  name     = "relay-backend"
  project  = var.project
  region   = var.region
  url_map  = google_compute_region_url_map.relay-backend.id
}

resource "google_compute_region_url_map" "relay-backend" {
  name            = "relay-backend"
  project         = var.project
  region          = var.region
  default_service = google_compute_region_backend_service.relay-backend.id
}

resource "google_compute_region_backend_service" "relay-backend" {
  name                  = "relay-backend"
  project               = var.project
  region                = var.region
  protocol              = "HTTP"
  load_balancing_scheme = "INTERNAL_MANAGED"
  timeout_sec           = 10
  health_checks         = [google_compute_region_health_check.relay-backend-lb.id]
  backend {
    group           = google_compute_region_instance_group_manager.relay-backend.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
}

resource "google_compute_instance_template" "relay-backend" {
  name         = "relay-backend-${var.git_hash}"
  project      = var.project
  machine_type = var.machine_type
  tags         = ["http-server"]

  network_interface {
    network    = google_compute_network.development.id
    subnetwork = google_compute_subnetwork.development.id
  }

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    auto_delete  = true
    boot         = true
  }

  metadata = {
    startup-script = <<-EOF1
      #!/bin/bash
      gsutil cp ${var.dev_artifacts_bucket}/bootstrap.sh bootstrap.sh
      chmod +x bootstrap.sh
      sudo ./bootstrap.sh -b ${var.dev_artifacts_bucket} -a relay_backend.tar.gz
      cat <<EOF > /app/app.env
      ENV=dev
      DEBUG_LOGS=1
      GOOGLE_PROJECT_ID=${var.project}
      REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
      MAGIC_URL="http://${google_compute_address.magic-backend.address}/magic"
      DATABASE_URL="${var.dev_artifacts_bucket}/database.bin"
      DATABASE_PATH="/app/database.bin"
      READY_DELAY=5s
      EOF
      sudo gsutil cp ${var.dev_artifacts_bucket}/database.bin /app/database.bin
      sudo systemctl start app.service
    EOF1
  }

  service_account {
    email  = var.service_account
    scopes = ["cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_region_health_check" "relay-backend-lb" {
  name                = "relay-backend"
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

resource "google_compute_health_check" "relay-backend-vm" {
  name                = "relay-backend-vm"
  check_interval_sec  = 5
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 10
  http_health_check {
    request_path = "/vm_health"
    port         = "80"
  }
}

resource "google_compute_region_instance_group_manager" "relay-backend" {
  name     = "relay-backend"
  project  = var.project
  region   = var.region
  version {
    instance_template = google_compute_instance_template.relay-backend.id
    name              = "primary"
  }
  base_instance_name = "relay-backend"
  target_size        = 2
  named_port {
    name = "http"
    port = 80
  }
  auto_healing_policies {
    health_check      = google_compute_health_check.relay-backend-vm.id
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

resource "google_compute_firewall" "relay-backend" {
  name          = "relay-backend"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["10.0.2.0/24"]
  target_tags   = ["http-server"]
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
}
*/

# ----------------------------------------------------------------------------------------

/*
resource "google_sql_database_instance" "postgres" {
  name = "postgres"
  database_version = "POSTGRES_14"
  region = "${var.region}"
  settings {
    tier = "db-f1-micro"
  }
  deletion_protection = false
}

resource "google_sql_database" "database" {
  name      = "database"
  instance  = "${google_sql_database_instance.postgres.name}"
}

resource "google_sql_user" "users" {
  name     = "developer"
  password = "developer"
  instance = "${google_sql_database_instance.postgres.name}"
}

output "postgres_host" {
  description = "The IP address of the postgres instance"
  value = "${google_sql_database_instance.postgres.ip_address.0.ip_address}"
}
*/

# ----------------------------------------------------------------------------------------
