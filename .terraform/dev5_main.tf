terraform {
  required_version = ">= 0.13"
  required_providers {
    google = {
      source = "hashicorp/google"
    }
    google-beta = {
      source = "hashicorp/google-beta"
    }
    random = {
      source = "hashicorp/random"
    }
    tls = {
      source = "hashicorp/tls"
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
  target_size        = 2

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


/****
HTTP load balancer
****/
module "gce-lb-http" {
  source            = "GoogleCloudPlatform/lb-http/google"
  name              = "magic-backend-mig-http-lb"
  project           = var.project
  #target_tags       = ["default"]
  #firewall_networks = ["default"]


  backends = {
    default = {
      description                     = null
      protocol                        = "HTTP"
      port                            = 80
      port_name                       = "http"
      timeout_sec                     = 10
      connection_draining_timeout_sec = null
      enable_cdn                      = false
      compression_mode                = null
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
        request_path        = "/health"
        port                = 80
        host                = null
        logging             = null
      }

      log_config = {
        enable      = false
        sample_rate = null
      }

      groups = [
        {
          group                        = google_compute_instance_group_manager.magic_backend_mig.instance_group
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
