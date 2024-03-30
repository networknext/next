# ----------------------------------------------------------------------------------------

variable "tag" { type = string }

variable "extra" { 
  type = string 
  default = ""
}

variable "vpn_address" { type = string }

variable "google_credentials" { type = string }
variable "google_location" { type = string }
variable "google_region" { type = string }
variable "google_zone" { type = string }
variable "google_zones" { type = list(string) }
variable "google_artifacts_bucket" { type = string }
variable "google_database_bucket" { type = string }

variable "cloudflare_api_token" { type = string }
variable "cloudflare_zone_id" { type = string }
variable "cloudflare_domain" { type = string }

variable "relay_backend_public_key" { type = string }

variable "server_backend_public_key" { type = string }

variable "test_buyer_public_key" { type = string }
variable "test_buyer_private_key" { type = string }

variable "raspberry_buyer_public_key" { type = string }
variable "raspberry_buyer_private_key" { type = string }

variable "ip2location_bucket_name" { type = string }

variable "test_server_tag" { type = string }

locals {
  google_project_id          = file("../../projects/prod-project-id.txt")
  google_project_number      = file("../../projects/prod-project-number.txt")
  google_service_account     = file("../../projects/prod-runtime-service-account.txt")
  maxmind_license_key        = file("~/secrets/maxmind.txt")
  relay_backend_private_key  = file("~/secrets/prod-relay-backend-private-key.txt")
  server_backend_private_key = file("~/secrets/prod-server-backend-private-key.txt")
  api_private_key            = file("~/secrets/prod-api-private-key.txt")
  ping_key                   = file("~/secrets/prod-ping-key.txt")
}

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0.0"
    }
    google-beta = {
      source = "hashicorp/google-beta"
      version = "~> 5.0.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
  backend "gcs" {
    bucket  = "xdp_network_next_terraform"
    prefix  = "production"
  }
}

provider "google" {
  credentials = file(var.google_credentials)
  project     = local.google_project_id
  region      = var.google_region
  zone        = var.google_zone
}

provider "google-beta" {
  credentials = file(var.google_credentials)
  project     = local.google_project_id
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
    domains = ["api.${var.cloudflare_domain}"]
  }
}

resource "google_compute_managed_ssl_certificate" "relay" {
  name = "relay"
  managed {
    domains = ["relay.${var.cloudflare_domain}"]
  }
}

resource "google_compute_managed_ssl_certificate" "portal" {
  name = "portal"
  managed {
    domains = ["portal.${var.cloudflare_domain}"]
  }
}

resource "google_compute_managed_ssl_certificate" "raspberry" {
  name = "raspberry"
  managed {
    domains = ["raspberry.${var.cloudflare_domain}"]
  }
}

# ----------------------------------------------------------------------------------------

resource "google_compute_network" "production" {
  name                    = "production"
  project                 = local.google_project_id
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "production" {
  name                     = "production"
  project                  = local.google_project_id
  ip_cidr_range            = "10.0.0.0/16"
  region                   = var.google_region
  network                  = google_compute_network.production.id
  private_ip_google_access = true
}

resource "google_compute_subnetwork" "internal_http_load_balancer" {
  name          = "internal-http-load-balancer"
  project       = local.google_project_id
  region        = var.google_region
  purpose       = "INTERNAL_HTTPS_LOAD_BALANCER"
  role          = "ACTIVE"
  network       = google_compute_network.production.id
  ip_cidr_range = "10.1.0.0/16"
}

# ----------------------------------------------------------------------------------------

resource "google_compute_firewall" "allow_ssh" {
  name          = "allow-ssh"
  project       = local.google_project_id
  direction     = "INGRESS"
  network       = google_compute_network.production.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "35.235.240.0/20"]
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
  target_tags = ["allow-ssh"]
}

resource "google_compute_firewall" "allow_health_checks" {
  name          = "allow-health-checks"
  project       = local.google_project_id
  direction     = "INGRESS"
  network       = google_compute_network.production.id
  source_ranges = ["35.191.0.0/16", "130.211.0.0/22"]

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  target_tags = ["allow-health-checks"]
}

