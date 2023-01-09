terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "3.20.0"
    }
  }
}

provider "google" {
  credentials = file(var.credentials_file)

  project = var.project
  region  = var.region
  zone    = var.zone
}

/****
 Magic backend instance template
****/
resource "google_compute_instance_template" "magic_backend_instance_template" {
  name        = "magic-backend-instance-template"
  description = "This template is used to create magic backend instances."

  labels = {
    environment = "dev"
  }

  instance_description = "description assigned to instances"
  machine_type         = "custom-1-1024"
  can_ip_forward       = false


  # Create a new boot disk from an image
  disk {
    source_image = "ubuntu-os-cloud/ubuntu-2004-lts"
    auto_delete  = true
    boot         = true
  }

  network_interface {
    network = "default"
  }

  metadata_startup_script = "#!/bin/bash\n mkdir -p /app\n cd /app\n gsutil cp gs://dev5_artifacts/bootstrap.sh .\n chmod +x bootstrap.sh\n ./bootstrap.sh -b gs://dev5_artifacts -a magic_backend.dev.tar.gz"

  metadata = {
    shutdown_script = "#!/bin/bash\n cd /app\n # Stopping the service sends a SIGTERM\n systemctl stop app.service"
  }

  service_account {
    # Google recommends custom service accounts that have cloud-platform scope and permissions granted via IAM Roles.
    email  = "dev5-terraform@dev-5-373713.iam.gserviceaccount.com"
    scopes = ["cloud-platform"]
  }
}

/****
 Magic backend MIG Health Check
****/
resource "google_compute_health_check" "magic_backend_health_check" {
  name = "magic-backend-mig-health-check"

  description = "Health check via http"

  timeout_sec         = 10
  check_interval_sec  = 10
  healthy_threshold   = 3
  unhealthy_threshold = 3

  http_health_check {
    port         = 80
    request_path = "/health"
  }
}

/****
Magic backend MIG
****/
resource "google_compute_instance_group_manager" "magic_backend_mig" {
  name = "magic-backend-mig"

  base_instance_name = "magic-backend"
  zone               = "us-central1-a"
  target_size        = 0

  version {
    name              = "magic-backend"
    instance_template = google_compute_instance_template.magic_backend_instance_template.id

  }

  auto_healing_policies {
    health_check      = google_compute_health_check.magic_backend_health_check.id
    initial_delay_sec = 300
  }
}

/*****
AUTOSCALE
******/

resource "google_compute_autoscaler" "autoscale_magic_backend" {
  name   = "autoscale-magic-backend"
  zone   = "us-central1-a"
  target = google_compute_instance_group_manager.magic_backend_mig.id

  autoscaling_policy {
    max_replicas    = 10
    min_replicas    = 0
    cooldown_period = 60

    cpu_utilization {
      target = 0.6
    }
  }
}
