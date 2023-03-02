# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

provider "google" {
  credentials = file("~/Documents/terraform.json")
  project = "heroic-grove-379322"
  region  = "us-central1"
  zone    = "us-central1-c"
}

# ----------------------------------------------------------------------------------------

resource "google_compute_instance" "test" {
  name         = "test"
  provider     = google-beta
  project      = "heroic-grove-379322"
  zone         = "us-central1-a"
  machine_type = "f1-micro"
  network_interface {
    network    = google_compute_network.network-next.id
    subnetwork = google_compute_subnetwork.network-next.id
  }
  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }
}

# ----------------------------------------------------------------------------------------

resource "google_compute_network" "network-next" {
  name                    = "network-next"
  provider                = google-beta
  project                 = "heroic-grove-379322"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "network-next" {
  name          = "network-next"
  provider      = google-beta
  project       = "heroic-grove-379322"
  ip_cidr_range = "10.0.0.0/24"
  region        = "us-central1"
  network       = google_compute_network.network-next.id
}

resource "google_compute_firewall" "network-next" {
  name          = "network-next"
  provider      = google-beta
  project       = "heroic-grove-379322"
  direction     = "INGRESS"
  network       = google_compute_network.network-next.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "35.235.240.0/20"]
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
}

# ----------------------------------------------------------------------------------------

resource "google_compute_subnetwork" "magic-backend" {
  name          = "magic-backend-subnet"
  provider      = google-beta
  project       = "heroic-grove-379322"
  ip_cidr_range = "10.0.1.0/24"
  region        = "us-central1"
  purpose       = "INTERNAL_HTTPS_LOAD_BALANCER"
  role          = "ACTIVE"
  network       = google_compute_network.network-next.id
}

resource "google_compute_forwarding_rule" "magic-backend" {
  name                  = "magic-backend"
  provider              = google-beta
  project               = "heroic-grove-379322"
  region                = "us-central1"
  depends_on            = [google_compute_subnetwork.magic-backend]
  ip_protocol           = "TCP"
  load_balancing_scheme = "INTERNAL_MANAGED"
  port_range            = "80"
  target                = google_compute_region_target_http_proxy.magic-backend.id
  network               = google_compute_network.network-next.id
  subnetwork            = google_compute_subnetwork.network-next.id
  network_tier          = "PREMIUM"
}

resource "google_compute_region_target_http_proxy" "magic-backend" {
  name     = "magic-backend"
  provider = google-beta
  project  = "heroic-grove-379322"
  region   = "us-central1"
  url_map  = google_compute_region_url_map.magic-backend.id
}

resource "google_compute_region_url_map" "magic-backend" {
  name            = "magic-backend"
  provider        = google-beta
  project         = "heroic-grove-379322"
  region          = "us-central1"
  default_service = google_compute_region_backend_service.magic-backend.id
}

resource "google_compute_region_backend_service" "magic-backend" {
  name                  = "magic-backend"
  provider              = google-beta
  project               = "heroic-grove-379322"
  region                = "us-central1"
  protocol              = "HTTP"
  load_balancing_scheme = "INTERNAL_MANAGED"
  timeout_sec           = 10
  health_checks         = [google_compute_region_health_check.magic-backend.id]
  backend {
    group           = google_compute_region_instance_group_manager.network-next.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
}

resource "google_compute_instance_template" "instance_template" {
  name         = "magic-backend"
  provider     = google-beta
  project      = "heroic-grove-379322"
  machine_type = "f1-micro"
  tags         = ["http-server"]

  network_interface {
    network    = google_compute_network.network-next.id
    subnetwork = google_compute_subnetwork.network-next.id
  }
  disk {
    source_image = "debian-cloud/debian-10"
    auto_delete  = true
    boot         = true
  }

  # install nginx and serve a simple web page
  metadata = {
    startup-script = <<-EOF1
      #! /bin/bash
      set -euo pipefail

      export DEBIAN_FRONTEND=noninteractive
      apt-get update
      apt-get install -y nginx-light jq

      NAME=$(curl -H "Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/hostname")
      IP=$(curl -H "Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip")
      METADATA=$(curl -f -H "Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/attributes/?recursive=True" | jq 'del(.["startup-script"])')

      cat <<EOF > /var/www/html/index.html
      <pre>
      Name: $NAME
      IP: $IP
      Metadata: $METADATA
      </pre>
      EOF
    EOF1
  }
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_region_health_check" "magic-backend" {
  name     = "magic-backend"
  provider = google-beta
  project  = "heroic-grove-379322"
  region   = "us-central1"
  http_health_check {
    port_specification = "USE_SERVING_PORT"
  }
}

resource "google_compute_region_instance_group_manager" "network-next" {
  name     = "magic-backend"
  provider = google-beta
  project  = "heroic-grove-379322"
  region   = "us-central1"
  version {
    instance_template = google_compute_instance_template.magic-backend.id
    name              = "primary"
  }
  base_instance_name = "magic-backend"
  target_size        = 2
}

resource "google_compute_firewall" "magic-backend" {
  name          = "magic-backend"
  provider      = google-beta
  project       = "heroic-grove-379322"
  direction     = "INGRESS"
  network       = google_compute_network.network-next.id
  source_ranges = ["10.0.1.0/24"]
  target_tags   = ["http-server"]
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
}

# ----------------------------------------------------------------------------------------