resource "google_compute_firewall" "allow_network_load_balancer_traffic" {
  name          = "allow-network-load-balancer-traffic"
  project       = local.google_project_id
  direction     = "INGRESS"
  network       = google_compute_network.production.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "35.235.240.0/20", "209.85.152.0/22", "209.85.204.0/22"]

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  target_tags = ["allow-health-checks"]
}

resource "google_compute_firewall" "allow_http" {
  name          = "allow-http"
  project       = local.google_project_id
  direction     = "INGRESS"
  network       = google_compute_network.production.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
  target_tags = ["allow-http"]
}

resource "google_compute_firewall" "allow_redis" {
  name          = "allow-redis"
  project       = local.google_project_id
  direction     = "INGRESS"
  network       = google_compute_network.production.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "tcp"
    ports    = ["6379"]
  }
  target_tags = ["allow-redis"]
}

resource "google_compute_firewall" "allow_udp_40000" {
  name          = "allow-udp-40000"
  project       = local.google_project_id
  direction     = "INGRESS"
  network       = google_compute_network.production.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
    ports    = ["40000"]
  }
  target_tags = ["allow-udp-40000"]
}

resource "google_compute_firewall" "allow_udp_all" {
  name          = "allow-udp-all"
  project       = local.google_project_id
  direction     = "INGRESS"
  network       = google_compute_network.production.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
  }
  target_tags = ["allow-udp-all"]
}

# ----------------------------------------------------------------------------------------

