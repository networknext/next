# ----------------------------------------------------------------------------------------

variable "tag" { type = string }

variable "extra" { 
  type = string 
  default = ""
}

variable "vpn_address" { type = string }

variable "google_credentials" { type = string }
variable "google_project" { type = string }
variable "google_location" { type = string }
variable "google_region" { type = string }
variable "google_zone" { type = string }
variable "google_service_account" { type = string }
variable "google_artifacts_bucket" { type = string }
variable "google_database_bucket" { type = string }
variable "google_machine_type" { type = string }

variable "cloudflare_api_token" { type = string }
variable "cloudflare_zone_id_api" { type = string }
variable "cloudflare_zone_id_relay_backend" { type = string }
variable "cloudflare_zone_id_server_backend" { type = string }

variable "relay_backend_public_key" { type = string }
variable "relay_backend_private_key" { type = string }
variable "server_backend_public_key" { type = string }
variable "server_backend_private_key" { type = string }
variable "ping_key" { type = string }
variable "api_private_key" { type = string }
variable "customer_public_key" { type = string }
variable "customer_private_key" { type = string }

variable "maxmind_license_key" { type = string }

variable "ip2location_bucket_name" { type = string }

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.51.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
  backend "gcs" {
    bucket  = "test_network_next_terraform_state"
    prefix  = "terraform/state"
  }
}

provider "google" {
  credentials = file(var.google_credentials)
  project     = var.google_project
  region      = var.google_region
  zone        = var.google_zone
}

provider "cloudflare" {
  api_token = file(var.cloudflare_api_token)
}

# ----------------------------------------------------------------------------------------

