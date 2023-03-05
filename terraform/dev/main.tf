# ----------------------------------------------------------------------------------------

variable "credentials" { type = string }
variable "project" { type = string }
variable "location" { type = string }
variable "region" { type = string }
variable "zone" { type = string }
variable "service_account" { type = string }
variable "artifacts_bucket" { type = string }
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
  ip_cidr_range            = "10.0.0.0/16"
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

resource "google_compute_subnetwork" "load_balancer" {
  name          = "load-balancer"
  project       = var.project
  region        = var.region
  purpose       = "INTERNAL_HTTPS_LOAD_BALANCER"
  role          = "ACTIVE"
  network       = google_compute_network.development.id
  ip_cidr_range = "10.1.0.0/16"
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

resource "google_redis_instance" "redis" {
  name               = "redis"
  tier               = "BASIC"
  memory_size_gb     = 1
  region             = "us-central1"
  redis_version      = "REDIS_6_X"
  authorized_network = google_compute_network.development.id
}

output "redis_address" {
  description = "The IP address of the redis instance"
  value       = google_redis_instance.redis.host
}

# ----------------------------------------------------------------------------------------

resource "google_compute_global_address" "postgres_private_address" {
  name          = "postgres-private-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.development.id
}

resource "google_service_networking_connection" "postgres" {
  network                 = google_compute_network.development.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.postgres_private_address.name]
}

resource "google_sql_database_instance" "postgres" {
  name = "postgres"
  database_version = "POSTGRES_14"
  region = "${var.region}"
  depends_on = [google_service_networking_connection.postgres]
  settings {
    tier = "db-f1-micro"
    ip_configuration {
      ipv4_enabled    = "false"
      private_network = google_compute_network.development.id
    }
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

resource "google_compute_network_peering_routes_config" "postgres" {
  peering              = google_service_networking_connection.postgres.peering
  network              = google_compute_network.development.name
  import_custom_routes = true
  export_custom_routes = true
}

output "postgres_address" {
  description = "The IP address of the postgres instance"
  value = "${google_compute_global_address.postgres_private_address.address}"
}

# ----------------------------------------------------------------------------------------

module "magic_backend" {

  source = "./internal_http_service"

  service_name = "magic-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a magic_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    EOF
    sudo systemctl start app.service
  EOF1

  machine_type               = var.machine_type
  git_hash                   = var.git_hash
  project                    = var.project
  region                     = var.region
  default_network            = google_compute_network.development.id
  default_subnetwork         = google_compute_subnetwork.development.id
  load_balancer_subnetwork   = google_compute_subnetwork.load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.load_balancer.ip_cidr_range
  service_account            = var.service_account
}

output "magic_backend_address" {
  description = "The IP address of the magic backend load balancer"
  value       = module.magic_backend.address
}

# ----------------------------------------------------------------------------------------

module "relay_gateway" {

  source = "./external_http_service"

  service_name = "relay-gateway"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a relay_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    GOOGLE_PROJECT_ID=${var.project}
    REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.artifacts_bucket}/database.bin"
    DATABASE_PATH="/app/database.bin"
    READY_DELAY=5s
    EOF
    sudo gsutil cp ${var.artifacts_bucket}/database.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  machine_type       = var.machine_type
  git_hash           = var.git_hash
  project            = var.project
  zone               = var.zone
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.service_account
}

output "relay_gateway_address" {
  description = "The IP address of the relay gateway load balancer"
  value       = module.relay_gateway.address
}

# ----------------------------------------------------------------------------------------

module "relay_backend" {

  source = "./internal_http_service"

  service_name = "relay-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a relay_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    GOOGLE_PROJECT_ID=${var.project}
    REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.artifacts_bucket}/database.bin"
    DATABASE_PATH="/app/database.bin"
    READY_DELAY=15s
    EOF
    sudo gsutil cp ${var.artifacts_bucket}/database.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  machine_type               = var.machine_type
  git_hash                   = var.git_hash
  project                    = var.project
  region                     = var.region
  default_network            = google_compute_network.development.id
  default_subnetwork         = google_compute_subnetwork.development.id
  load_balancer_subnetwork   = google_compute_subnetwork.load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.load_balancer.ip_cidr_range
  service_account            = var.service_account
}

output "relay_backend_address" {
  description = "The IP address of the relay backend load balancer"
  value       = module.relay_backend.address
}

# ----------------------------------------------------------------------------------------

module "analytics" {

  source = "./internal_http_service"

  service_name = "analytics"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a analytics.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    GOOGLE_PROJECT_ID=${var.project}
    DATABASE_URL="${var.artifacts_bucket}/database.bin"
    DATABASE_PATH="/app/database.bin"
    COST_MATRIX_URL="http://${module.relay_backend.address}/cost_matrix"
    ROUTE_MATRIX_URL="http://${module.relay_backend.address}/route_matrix"
    REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
    BIGQUERY_DATASET=dev
    EOF
    sudo gsutil cp ${var.artifacts_bucket}/database.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  machine_type               = var.machine_type
  git_hash                   = var.git_hash
  project                    = var.project
  region                     = var.region
  default_network            = google_compute_network.development.id
  default_subnetwork         = google_compute_subnetwork.development.id
  load_balancer_subnetwork   = google_compute_subnetwork.load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.load_balancer.ip_cidr_range
  service_account            = var.service_account
}

output "analytics_address" {
  description = "The IP address of the analytics load balancer"
  value       = module.analytics.address
}

# ----------------------------------------------------------------------------------------

module "api" {

  source = "./external_http_service"

  service_name = "api"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a api.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
    GOOGLE_PROJECT_ID=${var.project}
    DATABASE_URL="${var.artifacts_bucket}/database.bin"
    DATABASE_PATH="/app/database.bin"
    PGSQL_CONFIG="host=${google_sql_database_instance.postgres.ip_address.0.ip_address} port=5432 user=developer password=developer dbname=postgres sslmode=disable"
    EOF
    sudo gsutil cp ${var.artifacts_bucket}/database.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  machine_type       = var.machine_type
  git_hash           = var.git_hash
  project            = var.project
  zone               = var.zone
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.service_account
}

output "api_address" {
  description = "The IP address of the api load balancer"
  value       = module.api.address
}

// ---------------------------------------------------------------------------------------

module "portal_cruncher" {

  source = "./mig_service_with_health_check"

  service_name = "portal-cruncher"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a portal_cruncher.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
    GOOGLE_PROJECT_ID=${var.project}
    EOF
    sudo systemctl start app.service
  EOF1

  machine_type       = var.machine_type
  git_hash           = var.git_hash
  project            = var.project
  zone               = var.zone
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.service_account
}

// ---------------------------------------------------------------------------------------