resource "cloudflare_record" "api_domain" {
  zone_id = var.cloudflare_zone_id
  name    = "api"
  value   = module.api.address
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "server_backend_domain" {
  zone_id = var.cloudflare_zone_id
  name    = "server"
  value   = module.server_backend.address
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "relay_backend_domain" {
  zone_id = var.cloudflare_zone_id
  name    = "relay"
  value   = module.relay_gateway.address
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "portal_domain" {
  zone_id = var.cloudflare_zone_id
  name    = "portal"
  value   = module.portal.address
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "raspberry_domain" {
  zone_id = var.cloudflare_zone_id
  name    = "raspberry"
  value   = module.raspberry_backend.address
  type    = "A"
  proxied = false
}

# ----------------------------------------------------------------------------------------

module "redis_time_series" {

  source = "../../modules/redis_stack"

  service_name = "redis-time-series"

  machine_type             = "n1-standard-2"
  project                  = local.google_project_id
  region                   = var.google_region
  zone                     = var.google_zone
  default_network          = google_compute_network.production.id
  default_subnetwork       = google_compute_subnetwork.production.id
  service_account          = local.google_service_account
  tags                     = ["allow-redis", "allow-ssh"]
}

output "redis_time_series_address" {
  description = "The IP address of the redis time series database"
  value       = module.redis_time_series.address
}

# ----------------------------------------------------------------------------------------

resource "google_redis_cluster" "portal" {
  provider       = google-beta
  name           = "portal"
  shard_count    = 10
  psc_configs {
    network = google_compute_network.production.id
  }
  region = "us-central1"
  replica_count = 1
  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_DISABLED"
  authorization_mode = "AUTH_MODE_DISABLED"
  depends_on = [
    google_network_connectivity_service_connection_policy.default
  ]
}

resource "google_network_connectivity_service_connection_policy" "default" {
  provider = google-beta
  name = "redis"
  location = "us-central1"
  service_class = "gcp-memorystore-redis"
  description   = "redis cluster service connection policy"
  network = google_compute_network.production.id
  psc_config {
    subnetworks = [google_compute_subnetwork.production.id]
  }
}

locals {
  redis_portal_address = "${google_redis_cluster.portal.discovery_endpoints[0].address}:6379"
}

resource "google_redis_instance" "redis_relay_backend" {
  name                    = "redis-relay-backend"
  tier                    = "STANDARD_HA"
  memory_size_gb          = 10
  region                  = "us-central1"
  redis_version           = "REDIS_7_0"
  redis_configs           = { "maxmemory-gb" = "5" }
  authorized_network      = google_compute_network.production.id
}

resource "google_redis_instance" "redis_raspberry" {
  name               = "redis-raspberry"
  tier               = "STANDARD_HA"
  memory_size_gb     = 2
  region             = "us-central1"
  redis_version      = "REDIS_7_0"
  redis_configs      = { "activedefrag" = "yes", "maxmemory-policy" = "allkeys-lru" }
  authorized_network = google_compute_network.production.id
}

output "redis_portal_address" {
  description = "The IP address of the portal redis instance"
  value       = local.redis_portal_address
}

output "redis_relay_backend_address" {
  description = "The IP address of the relay backend redis instance"
  value       = google_redis_instance.redis_relay_backend.host
}

output "redis_raspberry_address" {
  description = "The IP address of the raspberry redis instance"
  value       = google_redis_instance.redis_raspberry.host
}

# ----------------------------------------------------------------------------------------

locals {

  pubsub_channels = [
    "route_matrix_update",
    "relay_to_relay_ping",
    "relay_update",
    "server_init",
    "server_update",
    "client_relay_ping",
    "server_relay_ping",
    "session_update",
    "session_summary",
  ]
  
}

resource "google_pubsub_schema" "pubsub_schema" {
  count      = length(local.pubsub_channels)
  name       = local.pubsub_channels[count.index]
  type       = "AVRO"
  definition = file("../../../schemas/pubsub/${local.pubsub_channels[count.index]}.json")
}

resource "google_pubsub_topic" "pubsub_topic" {
  count      = length(local.pubsub_channels)
  name       = local.pubsub_channels[count.index]
  schema_settings {
    schema = google_pubsub_schema.pubsub_schema[count.index].id
    encoding = "BINARY"
  }
  depends_on = [google_pubsub_schema.pubsub_schema]
} 

resource "google_project_iam_member" "pubsub_bigquery_admin" {
  project    = local.google_project_id
  role       = "roles/bigquery.admin"
  member     = "serviceAccount:service-${local.google_project_number}@gcp-sa-pubsub.iam.gserviceaccount.com"
  depends_on = [google_pubsub_topic.pubsub_topic]
}

resource "google_pubsub_subscription" "pubsub_subscription" {
  count                       = length(local.pubsub_channels)
  name                        = local.pubsub_channels[count.index]
  topic                       = google_pubsub_topic.pubsub_topic[count.index].name
  message_retention_duration  = "604800s"
  retain_acked_messages       = false
  ack_deadline_seconds        = 60
  expiration_policy {
    ttl = ""
  }
  bigquery_config {
    table = "${google_bigquery_table.table[local.pubsub_channels[count.index]].project}.${google_bigquery_table.table[local.pubsub_channels[count.index]].dataset_id}.${google_bigquery_table.table[local.pubsub_channels[count.index]].table_id}"
    use_topic_schema    = true
    drop_unknown_fields = true    
  }
  depends_on = [google_project_iam_member.pubsub_bigquery_admin]
}

# ----------------------------------------------------------------------------------------

locals {
  
  bigquery_tables = {
    "session_update"      = file("../../../schemas/bigquery/session_update.json")
    "session_summary"     = file("../../../schemas/bigquery/session_summary.json")
    "server_init"         = file("../../../schemas/bigquery/server_init.json")
    "server_update"       = file("../../../schemas/bigquery/server_update.json")
    "relay_update"        = file("../../../schemas/bigquery/relay_update.json")
    "route_matrix_update" = file("../../../schemas/bigquery/route_matrix_update.json")
    "relay_to_relay_ping" = file("../../../schemas/bigquery/relay_to_relay_ping.json")
    "client_relay_ping"     = file("../../../schemas/bigquery/client_relay_ping.json")
    "server_relay_ping"     = file("../../../schemas/bigquery/server_relay_ping.json")
  }

  bigquery_table_clustering = {
    "session_update"      = [ "session_id" ]
    "session_summary"     = [ "buyer_id", "user_hash" ]
    "server_update"       = [ "datacenter_id", "buyer_id" ]
    "server_init"         = [ "datacenter_id", "buyer_id" ]
    "relay_update"        = [ "relay_id" ]
    "route_matrix_update" = []
    "relay_to_relay_ping" = [ "source_relay_id" ]
    "client_relay_ping"   = [ "client_relay_id", "user_hash" ]
    "server_relay_ping"   = [ "server_relay_id" ]
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
    field = "timestamp"
  }
}

# ----------------------------------------------------------------------------------------

resource "google_compute_global_address" "postgres_private_address" {
  name          = "postgres-private-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.production.id
}

resource "google_service_networking_connection" "postgres" {
  network                 = google_compute_network.production.id
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
      private_network = google_compute_network.production.id
    }
    database_flags {
      name  = "max_connections"
      value = "1024"
    }
    backup_configuration {
      enabled = true
    }
    deletion_protection_enabled = true
  }
  deletion_protection = true
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
  network              = google_compute_network.production.name
  import_custom_routes = true
  export_custom_routes = true
}

output "postgres_address" {
  description = "The IP address of the postgres instance"
  value = "${google_compute_global_address.postgres_private_address.address}"
}

# ----------------------------------------------------------------------------------------

module "magic_backend" {

  source = "../../modules/internal_http_service_autoscale"

  service_name = "magic-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a magic_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=prod
    EOF
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "n1-highcpu-2"
  project                    = local.google_project_id
  region                     = var.google_region
  zones                      = var.google_zones
  default_network            = google_compute_network.production.id
  default_subnetwork         = google_compute_subnetwork.production.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = local.google_service_account
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

  source = "../../modules/external_http_service_autoscale"

  service_name = "relay-gateway"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh  -t ${var.tag} -b ${var.google_artifacts_bucket} -a relay_gateway.tar.gz
    cat <<EOF > /app/app.env
    ENV=prod
    GOOGLE_PROJECT_ID=${local.google_project_id}
    REDIS_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.google_database_bucket}/prod.bin"
    DATABASE_PATH="/app/database.bin"
    RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
    RELAY_BACKEND_PRIVATE_KEY=${local.relay_backend_private_key}
    PING_KEY=${local.ping_key}
    RELAY_BACKEND_ADDRESS=""
    EOF
    sudo gsutil cp ${var.google_database_bucket}/prod.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = "c3-highcpu-4"
  project                  = local.google_project_id
  region                   = var.google_region
  zones                    = var.google_zones
  default_network          = google_compute_network.production.id
  default_subnetwork       = google_compute_subnetwork.production.id
  service_account          = local.google_service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http"]
  min_size                 = 3
  max_size                 = 64
  target_cpu               = 60
  domain                   = "relay.${var.cloudflare_domain}"
  certificate              = google_compute_managed_ssl_certificate.relay.id
  
  depends_on = [
    google_redis_instance.redis_relay_backend
  ]
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
    ENV=prod
    ENABLE_RELAY_HISTORY=true
    GOOGLE_PROJECT_ID=${local.google_project_id}
    REDIS_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.google_database_bucket}/prod.bin"
    DATABASE_PATH="/app/database.bin"
    INITIAL_DELAY=360s
    MAX_JITTER=2
    MAX_PACKET_LOSS=0.1
    ENABLE_GOOGLE_PUBSUB=true
    ENABLE_REDIS_TIME_SERIES=true
    REDIS_TIME_SERIES_HOSTNAME="${module.redis_time_series.address}:6379"
    REDIS_PORTAL_CLUSTER="${local.redis_portal_address}"
    RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
    RELAY_BACKEND_PRIVATE_KEY=${local.relay_backend_private_key}
    EOF
    sudo gsutil cp ${var.google_database_bucket}/prod.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "c3-highcpu-44"
  project                    = local.google_project_id
  region                     = var.google_region
  zones                      = var.google_zones
  default_network            = google_compute_network.production.id
  default_subnetwork         = google_compute_subnetwork.production.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = local.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-http"]
  initial_delay              = 420
  target_size                = 3
  tier_1                     = true

  depends_on = [
    google_pubsub_topic.pubsub_topic, 
    google_pubsub_subscription.pubsub_subscription,
    google_redis_instance.redis_relay_backend,
    module.redis_time_series
  ]
}