resource "cloudflare_record" "api_domain" {
  zone_id = var.cloudflare_zone_id_api
  name    = "dev"
  value   = module.api.address
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "server_backend_domain" {
  zone_id = var.cloudflare_zone_id_server_backend
  name    = "dev"
  value   = module.server_backend.address
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "relay_backend_domain" {
  zone_id = var.cloudflare_zone_id_relay_backend
  name    = "dev"
  value   = module.relay_gateway.address
  type    = "A"
  proxied = false
}

# ----------------------------------------------------------------------------------------

resource "google_compute_network" "development" {
  name                    = "development"
  project                 = var.google_project
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "development" {
  name                     = "development"
  project                  = var.google_project
  ip_cidr_range            = "10.0.0.0/16"
  region                   = var.google_region
  network                  = google_compute_network.development.id
  private_ip_google_access = true
}

resource "google_compute_subnetwork" "internal_http_load_balancer" {
  name          = "internal-http-load-balancer"
  project       = var.google_project
  region        = var.google_region
  purpose       = "INTERNAL_HTTPS_LOAD_BALANCER"
  role          = "ACTIVE"
  network       = google_compute_network.development.id
  ip_cidr_range = "10.1.0.0/16"
}

# ----------------------------------------------------------------------------------------

resource "google_compute_firewall" "allow_ssh" {
  name          = "allow-ssh"
  project       = var.google_project
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
  project       = var.google_project
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
  project       = var.google_project
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
  project       = var.google_project
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
  project       = var.google_project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["${var.vpn_address}/32"]
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
  target_tags = ["allow-http-vpn-only"]
}

resource "google_compute_firewall" "allow_udp_40000" {
  name          = "allow-udp-40000"
  project       = var.google_project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
    ports    = ["40000"]
  }
  target_tags = ["allow-udp-40000"]
}

resource "google_compute_firewall" "allow_udp_all" {
  name          = "allow-udp-all"
  project       = var.google_project
  direction     = "INGRESS"
  network       = google_compute_network.development.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
  }
  target_tags = ["allow-udp-all"]
}

# ----------------------------------------------------------------------------------------

resource "google_redis_instance" "redis_portal" {
  name               = "redis-portal"
  tier               = "BASIC"
  memory_size_gb     = 1
  region             = "us-central1"
  redis_version      = "REDIS_6_X"
  redis_configs      = { "activedefrag" = "yes", "maxmemory-policy" = "allkeys-lru" }
  authorized_network = google_compute_network.development.id
}

resource "google_redis_instance" "redis_map_cruncher" {
  name               = "redis-map-cruncher"
  tier               = "BASIC"
  memory_size_gb     = 1
  region             = "us-central1"
  redis_version      = "REDIS_6_X"
  redis_configs      = { "activedefrag" = "yes", "maxmemory-policy" = "allkeys-lru" }
  authorized_network = google_compute_network.development.id
}

resource "google_redis_instance" "redis_raspberry" {
  name               = "redis-raspberry"
  tier               = "BASIC"
  memory_size_gb     = 1
  region             = "us-central1"
  redis_version      = "REDIS_6_X"
  redis_configs      = { "activedefrag" = "yes", "maxmemory-policy" = "allkeys-lru" }
  authorized_network = google_compute_network.development.id
}

resource "google_redis_instance" "redis_analytics" {
  name               = "redis-analytics"
  tier               = "BASIC"
  memory_size_gb     = 1
  region             = "us-central1"
  redis_version      = "REDIS_6_X"
  redis_configs      = { "activedefrag" = "yes", "maxmemory-policy" = "allkeys-lru" }
  authorized_network = google_compute_network.development.id
}

resource "google_redis_instance" "redis_relay_backend" {
  name               = "redis-relay-backend"
  tier               = "BASIC"
  memory_size_gb     = 1
  region             = "us-central1"
  redis_version      = "REDIS_6_X"
  authorized_network = google_compute_network.development.id
}

resource "google_redis_instance" "redis_server_backend" {
  name               = "redis-server-backend"
  tier               = "BASIC"
  memory_size_gb     = 1
  region             = "us-central1"
  redis_version      = "REDIS_6_X"
  authorized_network = google_compute_network.development.id
}

output "redis_portal_address" {
  description = "The IP address of the portal redis instance"
  value       = google_redis_instance.redis_portal.host
}

output "redis_map_cruncher_address" {
  description = "The IP address of the map cruncher redis instance"
  value       = google_redis_instance.redis_map_cruncher.host
}

output "redis_raspberry_address" {
  description = "The IP address of the raspberry redis instance"
  value       = google_redis_instance.redis_raspberry.host
}

output "redis_analytics_address" {
  description = "The IP address of the analytics redis instance"
  value       = google_redis_instance.redis_analytics.host
}

output "redis_relay_backend_address" {
  description = "The IP address of the relay backend redis instance"
  value       = google_redis_instance.redis_relay_backend.host
}

output "redis_server_backend_address" {
  description = "The IP address of the server backend redis instance"
  value       = google_redis_instance.redis_server_backend.host
}

# ----------------------------------------------------------------------------------------

locals {

  pubsub_channels = [
    "cost_matrix_update",
    "route_matrix_update",
    "relay_to_relay_ping",
    "relay_update",
    "server_init",
    "server_update",
    "near_relay_ping",
    "session_update",
    "session_summary",
    "match_data",
    "cost_matrix_stats",
  ]
  
}

resource "google_pubsub_topic" "pubsub_topic" {
  count = length(local.pubsub_channels)
  name  = local.pubsub_channels[count.index]
} 

resource "google_pubsub_subscription" "pubsub_subscription" {
  count                       = length(local.pubsub_channels)
  name                        = local.pubsub_channels[count.index]
  topic                       = google_pubsub_topic.pubsub_topic[count.index].name
  message_retention_duration  = "604800s"
  retain_acked_messages       = true
  ack_deadline_seconds        = 60
  expiration_policy {
    ttl = ""
  }
}

# ----------------------------------------------------------------------------------------


locals {
  
  bigquery_tables = {
/*
    "session_update" = <<EOF
    [
      {
        "name": "timestamp",
        "type": "TIMESTAMP",
        "mode": "REQUIRED",
        "description": "The timestamp when the session update occurred"
      },
      {
        "name": "session_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Unique identifier for this session"
      },
      {
        "name": "slice_number",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Slices are 10 second periods starting from slice number 0 at the start of the session"
      },
      {
        "name": "real_packet_loss",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Packet loss between the client and the server measured from game packets (%)"
      },
      {
        "name": "real_jitter",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Jitter between the client and the server measured from game packets (milliseconds)"
      },
      {
        "name": "real_out_of_order",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Percentage of packets that arrive out of order between the client and the server (%)"
      },
      {
        "name": "session_flags",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Session flags are used to diagnose what's happening with a session. Look up SessionFlags_ in the codebase for a list of flags"
      },
      {
        "name": "session_events",
        "type": "INT64",
        "mode": "NULLABLE",
        "description": "Customer specified set of 64bit event flags. Optional. NULL if no flags are set"
      },
      {
        "name": "internal_events",
        "type": "INT64",
        "mode": "NULLABLE",
        "description": "Internal SDK event flags. Optional. NULL if no flags are set"
      },
      {
        "name": "direct_rtt",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Latency between client and server as measured by direct pings (unaccelerated path). Milliseconds. IMPORTANT: Will be 0.0 on slice 0 always. Ignore. Not known yet"
      },
      {
        "name": "direct_jitter",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Jitter between client and server as measured by direct pings (unaccelerated path). Milliseconds"
      },
      {
        "name": "direct_packet_loss",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Packet loss between client and server as measured by direct pings (unaccelerated path). Percent"
      },
      {
        "name": "direct_kbps_up",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Bandwidth in the client to server direction along the direct path (unaccelerated). Kilobits per-second"
      },
      {
        "name": "direct_kbps_down",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Bandwidth in the client to server direction along the direct path (unaccelerated). Kilobits per-second"
      },
      {
        "name": "next",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if this slice went over network next"
      },
      {
        "name": "next_rtt",
        "type": "FLOAT64",
        "mode": "NULLABLE",
        "description": "Latency between client and server as measured by next pings (accelerated path). Milliseconds. NULL if not on network next"
      },
      {
        "name": "next_jitter",
        "type": "FLOAT64",
        "mode": "NULLABLE",
        "description": "Jitter between client and server as measured by next pings (accelerated path). Milliseconds. NULL if not on network next"
      },
      {
        "name": "next_packet_loss",
        "type": "FLOAT64",
        "mode": "NULLABLE",
        "description": "Packet loss between client and server as measured by next pings (accelerated path). Percent. NULL if not on network next"
      },
      {
        "name": "next_kbps_up",
        "type": "INT64",
        "mode": "NULLABLE",
        "description": "Bandwidth in the client to server direction along the next path (accelerated). Kilobits per-second"
      },
      {
        "name": "next_kbps_down",
        "type": "INT64",
        "mode": "NULLABLE",
        "description": "Bandwidth in the server to client direction along the next path (accelerated). Kilobits per-second"
      },
      {
        "name": "next_predicted_rtt",
        "type": "FLOAT64",
        "mode": "NULLABLE",
        "description": "Predicted latency between client and server from the control plane. Milliseconds. NULL if not on network next"
      },
      {
        "name": "next_route_relays",
        "type": "INT64",
        "mode": "REPEATED",
        "description": "Array of relay ids for the network next path (accelerated). NULL if not on network next"
      }
    ]
    EOF

    "session_summary" = <<EOF
    [
      {
        "name": "timestamp",
        "type": "TIMESTAMP",
        "mode": "REQUIRED",
        "description": "The timestamp when the server update occurred"
      },
      {
        "name": "session_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Unique identifier for this session"
      },
      {
        "name": "match_id",
        "type": "INT64",
        "mode": "NULLABLE",
        "description": "Match id if set on the server for this session. NULL if not specified."
      },
      {
        "name": "datacenter_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The datacenter the server is in"
      },
      {
        "name": "buyer_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The buyer this session belongs to"
      },
      {
        "name": "user_hash",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Pseudonymized hash of a unique user id passed up from the SDK"
      },
      {
        "name": "latitude",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Approximate latitude of the player from ip2location"
      },
      {
        "name": "longitude",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Approximate longitude of the player from ip2location"
      },
      {
        "name": "client_address",
        "type": "STRING",
        "mode": "REQUIRED",
        "description": "Client address and port number"
      },
      {
        "name": "server_address",
        "type": "STRING",
        "mode": "REQUIRED",
        "description": "Server address and port"
      },
      {
        "name": "connection_type",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Connection type: 0 = unknown, 1 = wired, 2 = wifi, 3 = cellular"
      },
      {
        "name": "platform_type",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Platform type: 0 = unknown, 1 = windows, 2 = mac, 3 = linux, 4 = switch, 5 = ps4, 6 = ios, 7 = xbox one, 8 = xbox series x, 9 = ps5"
      },
      {
        "name": "sdk_version_major",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The major SDK version on the server"
      },
      {
        "name": "sdk_version_minor",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The minor SDK version on the server"
      },
      {
        "name": "sdk_version_patch",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The patch SDK version on the server"
      },
      {
        "name": "client_to_server_packets_sent",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The total number of game packets sent from client to server in this session"
      },
      {
        "name": "server_to_client_packets_sent",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The total number of game packets sent from server to client in this session"
      },
      {
        "name": "client_to_server_packets_lost",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The total number of game packets lost from client to server in this session"
      },
      {
        "name": "server_to_client_packets_lost",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The total number of game packets lost from server to client in this session"
      },
      {
        "name": "client_to_server_packets_out_of_order",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The total number of game packets received out of order from client to server in this session"
      },
      {
        "name": "server_to_client_packets_out_of_order",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The total number of game packets received out of order from server to client in this session"
      },
      {
        "name": "total_next_envelope_bytes_up",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The total number of envelope bytes sent across network next in the client to server direction for this session"
      },
      {
        "name": "total_next_envelope_bytes_down",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The total number of envelope bytes sent across netwnork next in the server to client direction for this session"
      },
      {
        "name": "duration_on_next",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Total time spent on network next in this session (time accelerated). Seconds"
      },
      {
        "name": "session_duration",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Length of this session in seconds"
      },
      {
        "name": "start_timestamp",
        "type": "TIMESTAMP",
        "mode": "REQUIRED",
        "description": "The time when this session started"
      }
    ]
    EOF

    "server_update" = <<EOF
    [
      {
        "name": "timestamp",
        "type": "TIMESTAMP",
        "mode": "REQUIRED",
        "description": "The timestamp when the server update occurred"
      },
      {
        "name": "sdk_version_major",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The major SDK version number on the server"
      },
      {
        "name": "sdk_version_minor",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The minor SDK version number on the server"
      },
      {
        "name": "sdk_version_patch",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The patch SDK version number on the server"
      },
      {
        "name": "buyer_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The buyer this server belongs to"
      },
      {
        "name": "datacenter_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The datacenter this server is in"
      },
      {
        "name": "match_id",
        "type": "INT64",
        "mode": "NULLABLE",
        "description": "The current match id on the server (optional: NULL if not specified)"
      },
      {
        "name": "num_sessions",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The number of client sessions currently connected to the server"
      },
      {
        "name": "server_address",
        "type": "STRING",
        "mode": "REQUIRED",
        "description": "The address and port of the server, for example: '123.254.10.5:40000'"
      }
    ]
    EOF

    "server_init" = <<EOF
    [
      {
        "name": "timestamp",
        "type": "TIMESTAMP",
        "mode": "REQUIRED",
        "description": "The timestamp when the server init occurred"
      },
      {
        "name": "sdk_version_major",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The major SDK version number on the server"
      },
      {
        "name": "sdk_version_minor",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The minor SDK version number on the server"
      },
      {
        "name": "sdk_version_patch",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The patch SDK version number on the server"
      },
      {
        "name": "buyer_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The buyer this server belongs to"
      },
      {
        "name": "datacenter_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The datacenter this server is in"
      },
      {
        "name": "datacenter_name",
        "type": "STRING",
        "mode": "REQUIRED",
        "description": "The name of the datacenter, for example: 'google.iowa.1'"
      },
      {
        "name": "server_address",
        "type": "STRING",
        "mode": "REQUIRED",
        "description": "The address and port of the server, for example: '123.254.10.5:40000'"
      }
    ]
    EOF

    "relay_update" = <<EOF
    [
      {
        "name": "timestamp",
        "type": "TIMESTAMP",
        "mode": "REQUIRED",
        "description": "The timestamp when the relay update occurred"
      },
      {
        "name": "relay_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Unique relay id. The fnv1a hash of the relay address + port as a string"
      },
      {
        "name": "session_count",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The number of sessions currently going through this relay"
      },
      {
        "name": "max_sessions",
        "type": "INT64",
        "mode": "NULLABLE",
        "description": "The maximum number of sessions allowed through this relay (optional: NULL if not specified)"
      },
      {
        "name": "envelope_bandwidth_up_kbps",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The current amount of envelope bandwidth in the client to server direction through this relay"
      },
      {
        "name": "envelope_bandwidth_down_kbps",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The current amount of envelope bandwidth in the server to client direction through this relay"
      },
      {
        "name": "packets_sent_per_second",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The number of packets sent per-second by this relay"
      },
      {
        "name": "packets_received_per_second",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The number of packets received per-second by this relay"
      },
      {
        "name": "bandwidth_sent_kbps",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The amount of bandwidth sent by this relay in kilobits per-second"
      },
      {
        "name": "bandwidth_received_kbps",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The amount of bandwidth received by this relay in kilobits per-second"
      },
      {
        "name": "near_pings_per_second",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The number of near relay pings received by this relay per-second"
      },
      {
        "name": "relay_pings_per_second",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The number of relay pings sent from other relays received by this relay per-second"
      },
      {
        "name": "relay_flags",
        "type": "INT64",
        "mode": "NULLABLE",
        "description": "The current value of the relay flags. See RelayFlags_* in the source code"
      },
      {
        "name": "num_routable",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The number of other relays this relay can route to"
      },
      {
        "name": "num_unroutable",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The number of other relays this relay cannot route to"
      },
      {
        "name": "start_time",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The start time of the relay as a unix timestamp according to the clock on the relay"
      },
      {
        "name": "current_time",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The start time of the relay as a unix timestamp according to the clock on the relay. Together with start_time and timestamp this can be used to determine relay uptime, and clock desynchronization between the relay and the backend."
      },
      {
        "name": "relay_counters",
        "type": "INT64",
        "mode": "REPEATED",
        "description": "Array of counters used to diagnose what is going on with a relay. Search for RELAY_COUNTER_ in the codebase for counter names"
      }
    ]
    EOF

    "route_matrix_update" = <<EOF
    [
      {
        "name": "timestamp",
        "type": "TIMESTAMP",
        "mode": "REQUIRED",
        "description": "The timestamp when the route matrix update occurred"
      },
      {
        "name": "cost_matrix_size",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The size of the cost matrix in bytes"
      },
      {
        "name": "route_matrix_size",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The size of the route matrix in bytes"
      },
      {
        "name": "optimize_time",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Time it took produce this route matrix from the cost matrix (milliseconds)"
      },
      {
        "name": "num_relays",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The number of relays in the route matrix"
      },
      {
        "name": "num_dest_relays",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The number of destination relays in the route matrix"
      },
      {
        "name": "num_full_relays",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The number of full relays in the route matrix"
      },
      {
        "name": "num_datacenters",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The number of datacenters in the route matrix"
      },
      {
        "name": "total_routes",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The total number of routes in the route matrix"
      },
      {
        "name": "average_num_routes",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The average number of routes between any two relays"
      },
      {
        "name": "average_route_length",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The average number of relays per-route"
      },
      {
        "name": "no_route_percent",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs that have no route between them"
      },
      {
        "name": "one_route_percent",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with only one route between them"
      },
      {
        "name": "no_direct_route_percent",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with no direct route between them"
      },
      {
        "name": "rtt_bucket_no_improvement",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with no improvement"
      },
      {
        "name": "rtt_bucket_0_5ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 0-5ms reduction in latency"
      },
      {
        "name": "rtt_bucket_5_10ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 5-10ms reduction in latency"
      },
      {
        "name": "rtt_bucket_10_15ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 10-15ms reduction in latency"
      },
      {
        "name": "rtt_bucket_15_20ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 15-20ms reduction in latency"
      },
      {
        "name": "rtt_bucket_20_25ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 20-25ms reduction in latency"
      },
      {
        "name": "rtt_bucket_25_30ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 25-30ms reduction in latency"
      },
      {
        "name": "rtt_bucket_30_35ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 30-35ms reduction in latency"
      },
      {
        "name": "rtt_bucket_35_40ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 35-40ms reduction in latency"
      },
      {
        "name": "rtt_bucket_40_45ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 40-45ms reduction in latency"
      },
      {
        "name": "rtt_bucket_45_50ms",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 45-50ms reduction in latency"
      },
      {
        "name": "rtt_bucket_50ms_plus",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The percent of relay pairs with 50ms+ reduction in latency"
      }
    ]
    EOF

    "relay_to_relay_ping" = <<EOF
    [
      {
        "name": "timestamp",
        "type": "TIMESTAMP",
        "mode": "REQUIRED",
        "description": "The timestamp when the relay ping occurred"
      },
      {
        "name": "source_relay_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The id of the source relay"
      },
      {
        "name": "destination_relay_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The id of the destination relay"
      },
      {
        "name": "rtt",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Round trip latency between the two relays (milliseconds)"
      },
      {
        "name": "jitter",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Time variance in latency between the two relays (milliseconds)"
      },

      {
        "name": "packet_loss",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "The packet loss between the two relays (%)"
      }
    ]
    EOF

    "near_relay_ping" = <<EOF
    [
      {
        "name": "timestamp",
        "type": "TIMESTAMP",
        "mode": "REQUIRED",
        "description": "The timestamp when the relay update occurred"
      },
      {
        "name": "buyer_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The buyer this player belongs to"
      },
      {
        "name": "session_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Unique id for the session"
      },
      {
        "name": "match_id",
        "type": "INT64",
        "mode": "NULLABLE",
        "description": "Match id if currently set on the server. Optional. NULL if not specified"
      },
      {
        "name": "user_hash",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Pseudonymized hash of a user id passed up from the SDK"
      },
      {
        "name": "latitude",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Approximate latitude of the player from ip2location"
      },
      {
        "name": "longitude",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Approximate longitude of the player from ip2location"
      },
      {
        "name": "client_address",
        "type": "STRING",
        "mode": "REQUIRED",
        "description": "Client address and port number"
      },
      {
        "name": "connection_type",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Connection type: 0 = unknown, 1 = wired, 2 = wifi, 3 = cellular"
      },
      {
        "name": "platform_type",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Platform type: 0 = unknown, 1 = windows, 2 = mac, 3 = linux, 4 = switch, 5 = ps4, 6 = ios, 7 = xbox one, 8 = xbox series x, 9 = ps5"
      },
      {
        "name": "near_relay_id",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Relay id being pinged by the client"
      },
      {
        "name": "near_relay_rtt",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Round trip time ping between the client and the relay (milliseconds)"
      },
      {
        "name": "near_relay_jitter",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Jitter between the client and the relay (milliseconds)"
      },
      {
        "name": "near_relay_packet_loss",
        "type": "FLOAT64",
        "mode": "REQUIRED",
        "description": "Packet loss between the client and the relay (%)"
      }
    ]
    EOF
*/
  }

  bigquery_table_clustering = {

    "session_update" = [ "session_id" ]
    "session_summary" = [ "session_id", "buyer_id", "user_hash" ]
    "server_update" = [ "datacenter_id", "buyer_id" ]
    "server_init" = [ "datacenter_id", "buyer_id" ]
    "relay_update" = [ "relay_id" ]
    "route_matrix_update" = []
    "cost_matrix_update" = []
    "relay_to_relay_ping" = [ "source_relay_id" ]
    "near_relay_ping" = [ "near_relay_id", "user_hash" ]
  }
}

resource "google_bigquery_dataset" "dataset" {
  dataset_id                  = "analytics"
  friendly_name               = "Analytics"
  description                 = "This dataset contains Network Next raw analytics data. It is retained for 90 days."
  location                    = "US"
  default_table_expiration_ms = 7776000000 # 90 days
}

resource "google_bigquery_table" "table" {
  for_each            = local.bigquery_tables
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = each.key
  schema              = each.value
  clustering          = local.bigquery_table_clustering[each.key]
  deletion_protection = false
  time_partitioning {
    type = "DAY"
  }
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
  region = "${var.google_region}"
  depends_on = [google_service_networking_connection.postgres]
  settings {
    tier = "db-f1-micro"
    ip_configuration {
      ipv4_enabled    = "false"
      private_network = google_compute_network.development.id
    }
    database_flags {
      name  = "max_connections"
      value = "1024"
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

  source = "../../modules/internal_http_service"

  service_name = "magic-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a magic_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    EOF
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = var.google_machine_type
  project                    = var.google_project
  region                     = var.google_region
  default_network            = google_compute_network.development.id
  default_subnetwork         = google_compute_subnetwork.development.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-http"]
}

output "magic_backend_address" {
  description = "The IP address of the magic backend load balancer"
  value       = module.magic_backend.address
}

# ----------------------------------------------------------------------------------------

module "relay_gateway" {

  source = "../../modules/external_http_service"

  service_name = "relay-gateway"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh  -t ${var.tag} -b ${var.google_artifacts_bucket} -a relay_gateway.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    GOOGLE_PROJECT_ID=${var.google_project}
    REDIS_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.google_database_bucket}/dev.bin"
    DATABASE_PATH="/app/database.bin"
    RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
    RELAY_BACKEND_PRIVATE_KEY=${var.relay_backend_private_key}
    PING_KEY=${var.ping_key}
    EOF
    sudo gsutil cp ${var.google_database_bucket}/dev.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = var.google_machine_type
  project                  = var.google_project
  zone                     = var.google_zone
  default_network          = google_compute_network.development.id
  default_subnetwork       = google_compute_subnetwork.development.id
  service_account          = var.google_service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http"]
}

output "relay_gateway_address" {
  description = "The IP address of the relay gateway load balancer"
  value       = module.relay_gateway.address
}

# ----------------------------------------------------------------------------------------

module "relay_backend" {

  source = "../../modules/internal_http_service"

  service_name = "relay-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a relay_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    GOOGLE_PROJECT_ID=${var.google_project}
    REDIS_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.google_database_bucket}/dev.bin"
    DATABASE_PATH="/app/database.bin"
    INITIAL_DELAY=15s
    ENABLE_GOOGLE_PUBSUB=true
    EOF
    sudo gsutil cp ${var.google_database_bucket}/dev.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = var.google_machine_type
  project                    = var.google_project
  region                     = var.google_region
  default_network            = google_compute_network.development.id
  default_subnetwork         = google_compute_subnetwork.development.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-http"]

  depends_on = [google_pubsub_topic.pubsub_topic, google_pubsub_subscription.pubsub_subscription]
}

output "relay_backend_address" {
  description = "The IP address of the relay backend load balancer"
  value       = module.relay_backend.address
}

# ----------------------------------------------------------------------------------------

module "analytics" {

  source = "../../modules/internal_http_service"

  service_name = "analytics"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a analytics.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    GOOGLE_PROJECT_ID=${var.google_project}
    DATABASE_URL="${var.google_database_bucket}/dev.bin"
    DATABASE_PATH="/app/database.bin"
    COST_MATRIX_URL="http://${module.relay_backend.address}/cost_matrix"
    ROUTE_MATRIX_URL="http://${module.relay_backend.address}/route_matrix"
    REDIS_HOSTNAME="${google_redis_instance.redis_analytics.host}:6379"
    ENABLE_GOOGLE_PUBSUB=true
    EOF
    sudo gsutil cp ${var.google_database_bucket}/dev.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = var.google_machine_type
  project                    = var.google_project
  region                     = var.google_region
  default_network            = google_compute_network.development.id
  default_subnetwork         = google_compute_subnetwork.development.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks"]

  depends_on = [google_pubsub_topic.pubsub_topic, google_pubsub_subscription.pubsub_subscription, google_bigquery_table.table]
}

output "analytics_address" {
  description = "The IP address of the analytics load balancer"
  value       = module.analytics.address
}

# ----------------------------------------------------------------------------------------

module "api" {

  source = "../../modules/external_http_service"

  service_name = "api"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a api.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_PORTAL_HOSTNAME="${google_redis_instance.redis_portal.host}:6379"
    REDIS_RELAY_BACKEND_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    GOOGLE_PROJECT_ID=${var.google_project}
    DATABASE_URL="${var.google_database_bucket}/dev.bin"
    DATABASE_PATH="/app/database.bin"
    PGSQL_CONFIG="host=${google_sql_database_instance.postgres.ip_address.0.ip_address} port=5432 user=developer password=developer dbname=database sslmode=disable"
    API_PRIVATE_KEY=${var.api_private_key}
    ALLOWED_ORIGIN="*"
    EOF
    sudo gsutil cp ${var.google_database_bucket}/dev.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = var.google_machine_type
  project                  = var.google_project
  zone                     = var.google_zone
  default_network          = google_compute_network.development.id
  default_subnetwork       = google_compute_subnetwork.development.id
  service_account          = var.google_service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http-vpn-only"]
}

output "api_address" {
  description = "The IP address of the api load balancer"
  value       = module.api.address
}

// ---------------------------------------------------------------------------------------

module "portal_cruncher" {

  source = "../../modules/internal_mig_with_health_check"

  service_name = "portal-cruncher"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a portal_cruncher.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_PORTAL_HOSTNAME="${google_redis_instance.redis_portal.host}:6379"
    REDIS_RELAY_BACKEND_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    REDIS_SERVER_BACKEND_HOSTNAME="${google_redis_instance.redis_server_backend.host}:6379"
    IP2LOCATION_BUCKET_NAME=${var.ip2location_bucket_name}
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = var.google_machine_type
  project            = var.google_project
  region             = var.google_region
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-http"]
  target_size        = 2
}

// ---------------------------------------------------------------------------------------

module "map_cruncher" {

  source = "../../modules/internal_mig_with_health_check"

  service_name = "map-cruncher"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a map_cruncher.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_HOSTNAME="${google_redis_instance.redis_map_cruncher.host}:6379"
    REDIS_SERVER_BACKEND_HOSTNAME="${google_redis_instance.redis_server_backend.host}:6379"
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = var.google_machine_type
  project            = var.google_project
  region             = var.google_region
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-http"]
}

# ----------------------------------------------------------------------------------------

module "server_backend" {

  source = "../../modules/external_udp_service"

  service_name = "server-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a server_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    UDP_PORT=40000
    UDP_BIND_ADDRESS="##########:40000"
    GOOGLE_PROJECT_ID=${var.google_project}
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    REDIS_HOSTNAME="${google_redis_instance.redis_server_backend.host}:6379"
    RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
    RELAY_BACKEND_PRIVATE_KEY=${var.relay_backend_private_key}
    SERVER_BACKEND_ADDRESS="##########:40000"
    SERVER_BACKEND_PUBLIC_KEY=${var.server_backend_public_key}
    SERVER_BACKEND_PRIVATE_KEY=${var.server_backend_private_key}
    ROUTE_MATRIX_URL="http://${module.relay_backend.address}/route_matrix"
    PING_KEY=${var.ping_key}
    IP2LOCATION_BUCKET_NAME=${var.ip2location_bucket_name}
    ENABLE_GOOGLE_PUBSUB=true
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = var.google_machine_type
  project            = var.google_project
  region             = var.google_region
  port               = 40000
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-udp-40000"]
  target_size        = 2

  depends_on = [google_pubsub_topic.pubsub_topic, google_pubsub_subscription.pubsub_subscription]
}

output "server_backend_address" {
  description = "The IP address of the server backend load balancer"
  value       = module.server_backend.address
}

# ----------------------------------------------------------------------------------------

module "raspberry_backend" {

  source = "../../modules/external_http_service"

  service_name = "raspberry-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a raspberry_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_HOSTNAME="${google_redis_instance.redis_raspberry.host}:6379"
    EOF
    sudo systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = var.google_machine_type
  project                  = var.google_project
  zone                     = var.google_zone
  default_network          = google_compute_network.development.id
  default_subnetwork       = google_compute_subnetwork.development.id
  service_account          = var.google_service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http"]
}

output "raspberry_backend_address" {
  description = "The IP address of the raspberry backend load balancer"
  value       = module.raspberry_backend.address
}

# ----------------------------------------------------------------------------------------

module "raspberry_server" {

  source = "../../modules/external_mig_without_health_check"

  service_name = "raspberry-server"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a raspberry_server.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    NEXT_LOG_LEVEL=4
    NEXT_DATACENTER=cloud
    NEXT_CUSTOMER_PRIVATE_KEY=${var.customer_private_key}
    RASPBERRY_BACKEND_URL="http://${module.raspberry_backend.address}"
    EOF
    sudo gsutil cp ${var.google_artifacts_bucket}/${var.tag}/libnext.so /usr/local/lib/libnext.so
    sudo ldconfig
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = var.google_machine_type
  project            = var.google_project
  region             = var.google_region
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-udp-all"]
  target_size        = 8
}

# ----------------------------------------------------------------------------------------

module "raspberry_client" {

  source = "../../modules/external_mig_without_health_check"

  service_name = "raspberry-client"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a raspberry_client.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    NEXT_LOG_LEVEL=4
    NEXT_CUSTOMER_PUBLIC_KEY=${var.customer_public_key}
    RASPBERRY_BACKEND_URL="http://${module.raspberry_backend.address}"
    RASPBERRY_NUM_CLIENTS=64
    EOF
    sudo gsutil cp ${var.google_artifacts_bucket}/${var.tag}/libnext.so /usr/local/lib/libnext.so
    sudo ldconfig
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = var.google_machine_type
  project            = var.google_project
  region             = var.google_region
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh"]
  target_size        = 16
}

# ----------------------------------------------------------------------------------------

module "ip2location" {

  source = "../../modules/external_mig_without_health_check"

  service_name = "ip2location"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a ip2location.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    MAXMIND_LICENSE_KEY=${var.maxmind_license_key}
    IP2LOCATION_BUCKET_NAME=${var.ip2location_bucket_name}
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = var.google_machine_type
  project            = var.google_project
  region             = var.google_region
  default_network    = google_compute_network.development.id
  default_subnetwork = google_compute_subnetwork.development.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-udp-all"]
  target_size        = 1
}

# ----------------------------------------------------------------------------------------
