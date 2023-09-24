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
variable "google_zones" { type = list(string) }
variable "google_service_account" { type = string }
variable "google_artifacts_bucket" { type = string }
variable "google_database_bucket" { type = string }

variable "cloudflare_api_token" { type = string }
variable "cloudflare_zone_id" { type = string }
variable "cloudflare_domain" { type = string }

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
    prefix  = "staging"
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

resource "google_compute_managed_ssl_certificate" "api" {
  name = "api"
  managed {
    domains = ["api-staging.${var.cloudflare_domain}"]
  }
}

resource "google_compute_managed_ssl_certificate" "relay" {
  name = "relay"
  managed {
    domains = ["relay-staging.${var.cloudflare_domain}"]
  }
}

resource "google_compute_managed_ssl_certificate" "portal" {
  name = "portal"
  managed {
    domains = ["portal-staging.${var.cloudflare_domain}"]
  }
}

# ----------------------------------------------------------------------------------------

resource "cloudflare_record" "api_domain" {
  zone_id = var.cloudflare_zone_id
  name    = "api-staging"
  value   = module.api.address
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "server_backend_domain" {
  zone_id = var.cloudflare_zone_id
  name    = "server-staging"
  value   = module.server_backend.address
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "relay_backend_domain" {
  zone_id = var.cloudflare_zone_id
  name    = "relay-staging"
  value   = module.relay_gateway.address
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "portal_domain" {
  zone_id = var.cloudflare_zone_id
  name    = "portal-staging"
  value   = module.portal.address
  type    = "A"
  proxied = false
}

# ----------------------------------------------------------------------------------------

resource "google_compute_network" "staging" {
  name                    = "staging"
  project                 = var.google_project
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "staging" {
  name                     = "staging"
  project                  = var.google_project
  ip_cidr_range            = "10.0.0.0/16"
  region                   = var.google_region
  network                  = google_compute_network.staging.id
  private_ip_google_access = true
}

resource "google_compute_subnetwork" "internal_http_load_balancer" {
  name          = "internal-http-load-balancer"
  project       = var.google_project
  region        = var.google_region
  purpose       = "INTERNAL_HTTPS_LOAD_BALANCER"
  role          = "ACTIVE"
  network       = google_compute_network.staging.id
  ip_cidr_range = "10.1.0.0/16"
}

# ----------------------------------------------------------------------------------------

resource "google_compute_firewall" "allow_ssh" {
  name          = "allow-ssh"
  project       = var.google_project
  direction     = "INGRESS"
  network       = google_compute_network.staging.id
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
  network       = google_compute_network.staging.id
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
  network       = google_compute_network.staging.id
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
  network       = google_compute_network.staging.id
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
  network       = google_compute_network.staging.id
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
  network       = google_compute_network.staging.id
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
  network       = google_compute_network.staging.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
  }
  target_tags = ["allow-udp-all"]
}

# ----------------------------------------------------------------------------------------

resource "google_redis_instance" "redis_relay_backend" {
  name                    = "redis-relay-backend"
  tier                    = "STANDARD_HA"
  memory_size_gb          = 10
  region                  = "us-central1"
  redis_version           = "REDIS_7_0"
  redis_configs           = { "maxmemory-gb" = "5" }
  authorized_network      = google_compute_network.staging.id
}

resource "google_redis_instance" "redis_map_cruncher" {
  name                    = "redis-map-cruncher"
  tier                    = "STANDARD_HA"
  memory_size_gb          = 2
  region                  = "us-central1"
  redis_version           = "REDIS_7_0"
  redis_configs           = { "activedefrag" = "yes", "maxmemory-policy" = "allkeys-lru", "maxmemory-gb" = "1" }
  authorized_network      = google_compute_network.staging.id
}

resource "google_redis_instance" "redis_analytics" {
  name                    = "redis-analytics"
  tier                    = "STANDARD_HA"
  memory_size_gb          = 2
  region                  = "us-central1"
  redis_version           = "REDIS_7_0"
  redis_configs           = { "activedefrag" = "yes", "maxmemory-policy" = "allkeys-lru", "maxmemory-gb" = "1" }
  authorized_network      = google_compute_network.staging.id
}

locals {
  redis_portal_address = "10.0.0.207:6379"
  redis_server_backend_address = "10.0.0.207:6379"
}

output "redis_portal_address" {
  description = "The IP address of the portal redis instance"
  value       = local.redis_portal_instance
}

output "redis_relay_backend_address" {
  description = "The IP address of the relay backend redis instance"
  value       = google_redis_instance.redis_relay_backend.host
}

output "redis_server_backend_address" {
  description = "The IP address of the server backend redis instance"
  value       = google_redis_instance.redis_server_backend.host
}

output "redis_map_cruncher_address" {
  description = "The IP address of the map cruncher redis instance"
  value       = google_redis_instance.redis_map_cruncher.host
}

output "redis_analytics_address" {
  description = "The IP address of the analytics redis instance"
  value       = google_redis_instance.redis_analytics.host
}

# ----------------------------------------------------------------------------------------

locals {

  pubsub_channels = [
    "route_matrix_update",
    "relay_to_relay_ping",
    "relay_update",
    "server_init",
    "server_update",
    "near_relay_ping",
    "session_update",
    "session_summary",
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
        "description": "True if this slice is being accelerated over network next"
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
      },
      {
        "name": "fallback_to_direct",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the SDK has encountered a fatal error and cannot continue acceleration. Typically this only happens when the system is misconfigured or extremely overloaded."
      },
      {
        "name": "reported",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if this session was reported by the player"
      },
      {
        "name": "latency_reduction",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if this session took network next this slice to reduce latency"
      },
      {
        "name": "packet_loss_reduction",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if this session took network next this slice to reduce packet loss"
      },
      {
        "name": "force_next",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if this session took network next this slice because it was forced to"
      },
      {
        "name": "long_session_update",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the processing for this slice on the server backend took a long time. This may indicate that the server backend is overloaded."
      },
      {
        "name": "client_next_bandwidth_over_limit",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the client to server next bandwidth went over the envelope limit this slice and was sent over direct."
      },
      {
        "name": "server_next_bandwidth_over_limit",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the server to client next bandwidth went over the envelope limit this slice and was sent over direct."
      },
      {
        "name": "veto",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the routing logic decided that this session should no longer be accelerated for some reason."
      },
      {
        "name": "disabled",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the buyer is disabled. Disabled buyers don't perform any acceleration or analytics on network next."
      },
      {
        "name": "not_selected",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "If the route shader selection % is any value other than 100%, then this is true for sessions that were not selected for acceleration."
      },
      {
        "name": "a",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "This session was part of an AB test, and is in the A group."
      },
      {
        "name": "b",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "This session was part of an AB test, and is in the B group."
      },
      {
        "name": "latency_worse",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if we made latency worse."
      },
      {
        "name": "location_veto",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if we could not locate the player, eg. lat long is at null island (0,0)."
      },
      {
        "name": "mispredict",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if we significantly mispredicted the latency reduction we could provide for this session."
      },
      {
        "name": "lack_of_diversity",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if route diversity is set in the route shader, and we don't have enough route diversity to accelerate this session."
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
      },
      {
        "name": "error",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "Error flags to diagnose what's happening with a session. Look up SessionError_* in the codebase for a list of errors. 0 if no error has occurred."
      },
      {
        "name": "reported",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if this session was reported by the player"
      },
      {
        "name": "latency_reduction",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if this session took network next to reduce latency"
      },
      {
        "name": "packet_loss_reduction",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if this session took network next to reduce packet loss"
      },
      {
        "name": "force_next",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if this session took network next because it was forced to"
      },
      {
        "name": "long_session_update",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the processing for any slices in this session took a long time. This may indicate that the server backend is overloaded."
      },
      {
        "name": "client_next_bandwidth_over_limit",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the client to server next bandwidth went over the envelope limit at some point and was sent over direct."
      },
      {
        "name": "server_next_bandwidth_over_limit",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the server to client next bandwidth went over the envelope limit at some point and was sent over direct."
      },
      {
        "name": "veto",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the routing logic decided that this session should no longer be accelerated for some reason."
      },
      {
        "name": "disabled",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if the buyer is disabled. Disabled buyers don't perform any acceleration or analytics on network next."
      },
      {
        "name": "not_selected",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "If the route shader selection % is any value other than 100%, then this is true for sessions that were not selected for acceleration."
      },
      {
        "name": "a",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "This session was part of an AB test, and is in the A group."
      },
      {
        "name": "b",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "This session was part of an AB test, and is in the B group."
      },
      {
        "name": "latency_worse",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if we made latency worse."
      },
      {
        "name": "location_veto",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if we could not locate the player, eg. lat long is at null island (0,0)."
      },
      {
        "name": "mispredict",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if we significantly mispredicted the latency reduction we could provide for this session."
      },
      {
        "name": "lack_of_diversity",
        "type": "BOOL",
        "mode": "REQUIRED",
        "description": "True if route diversity is set in the route shader, and we don't have enough route diversity to accelerate this session."
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
        "name": "database_size",
        "type": "INT64",
        "mode": "REQUIRED",
        "description": "The size of the database.bin in bytes (it is included in the route matrix size)."
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

  }

  bigquery_table_clustering = {

    "session_update" = [ "session_id" ]
    "session_summary" = [ "session_id", "buyer_id", "user_hash" ]
    "server_update" = [ "datacenter_id", "buyer_id" ]
    "server_init" = [ "datacenter_id", "buyer_id" ]
    "relay_update" = [ "relay_id" ]
    "route_matrix_update" = []
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
  network       = google_compute_network.staging.id
}

resource "google_service_networking_connection" "postgres" {
  network                 = google_compute_network.staging.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.postgres_private_address.name]
}

resource "google_sql_database_instance" "postgres" {
  name = "postgres"
  database_version = "POSTGRES_14"
  region = "${var.google_region}"
  depends_on = [google_service_networking_connection.postgres]
  settings {
    tier = "db-custom-1-3840"
    ip_configuration {
      ipv4_enabled    = "false"
      private_network = google_compute_network.staging.id
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
  network              = google_compute_network.staging.name
  import_custom_routes = true
  export_custom_routes = true
}

output "postgres_address" {
  description = "The IP address of the postgres instance"
  value = "${google_compute_global_address.postgres_private_address.address}"
}

# ----------------------------------------------------------------------------------------

module "magic_backend" {

  source = "../modules/internal_http_service_autoscale"

  service_name = "magic-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a magic_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    ENABLE_PROFILER=1
    EOF
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "n1-highcpu-2"
  project                    = var.google_project
  region                     = var.google_region
  zones                      = var.google_zones
  default_network            = google_compute_network.staging.id
  default_subnetwork         = google_compute_subnetwork.staging.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-http"]
  min_size                   = 3
  max_size                   = 16
  target_cpu                 = 60
}

output "magic_backend_address" {
  description = "The IP address of the magic backend load balancer"
  value       = module.magic_backend.address
}

# ----------------------------------------------------------------------------------------

module "relay_gateway" {

  source = "../modules/external_http_service_autoscale"

  service_name = "relay-gateway"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh  -t ${var.tag} -b ${var.google_artifacts_bucket} -a relay_gateway.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    GOOGLE_PROJECT_ID=${var.google_project}
    REDIS_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.google_database_bucket}/staging.bin"
    DATABASE_PATH="/app/database.bin"
    RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
    RELAY_BACKEND_PRIVATE_KEY=${var.relay_backend_private_key}
    PING_KEY=${var.ping_key}
    ENABLE_PROFILER=1
    EOF
    sudo gsutil cp ${var.google_database_bucket}/staging.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = "c3-highcpu-4"
  project                  = var.google_project
  region                   = var.google_region
  zones                    = var.google_zones
  default_network          = google_compute_network.staging.id
  default_subnetwork       = google_compute_subnetwork.staging.id
  service_account          = var.google_service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http"]
  min_size                 = 3
  max_size                 = 64
  target_cpu               = 60
  domain                   = "relay-staging.${var.cloudflare_domain}"
  certificate              = google_compute_managed_ssl_certificate.relay.id
}

output "relay_gateway_address" {
  description = "The IP address of the relay gateway load balancer"
  value       = module.relay_gateway.address
}

# ----------------------------------------------------------------------------------------

module "relay_backend" {

  source = "../modules/internal_http_service"

  service_name = "relay-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a relay_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    GOOGLE_PROJECT_ID=${var.google_project}
    REDIS_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.google_database_bucket}/staging.bin"
    DATABASE_PATH="/app/database.bin"
    INITIAL_DELAY=15s
    ENABLE_GOOGLE_PUBSUB=true
    ENABLE_PROFILER=1
    EOF
    sudo gsutil cp ${var.google_database_bucket}/staging.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "c3-highcpu-4"
  project                    = var.google_project
  region                     = var.google_region
  zones                      = var.google_zones
  default_network            = google_compute_network.staging.id
  default_subnetwork         = google_compute_subnetwork.staging.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-http"]
  target_size                = 3

  depends_on = [google_pubsub_topic.pubsub_topic, google_pubsub_subscription.pubsub_subscription]
}

output "relay_backend_address" {
  description = "The IP address of the relay backend load balancer"
  value       = module.relay_backend.address
}

# ----------------------------------------------------------------------------------------

module "analytics" {

  source = "../modules/internal_http_service_autoscale"

  service_name = "analytics"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a analytics.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    GOOGLE_PROJECT_ID=${var.google_project}
    DATABASE_URL="${var.google_database_bucket}/staging.bin"
    DATABASE_PATH="/app/database.bin"
    COST_MATRIX_URL="http://${module.relay_backend.address}/cost_matrix"
    ROUTE_MATRIX_URL="http://${module.relay_backend.address}/route_matrix"
    REDIS_HOSTNAME="${google_redis_instance.redis_analytics.host}:6379"
    ENABLE_GOOGLE_PUBSUB=true
    ENABLE_GOOGLE_BIGQUERY=true
    ENABLE_PROFILER=1
    REPS=10
    EOF
    sudo gsutil cp ${var.google_database_bucket}/staging.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "n1-highcpu-2"
  project                    = var.google_project
  region                     = var.google_region
  zones                      = var.google_zones
  default_network            = google_compute_network.staging.id
  default_subnetwork         = google_compute_subnetwork.staging.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks"]
  min_size                   = 3
  max_size                   = 64
  target_cpu                 = 90

  depends_on = [google_pubsub_topic.pubsub_topic, google_pubsub_subscription.pubsub_subscription]
}

output "analytics_address" {
  description = "The IP address of the analytics load balancer"
  value       = module.analytics.address
}

# ----------------------------------------------------------------------------------------

module "api" {

  source = "../modules/external_http_service_autoscale"

  service_name = "api"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a api.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    REDIS_PORTAL_CLUSTER="${local.redis_portal_address}"
    REDIS_RELAY_BACKEND_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    REDIS_MAP_CRUNCHER_HOSTNAME="${google_redis_instance.redis_map_cruncher.host}:6379"
    SESSION_CRUNCHER_URL="http://${module.session_cruncher.address}"
    GOOGLE_PROJECT_ID=${var.google_project}
    DATABASE_URL="${var.google_database_bucket}/staging.bin"
    DATABASE_PATH="/app/database.bin"
    PGSQL_CONFIG="host=${google_sql_database_instance.postgres.ip_address.0.ip_address} port=5432 user=developer password=developer dbname=database sslmode=disable"
    API_PRIVATE_KEY=${var.api_private_key}
    ALLOWED_ORIGIN="*"
    ENABLE_PROFILER=1
    EOF
    sudo gsutil cp ${var.google_database_bucket}/staging.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = "n1-highcpu-2"
  project                  = var.google_project
  region                   = var.google_region
  zones                    = var.google_zones
  default_network          = google_compute_network.staging.id
  default_subnetwork       = google_compute_subnetwork.staging.id
  service_account          = var.google_service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http-vpn-only"]
  min_size                 = 3
  max_size                 = 16
  target_cpu               = 60
  domain                   = "api-staging.${var.cloudflare_domain}"
  certificate              = google_compute_managed_ssl_certificate.api.id
}

output "api_address" {
  description = "The IP address of the api load balancer"
  value       = module.api.address
}

// ---------------------------------------------------------------------------------------

module "session_cruncher" {

  source = "../modules/internal_http_service"

  service_name = "session-cruncher"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a session_cruncher.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    DEBUG_LOGS=1
    EOF
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "c3-highcpu-44"
  project                    = var.google_project
  region                     = var.google_region
  zones                      = var.google_zones
  default_network            = google_compute_network.staging.id
  default_subnetwork         = google_compute_subnetwork.staging.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = var.google_service_account
  tags                       = ["allow-ssh", "allow-http"]
  target_size                = 1
}

// ---------------------------------------------------------------------------------------

module "portal_cruncher" {

  source = "../modules/internal_mig_with_health_check_autoscale"

  service_name = "portal-cruncher"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a portal_cruncher.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    REDIS_PORTAL_CLUSTER="${local.redis_portal_address}"
    REDIS_SERVER_BACKEND_CLUSTER="${local.redis_server_backend_address}"
    REDIS_RELAY_BACKEND_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    SESSION_CRUNCHER_URL="http://${module.session_cruncher.address}"
    IP2LOCATION_BUCKET_NAME=${var.ip2location_bucket_name}
    ENABLE_PROFILER=1
    REPS=10
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "n1-highcpu-2"
  project            = var.google_project
  region             = var.google_region
  zones              = var.google_zones
  default_network    = google_compute_network.staging.id
  default_subnetwork = google_compute_subnetwork.staging.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-http"]
  min_size           = 3
  max_size           = 64
  target_cpu         = 60
}

// ---------------------------------------------------------------------------------------

module "map_cruncher" {

  source = "../modules/internal_mig_with_health_check"

  service_name = "map-cruncher"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a map_cruncher.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    REPS=64
    REDIS_HOSTNAME="${google_redis_instance.redis_map_cruncher.host}:6379"
    REDIS_SERVER_BACKEND_HOSTNAME="${google_redis_instance.redis_server_backend.host}:6379"
    ENABLE_PROFILER=1
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "c3-highcpu-4"
  project            = var.google_project
  region             = var.google_region
  zones              = var.google_zones
  default_network    = google_compute_network.staging.id
  default_subnetwork = google_compute_subnetwork.staging.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-http"]
}

# ----------------------------------------------------------------------------------------

module "server_backend" {

  source = "../modules/external_udp_service_autoscale"

  service_name = "server-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a server_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    UDP_PORT=40000
    UDP_BIND_ADDRESS="##########:40000"
    UDP_NUM_THREADS=64
    GOOGLE_PROJECT_ID=${var.google_project}
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    REDIS_CLUSTER="${local.redis_server_backend_address}"
    RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
    RELAY_BACKEND_PRIVATE_KEY=${var.relay_backend_private_key}
    SERVER_BACKEND_ADDRESS="##########:40000"
    SERVER_BACKEND_PUBLIC_KEY=${var.server_backend_public_key}
    SERVER_BACKEND_PRIVATE_KEY=${var.server_backend_private_key}
    ROUTE_MATRIX_URL="http://${module.relay_backend.address}/route_matrix"
    PING_KEY=${var.ping_key}
    IP2LOCATION_BUCKET_NAME=${var.ip2location_bucket_name}
    ENABLE_GOOGLE_PUBSUB=true
    ENABLE_PROFILER=1
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "c3-highcpu-8"
  project            = var.google_project
  region             = var.google_region
  zones              = var.google_zones
  port               = 40000
  default_network    = google_compute_network.staging.id
  default_subnetwork = google_compute_subnetwork.staging.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-health-checks", "allow-udp-40000"]
  min_size           = 3
  max_size           = 64
  target_cpu         = 40

  depends_on = [google_pubsub_topic.pubsub_topic, google_pubsub_subscription.pubsub_subscription]
}

output "server_backend_address" {
  description = "The IP address of the server backend load balancer"
  value       = module.server_backend.address
}

# ----------------------------------------------------------------------------------------

module "ip2location" {

  source = "../modules/external_mig_without_health_check"

  service_name = "ip2location"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a ip2location.tar.gz
    cat <<EOF > /app/app.env
    ENV=staging
    MAXMIND_LICENSE_KEY=${var.maxmind_license_key}
    IP2LOCATION_BUCKET_NAME=${var.ip2location_bucket_name}
    ENABLE_PROFILER=1
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "n1-highcpu-2"
  project            = var.google_project
  region             = var.google_region
  zones              = var.google_zones
  default_network    = google_compute_network.staging.id
  default_subnetwork = google_compute_subnetwork.staging.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-udp-all"]
  target_size        = 1
}

# ----------------------------------------------------------------------------------------

module "load_test_relays" {

  source = "../modules/external_mig_without_health_check"

  service_name = "load-test-relays"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a load_test_relays.tar.gz
    cat <<EOF > /app/app.env
    RELAY_BACKEND_HOSTNAME=http://${module.relay_gateway.address}
    RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
    RELAY_PRIVATE_KEY=lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=
    ENABLE_PROFILER=1
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "n1-highcpu-2"
  project            = var.google_project
  region             = var.google_region
  zones              = var.google_zones
  default_network    = google_compute_network.staging.id
  default_subnetwork = google_compute_subnetwork.staging.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-udp-all"]
  target_size        = 1
}

# ----------------------------------------------------------------------------------------

module "load_test_servers" {

  source = "../modules/external_mig_without_health_check"

  service_name = "load-test-servers"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a load_test_servers.tar.gz
    cat <<EOF > /app/app.env
    NUM_SERVERS=50000
    SERVER_BACKEND_ADDRESS=${module.server_backend.address}:40000
    NEXT_CUSTOMER_PRIVATE_KEY=leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn
    ENABLE_PROFILER=1
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "n1-highcpu-2"
  project            = var.google_project
  region             = var.google_region
  zones              = var.google_zones
  default_network    = google_compute_network.staging.id
  default_subnetwork = google_compute_subnetwork.staging.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-udp-all"]
  target_size        = 2
}

# ----------------------------------------------------------------------------------------

module "load_test_sessions" {

  source = "../modules/external_mig_without_health_check"

  service_name = "load-test-sessions"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a load_test_sessions.tar.gz
    cat <<EOF > /app/app.env
    NUM_SESSIONS=50000
    SERVER_BACKEND_ADDRESS=${module.server_backend.address}:40000
    NEXT_CUSTOMER_PRIVATE_KEY=leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn
    ENABLE_PROFILER=1
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "n1-highcpu-8"
  project            = var.google_project
  region             = var.google_region
  zones              = var.google_zones
  default_network    = google_compute_network.staging.id
  default_subnetwork = google_compute_subnetwork.staging.id
  service_account    = var.google_service_account
  tags               = ["allow-ssh", "allow-udp-all"]
  target_size        = 2
}

# ----------------------------------------------------------------------------------------

module "portal" {

  source = "../modules/nginx"      # todo: this should become autoscale

  service_name = "portal"

  artifact                 = "${var.google_artifacts_bucket}/${var.tag}/portal.tar.gz"
  config                   = "${var.google_artifacts_bucket}/${var.tag}/nginx.conf"
  tag                      = var.tag
  extra                    = var.extra
  machine_type             = "n1-highcpu-2"
  project                  = var.google_project
  region                   = var.google_region
  zones                    = var.google_zones
  default_network          = google_compute_network.staging.id
  default_subnetwork       = google_compute_subnetwork.staging.id
  service_account          = var.google_service_account
  tags                     = ["allow-ssh", "allow-http", "allow-https"]
  domain                   = "portal-staging.${var.cloudflare_domain}"
  certificate              = google_compute_managed_ssl_certificate.portal.id
}

output "portal_address" {
  description = "The IP address of the portal load balancer"
  value       = module.portal.address
}

# ----------------------------------------------------------------------------------------

resource "google_compute_router" "router" {
  name    = "router-to-internet"
  network = google_compute_network.staging.id
  project = var.google_project
  region  = var.google_region
}

resource "google_compute_router_nat" "nat" {
  name                               = "nat"
  router                             = google_compute_router.router.name
  region                             = var.google_region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

# ----------------------------------------------------------------------------------------