output "relay_backend_address" {
  description = "The IP address of the relay backend load balancer"
  value       = module.relay_backend.address
}

# ----------------------------------------------------------------------------------------

module "api" {

  source = "../../modules/external_http_service_autoscale"

  service_name = "api"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a api.tar.gz
    cat <<EOF > /app/app.env
    ENV=prod
    ENABLE_REDIS_TIME_SERIES=true
    REDIS_TIME_SERIES_HOSTNAME="${module.redis_time_series.address}:6379"
    REDIS_PORTAL_CLUSTER="${local.redis_portal_address}"
    REDIS_RELAY_BACKEND_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    SESSION_CRUNCHER_URL="http://${module.session_cruncher.address}"
    SERVER_CRUNCHER_URL="http://${module.server_cruncher.address}"
    GOOGLE_PROJECT_ID=${local.google_project_id}
    DATABASE_URL="${var.google_database_bucket}/prod.bin"
    DATABASE_PATH="/app/database.bin"
    PGSQL_CONFIG="host=${google_sql_database_instance.postgres.ip_address.0.ip_address} port=5432 user=developer password=developer dbname=database sslmode=disable"
    API_PRIVATE_KEY=${local.api_private_key}
    ALLOWED_ORIGIN="*"
    EOF
    sudo gsutil cp ${var.google_database_bucket}/prod.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = "n1-highcpu-2"
  project                  = local.google_project_id
  region                   = var.google_region
  zones                    = var.google_zones
  default_network          = google_compute_network.production.id
  default_subnetwork       = google_compute_subnetwork.production.id
  service_account          = local.google_service_account
  tags                     = ["allow-ssh", "allow-health-checks", "allow-http"]
  min_size                 = 3
  max_size                 = 16
  target_cpu               = 60
  domain                   = "api.${var.cloudflare_domain}"
  certificate              = google_compute_managed_ssl_certificate.api.id
}

