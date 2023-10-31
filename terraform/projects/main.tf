# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0.0"
    }
  }
}

locals {
  org_id = "434699063105"
  billing_account = "018C15-D3C7AC-4722E8"
  company_name = "auspicious"
}

# ----------------------------------------------------------------------------------------

# create projects

resource "random_id" "postfix" {
  byte_length = 8
}

resource "google_project" "storage" {
  name            = "Storage"
  project_id      = "storage-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
}

resource "google_project" "dev" {
  name            = "Development"
  project_id      = "dev-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
}

resource "google_project" "dev_relays" {
  name            = "Development Relays"
  project_id      = "dev-relays-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
}

resource "google_project" "staging" {
  name            = "Staging"
  project_id      = "staging-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
}

resource "google_project" "prod" {
  name            = "Production"
  project_id      = "prod-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
}

resource "google_project" "prod_relays" {
  name            = "Production Relays"
  project_id      = "prod-relays-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
}

# ----------------------------------------------------------------------------------------

# configure storage project

locals {
  storage_services = [
    "storage.googleapis.com",         # cloud storage
  ]
}

resource "google_project_service" "storage" {
  count    = length(local.storage_services)
  project  = google_project.storage.project_id
  service  = "pubsub.googleapis.com"
  timeouts {
    create = "30m"
    update = "40m"
  }
  disable_dependent_services = true
}

resource "google_storage_bucket" "dev_backend_artifacts" {
  name          = "${local.company_name}_network_next_backend_artifacts"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  public_access_prevention = "enforced"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "dev_relay_artifacts" {
  name          = "${local.company_name}_network_next_relay_artifacts"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_member" "dev_relay_artifacts" {
  bucket = google_storage_bucket.dev_relay_artifacts.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

resource "google_storage_bucket" "dev_database_files" {
  name          = "${local.company_name}_network_next_database_files"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  public_access_prevention = "enforced"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "dev_sdk_config" {
  name          = "${local.company_name}_network_next_sdk_config"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_member" "dev_sdk_config" {
  bucket = google_storage_bucket.dev_sdk_config.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# ----------------------------------------------------------------------------------------

# configure dev project

locals {
  dev_services = [
    "pubsub.googleapis.com",          # google pubsub
    "storage.googleapis.com",         # cloud storage
    "bigquery.googleapis.com",        # bigquery
    "compute.googleapis.com",         # compute engine
    "redis.googleapis.com",           # redis
    "sql-component.googleapis.com",   # postgres
  ]
}

resource "google_project_service" "dev" {
  count    = length(local.dev_services)
  project  = google_project.dev.project_id
  service  = "pubsub.googleapis.com"
  timeouts {
    create = "30m"
    update = "40m"
  }
  disable_dependent_services = true
}

# ----------------------------------------------------------------------------------------

# configure dev relays project

locals {
  dev_relays_services = [
    "compute.googleapis.com",         # compute engine
  ]
}

resource "google_project_service" "dev_relays" {
  count    = length(local.dev_relays_services)
  project  = google_project.dev_relays.project_id
  service  = "pubsub.googleapis.com"
  timeouts {
    create = "30m"
    update = "40m"
  }
  disable_dependent_services = true
}

# ----------------------------------------------------------------------------------------

# configure staging project

locals {
  staging_services = [
    "pubsub.googleapis.com",          # google pubsub
    "storage.googleapis.com",         # cloud storage
    "bigquery.googleapis.com",        # bigquery
    "compute.googleapis.com",         # compute engine
    "redis.googleapis.com",           # redis
    "sql-component.googleapis.com",   # postgres
  ]
}

resource "google_project_service" "staging" {
  count    = length(local.staging_services)
  project  = google_project.staging.project_id
  service  = "pubsub.googleapis.com"
  timeouts {
    create = "30m"
    update = "40m"
  }
  disable_dependent_services = true
}

# ----------------------------------------------------------------------------------------

# configure prod project

locals {
  prod_services = [
    "pubsub.googleapis.com",          # google pubsub
    "storage.googleapis.com",         # cloud storage
    "bigquery.googleapis.com",        # bigquery
    "compute.googleapis.com",         # compute engine
    "redis.googleapis.com",           # redis
    "sql-component.googleapis.com",   # postgres
  ]
}

resource "google_project_service" "prod" {
  count    = length(local.prod_services)
  project  = google_project.prod.project_id
  service  = "pubsub.googleapis.com"
  timeouts {
    create = "30m"
    update = "40m"
  }
  disable_dependent_services = true
}

# ----------------------------------------------------------------------------------------

# configure prod relays project

locals {
  prod_relays_services = [
    "compute.googleapis.com",         # compute engine
  ]
}

resource "google_project_service" "prod_relays" {
  count    = length(local.prod_relays_services)
  project  = google_project.prod_relays.project_id
  service  = "pubsub.googleapis.com"
  timeouts {
    create = "30m"
    update = "40m"
  }
  disable_dependent_services = true
}

# ----------------------------------------------------------------------------------------
