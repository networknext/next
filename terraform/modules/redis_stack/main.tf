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
variable "machine_type" { type = string }
variable "project" { type = string }
variable "region" { type = string }
variable "zone" { type = string }
variable "default_network" { type = string }
variable "default_subnetwork" { type = string }
variable "service_account" { type = string }
variable "tags" { type = list }

# ----------------------------------------------------------------------------------------

resource "google_compute_instance" "service" {

  name         = var.service_name
  machine_type = var.machine_type
  zone         = var.zone
  tags         = var.tags

  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }

  network_interface {
    network    = var.default_network
    subnetwork = var.default_subnetwork
  }

  metadata = {
    startup-script = <<-EOF
    #! /bin/bash
    set -euo pipefail
    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg
    chmod 644 /usr/share/keyrings/redis-archive-keyring.gpg
    echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/redis.list
    apt-get update -y
    apt-get install -y redis-stack-server
    echo "protected-mode no" >> /etc/redis-stack.conf
    systemctl enable redis-stack-server
    systemctl start redis-stack-server
    EOF
  }

  lifecycle {
    create_before_destroy = true
  }

  service_account {
    email  = var.service_account
    scopes = ["cloud-platform"]
  }
}

output "address" {
  description = "The internal IP address of the redis server"
  value = google_compute_instance.service.network_interface.0.network_ip
}

# ----------------------------------------------------------------------------------------