output "api_address" {
  description = "The IP address of the api load balancer"
  value       = module.api.address
}

// ---------------------------------------------------------------------------------------

module "session_cruncher" {

  source = "../../modules/internal_http_service"

  service_name = "session-cruncher"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a session_cruncher.tar.gz
    cat <<EOF > /app/app.env
    ENV=prod
    ENABLE_REDIS_TIME_SERIES=true
    REDIS_TIME_SERIES_HOSTNAME="${module.redis_time_series.address}:6379"
    GOOGLE_PROJECT_ID=${local.google_project_id}
    DATABASE_URL="${var.google_database_bucket}/prod.bin"
    DATABASE_PATH="/app/database.bin"
    EOF
    sudo gsutil cp ${var.google_database_bucket}/prod.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "c3-highmem-4"
  project                    = local.google_project_id
  region                     = var.google_region
  zones                      = var.google_zones
  default_network            = google_compute_network.production.id
  default_subnetwork         = google_compute_subnetwork.production.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = local.google_service_account
  tags                       = ["allow-ssh", "allow-http"]
  target_size                = 1
}

// ---------------------------------------------------------------------------------------

module "server_cruncher" {

  source = "../../modules/internal_http_service"

  service_name = "server-cruncher"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a server_cruncher.tar.gz
    cat <<EOF > /app/app.env
    ENV=prod
    EOF
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "c3-highmem-4"
  project                    = local.google_project_id
  region                     = var.google_region
  zones                      = var.google_zones
  default_network            = google_compute_network.production.id
  default_subnetwork         = google_compute_subnetwork.production.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = local.google_service_account
  tags                       = ["allow-ssh", "allow-http"]
  target_size                = 1
}

# ----------------------------------------------------------------------------------------

module "server_backend" {

  source = "../../modules/external_udp_service_autoscale"

