# ----------------------------------------------------------------------------------------

variable "credentials" { type = string }
variable "project" { type = string }
variable "location" { type = string }
variable "region" { type = string }
variable "zone" { type = string }
variable "service_account" { type = string }
variable "dev_artifacts_bucket" { type = string }
variable "machine_type" { type = string }

# ----------------------------------------------------------------------------------------

resource "random_string" "run_id" {
  length  = 8
  lower   = true
  upper   = false
  numeric = false
  special = false
}

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
  name         = "magic-backend-${random_string.run_id.result}"
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

/*
module "gce-lb-http" {
  source  = "GoogleCloudPlatform/lb-http/google"
  version = "~> 6.0"
  name    = var.network_prefix
  project = var.project
  target_tags = [
    "${var.network_prefix}-group1",
    module.cloud-nat-group1.router_name,
    "${var.network_prefix}-group2",
    module.cloud-nat-group2.router_name
  ]
  firewall_networks = [google_compute_network.default.name]

  backends = {
    default = {

      description                     = null
      protocol                        = "HTTP"
      port                            = 80
      port_name                       = "http"
      timeout_sec                     = 10
      connection_draining_timeout_sec = null
      enable_cdn                      = false
      security_policy                 = null
      session_affinity                = null
      affinity_cookie_ttl_sec         = null
      custom_request_headers          = null
      custom_response_headers         = null

      health_check = {
        check_interval_sec  = null
        timeout_sec         = null
        healthy_threshold   = null
        unhealthy_threshold = null
        request_path        = "/"
        port                = 80
        host                = null
        logging             = null
      }

      log_config = {
        enable      = true
        sample_rate = 1.0
      }

      groups = [
        {
          group                        = module.mig1.instance_group
          balancing_mode               = null
          capacity_scaler              = null
          description                  = null
          max_connections              = null
          max_connections_per_instance = null
          max_connections_per_endpoint = null
          max_rate                     = null
          max_rate_per_instance        = null
          max_rate_per_endpoint        = null
          max_utilization              = null
        }
      ]

      iap_config = {
        enable               = false
        oauth2_client_id     = ""
        oauth2_client_secret = ""
      }
    }
  }
}
*/

# ----------------------------------------------------------------------------------------













































// ---------------------------------------------------------------------------------------

/*
resource "google_sql_database_instance" "postgres" {
  name = "postgres"
  database_version = "POSTGRES_14"
  region = "${var.region}"
  settings {
    tier = "db-f1-micro"
    ip_configuration {
      private_network = google_compute_network.development.id
    }
  }
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

# ----------------------------------------------------------------------------------------

/*
resource "google_compute_network" "private_network" {
  provider = google-beta

  name = "private-network"
}

resource "google_compute_global_address" "private_ip_address" {
  provider = google-beta

  name          = "private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.private_network.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  provider = google-beta

  network                 = google_compute_network.private_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

resource "random_id" "db_name_suffix" {
  byte_length = 4
}

resource "google_sql_database_instance" "instance" {
  provider = google-beta

  name             = "private-instance-${random_id.db_name_suffix.hex}"
  region           = "us-central1"
  database_version = "MYSQL_5_7"

  depends_on = [google_service_networking_connection.private_vpc_connection]

  settings {
    tier = "db-f1-micro"
    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = google_compute_network.private_network.id
      enable_private_path_for_google_cloud_services = true
    }
  }
}

provider "google-beta" {
  region = "us-central1"
  zone   = "us-central1-a"
}
*/