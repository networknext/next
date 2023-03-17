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
variable "vpn_address" { type = string }

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

resource "google_compute_subnetwork" "internal_http_load_balancer" {
  name          = "internal-http-load-balancer"
  project       = var.project
  region        = var.region
  purpose       = "INTERNAL_HTTPS_LOAD_BALANCER"
  role          = "ACTIVE"
  network       = google_compute_network.development.id
  ip_cidr_range = "10.1.0.0/16"
}

# ----------------------------------------------------------------------------------------

resource "google_compute_firewall" "allow_ssh" {
  name          = "allow-ssh"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "35.235.240.0/20"]
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
  target_tags = ["allow-ssh"]
}

resource "google_compute_firewall" "allow_health_checks" {
  name          = "allow-health-checks"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["35.191.0.0/16", "130.211.0.0/22"]

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  target_tags = ["allow-health-checks"]
}

resource "google_compute_firewall" "allow_network_load_balancer_traffic" {
  name          = "allow-network-load-balancer-traffic"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "35.235.240.0/20", "209.85.152.0/22", "209.85.204.0/22"]

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  target_tags = ["allow-health-checks"]
}

resource "google_compute_firewall" "allow_http" {
  name          = "allow-http"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
  target_tags = ["allow-http"]
}

resource "google_compute_firewall" "allow_http_vpn_only" {
  name          = "allow-http-vpn-only"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["${var.vpn_address}/32"]
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
  target_tags = ["allow-http-vpn-only"]
}

resource "google_compute_firewall" "allow_udp_45000" {
  name          = "allow-udp-45000"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
    ports    = ["45000"]
  }
  target_tags = ["allow-udp-45000"]
}

resource "google_compute_firewall" "allow_udp_all" {
  name          = "allow-udp-all"
  project       = var.project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
  }
  target_tags = ["allow-udp-all"]
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
    access_config {}
  }
  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }
  tags = ["allow-ssh"]
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
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-http"]
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
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a relay_gateway.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    GOOGLE_PROJECT_ID=${var.project}
    REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.artifacts_bucket}/database.bin"
    DATABASE_PATH="/app/database.bin"
    RELAY_BACKEND_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=
    RELAY_BACKEND_PRIVATE_KEY=ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=
    EOF
    sudo gsutil cp ${var.artifacts_bucket}/database.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  machine_type             = var.machine_type
  git_hash                 = var.git_hash
  project                  = var.project
  zone                     = var.zone
  default_network          = google_compute_network.development.id
  default_subnetwork       = google_compute_subnetwork.development.id
  service_account          = var.service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http"]
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
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-http"]
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
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.service_account
  tags                       = ["allow-ssh", "allow-health-checks"]
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

  machine_type             = var.machine_type
  git_hash                 = var.git_hash
  project                  = var.project
  zone                     = var.zone
  default_network          = google_compute_network.development.id
  default_subnetwork       = google_compute_subnetwork.development.id
  service_account          = var.service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http-vpn-only"]
}

output "api_address" {
  description = "The IP address of the api load balancer"
  value       = module.api.address
}

// ---------------------------------------------------------------------------------------

module "portal_cruncher" {

  source = "./internal_mig_with_health_check"

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
    EOF
    sudo systemctl start app.service
  EOF1

  machine_type       = var.machine_type
  git_hash           = var.git_hash
  project            = var.project
  region             = var.region
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-http"]
}

// ---------------------------------------------------------------------------------------

module "map_cruncher" {

  source = "./internal_mig_with_health_check"

  service_name = "map-cruncher"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a map_cruncher.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
    EOF
    sudo systemctl start app.service
  EOF1

  machine_type       = var.machine_type
  git_hash           = var.git_hash
  project            = var.project
  region             = var.region
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-http"]
}

# ----------------------------------------------------------------------------------------

module "server_backend" {

  source = "./external_udp_service"

  service_name = "server-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a server_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    UDP_PORT=45000
    GOOGLE_PROJECT_ID=${var.project}
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
    RELAY_BACKEND_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=
    RELAY_BACKEND_PRIVATE_KEY=ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=
    SERVER_BACKEND_ADDRESS="##########:45000"
    SERVER_BACKEND_PUBLIC_KEY=TGHKjEeHPtSgtZfDyuDPcQgtJTyRDtRvGSKvuiWWo0A=
    SERVER_BACKEND_PRIVATE_KEY=FXwFqzjGlIwUDwiq1N5Um5VUesdr4fP2hVV2cnJ+yARMYcqMR4c+1KC1l8PK4M9xCC0lPJEO1G8ZIq+6JZajQA==
    ROUTE_MATRIX_URL="http://${module.relay_backend.address}/route_matrix"
    EOF
    sudo gsutil cp ${var.artifacts_bucket}/database.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  machine_type       = var.machine_type
  git_hash           = var.git_hash
  project            = var.project
  region             = var.region
  port               = 45000
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-udp-45000"]
}

output "server_backend_address" {
  description = "The IP address of the server backend load balancer"
  value       = module.server_backend.address
}

# ----------------------------------------------------------------------------------------

module "raspberry_backend" {

  source = "./external_http_service"

  service_name = "raspberry-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a raspberry_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_HOSTNAME="${google_redis_instance.redis.host}:6379"
    EOF
    sudo systemctl start app.service
  EOF1

  machine_type             = var.machine_type
  git_hash                 = var.git_hash
  project                  = var.project
  zone                     = var.zone
  default_network          = google_compute_network.development.id
  default_subnetwork       = google_compute_subnetwork.development.id
  service_account          = var.service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http"]
}

output "raspberry_backend_address" {
  description = "The IP address of the raspberry backend load balancer"
  value       = module.raspberry_backend.address
}

# ----------------------------------------------------------------------------------------

module "raspberry_server" {

  source = "./external_mig_without_health_check"

  service_name = "raspberry-server"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a raspberry_server.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    NEXT_LOG_LEVEL=4
    NEXT_DATACENTER=cloud
    NEXT_CUSTOMER_PRIVATE_KEY=UoFYERKJnCtieFM9lnPGJHvHDRAuOYDIbMKhx3QnkTnGrsPwsQFuB3XyZTncixbOURcPalgP3J35OJmKr35wwX1wcbiQzBG3
    RASPBERRY_BACKEND_URL="http://${module.raspberry_backend.address}"
    EOF
    sudo gsutil cp ${var.artifacts_bucket}/libnext5.so /usr/local/lib/libnext5.so
    sudo ldconfig
    sudo systemctl start app.service
  EOF1

  machine_type       = var.machine_type
  git_hash           = var.git_hash
  project            = var.project
  region             = var.region
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.service_account
  tags               = ["allow-ssh", "allow-udp-all"]
}

# ----------------------------------------------------------------------------------------

module "raspberry_client" {

  source = "./external_mig_without_health_check"

  service_name = "raspberry-client"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.artifacts_bucket}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -b ${var.artifacts_bucket} -a raspberry_client.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    NEXT_LOG_LEVEL=4
    NEXT_CUSTOMER_PUBLIC_KEY=leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==
    RASPBERRY_BACKEND_URL="http://${module.raspberry_backend.address}"
    RASPBERRY_NUM_CLIENTS=1
    EOF
    sudo gsutil cp ${var.artifacts_bucket}/libnext5.so /usr/local/lib/libnext5.so
    sudo ldconfig
    sudo systemctl start app.service
  EOF1

  machine_type       = var.machine_type
  git_hash           = var.git_hash
  project            = var.project
  region             = var.region
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.service_account
  tags               = ["allow-ssh"]
}

# ----------------------------------------------------------------------------------------