  service_name = "server-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a server_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=prod
    UDP_PORT=40000
    UDP_BIND_ADDRESS="##########:40000"
    UDP_NUM_THREADS=8
    UDP_SOCKET_READ_BUFFER=104857600
    UDP_SOCKET_WRITE_BUFFER=104857600
    GOOGLE_PROJECT_ID=${local.google_project_id}
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    REDIS_CLUSTER="${local.redis_portal_address}"
    RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
    RELAY_BACKEND_PRIVATE_KEY=${local.relay_backend_private_key}
    SERVER_BACKEND_ADDRESS="##########:40000"
    SERVER_BACKEND_PUBLIC_KEY=${var.server_backend_public_key}
    SERVER_BACKEND_PRIVATE_KEY=${local.server_backend_private_key}
    ROUTE_MATRIX_URL="http://${module.relay_backend.address}/route_matrix"
    PING_KEY=${local.ping_key}
    IP2LOCATION_BUCKET_NAME=${var.ip2location_bucket_name}
    ENABLE_GOOGLE_PUBSUB=true
    ENABLE_REDIS_TIME_SERIES=true
    REDIS_TIME_SERIES_HOSTNAME="${module.redis_time_series.address}:6379"
    REDIS_PORTAL_CLUSTER="${local.redis_portal_address}"
    REDIS_RELAY_BACKEND_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    SESSION_CRUNCHER_URL="http://${module.session_cruncher.address}"
    SERVER_CRUNCHER_URL="http://${module.server_cruncher.address}"
    PORTAL_NEXT_SESSIONS_ONLY=false
    ENABLE_IP2LOCATION=true
    EOF
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "c3-highcpu-44"
  project                    = local.google_project_id
  region                     = var.google_region
  zones                      = var.google_zones
  port                       = 40000
  default_network            = google_compute_network.production.id
  default_subnetwork         = google_compute_subnetwork.production.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = local.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-udp-40000"]
  min_size                   = 3
  max_size                   = 64
  target_cpu                 = 30

  depends_on = [
    google_pubsub_topic.pubsub_topic, 
    google_pubsub_subscription.pubsub_subscription,
    google_redis_cluster.portal
  ]
}

output "server_backend_address" {
  description = "The IP address of the server backend load balancer"
  value       = module.server_backend.address
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
    ENV=prod
    MAXMIND_LICENSE_KEY=${local.maxmind_license_key}
    IP2LOCATION_BUCKET_NAME=${var.ip2location_bucket_name}
    EOF
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "n1-highcpu-2"
  project            = local.google_project_id
  region             = var.google_region
  zones              = var.google_zones
  default_network    = google_compute_network.production.id
  default_subnetwork = google_compute_subnetwork.production.id
  service_account    = local.google_service_account
  tags               = ["allow-ssh", "allow-udp-all"]
  target_size        = 1
}

# ----------------------------------------------------------------------------------------

module "portal" {

  source = "../../modules/nginx"

  service_name = "portal"

  artifact                 = "${var.google_artifacts_bucket}/${var.tag}/portal.tar.gz"
  config                   = "${var.google_artifacts_bucket}/${var.tag}/nginx.conf"
  tag                      = var.tag
  extra                    = var.extra
  machine_type             = "n1-highcpu-2"
  project                  = local.google_project_id
  region                   = var.google_region
  zones                    = var.google_zones
  default_network          = google_compute_network.production.id
  default_subnetwork       = google_compute_subnetwork.production.id
  service_account          = local.google_service_account
  tags                     = ["allow-ssh", "allow-http", "allow-https"]
  domain                   = "portal.${var.cloudflare_domain}"
  certificate              = google_compute_managed_ssl_certificate.portal.id
  target_size              = 3
}

output "portal_address" {
  description = "The IP address of the portal load balancer"
  value       = module.portal.address
}

# ----------------------------------------------------------------------------------------

resource "google_compute_router" "router" {
  name    = "router-to-internet" 
  network = google_compute_network.production.id
  project = local.google_project_id
  region  = var.google_region
}

resource "google_compute_router_nat" "nat" {
  name                               = "nat"
  router                             = google_compute_router.router.name
  region                             = var.google_region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
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
    ENV=prod
    REDIS_HOSTNAME="${google_redis_instance.redis_raspberry.host}:6379"
    EOF
    sudo systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = "n1-standard-2"
  project                  = local.google_project_id
  region                   = var.google_region
  zones                    = var.google_zones
  default_network          = google_compute_network.production.id
  default_subnetwork       = google_compute_subnetwork.production.id
  service_account          = local.google_service_account
  tags                     = ["allow-ssh", "allow-http", "allow-https"]
  domain                   = "raspberry.${var.cloudflare_domain}"
  certificate              = google_compute_managed_ssl_certificate.raspberry.id
  target_size              = 3
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
    ENV=prod
    NEXT_LOG_LEVEL=1
    NEXT_DATACENTER=cloud
    NEXT_BUYER_PRIVATE_KEY=${var.raspberry_buyer_private_key}
    NEXT_SERVER_BACKEND_HOSTNAME="server.${var.cloudflare_domain}"
    NEXT_SERVER_BACKEND_PUBLIC_KEY="${var.server_backend_public_key}"
    NEXT_RELAY_BACKEND_PUBLIC_KEY="${var.relay_backend_public_key}"
    RASPBERRY_BACKEND_URL="https://raspberry.${var.cloudflare_domain}"
    EOF
    sudo gsutil cp ${var.google_artifacts_bucket}/${var.tag}/libnext.so /usr/local/lib/libnext.so
    sudo ldconfig
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "n1-standard-2"
  project            = local.google_project_id
  region             = var.google_region
  zones              = var.google_zones
  default_network    = google_compute_network.production.id
  default_subnetwork = google_compute_subnetwork.production.id
  service_account    = local.google_service_account
  tags               = ["allow-ssh", "allow-udp-all"]
  target_size        = 320
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
    ENV=prod
    NEXT_LOG_LEVEL=1
    NEXT_BUYER_PUBLIC_KEY=${var.raspberry_buyer_public_key}
    RASPBERRY_BACKEND_URL="https://raspberry.${var.cloudflare_domain}"
    RASPBERRY_NUM_CLIENTS=256
    NEXT_SERVER_BACKEND_PUBLIC_KEY="${var.server_backend_public_key}"
    NEXT_RELAY_BACKEND_PUBLIC_KEY="${var.relay_backend_public_key}"
    EOF
    sudo gsutil cp ${var.google_artifacts_bucket}/${var.tag}/libnext.so /usr/local/lib/libnext.so
    sudo ldconfig
    sudo systemctl start app.service
  EOF1

  tag                = var.tag
  extra              = var.extra
  machine_type       = "n1-standard-2"
  project            = local.google_project_id
  region             = var.google_region
  zones              = var.google_zones
  default_network    = google_compute_network.production.id
  default_subnetwork = google_compute_subnetwork.production.id
  service_account    = local.google_service_account
  tags               = ["allow-ssh"]
  target_size        = 400
}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "test_server_address" {
  name = "test-server-address"
}

resource "google_compute_instance" "test_server" {

  name         = "test-server-${var.tag}"
  machine_type = "n1-standard-2"
  zone         = var.google_zone
  tags         = ["allow-ssh", "allow-udp-all"]

  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }

  network_interface {
    network    = google_compute_network.production.id
    subnetwork = google_compute_subnetwork.production.id
    access_config {
      nat_ip = google_compute_address.test_server_address.address
    }
  }

  metadata = {
    startup-script = <<-EOF2
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a server.tar.gz
    cat <<EOF > /app/app.env
    ENV=prod
    NEXT_LOG_LEVEL=4
    NEXT_DATACENTER="cloud"
    NEXT_SERVER_ADDRESS="${google_compute_address.test_server_address.address}:30000"
    NEXT_SERVER_BACKEND_HOSTNAME="server.${var.cloudflare_domain}"
    NEXT_BUYER_PUBLIC_KEY="${var.test_buyer_public_key}"
    NEXT_BUYER_PRIVATE_KEY="${var.test_buyer_private_key}"
    NEXT_RELAY_BACKEND_PUBLIC_KEY="${var.relay_backend_public_key}"
    NEXT_SERVER_BACKEND_PUBLIC_KEY="${var.server_backend_public_key}"
    EOF
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/libnext.so /usr/local/lib/libnext.so
    ldconfig
    systemctl start app.service
    EOF2
  }

  service_account {
    email  = local.google_service_account
    scopes = ["cloud-platform"]
  }
}

output "test_server_address" {
  description = "The IP address of the test server"
  value = "${google_compute_address.test_server_address.address}:30000"
}

# ----------------------------------------------------------------------------------------
