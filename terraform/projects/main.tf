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

# create service accounts

resource "google_service_account" "terraform" {
  project  = google_project.storage.project_id
  account_id   = "terraform"
  display_name = "Terraform Service Account"
}

resource "google_service_account_key" "terraform" {
  service_account_id = google_service_account.terraform.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}

resource "google_service_account" "dev_runtime" {
  project  = google_project.dev.project_id
  account_id   = "dev-runtime"
  display_name = "Development Runtime Service Account"
}

resource "google_service_account" "staging_runtime" {
  project  = google_project.staging.project_id
  account_id   = "staging-runtime"
  display_name = "Staging Runtime Service Account"
}

resource "google_service_account" "prod_runtime" {
  project  = google_project.staging.project_id
  account_id   = "prod-runtime"
  display_name = "Production Runtime Service Account"
}

output google_terraform_json {
  value = google_service_account_key.terraform.private_key
  sensitive = true
}

resource "local_file" "google_terraform_json" {
    filename = "terraform-google.json"
    content  =  base64decode(google_service_account_key.terraform.private_key)
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

resource "google_storage_bucket" "backend_artifacts" {
  name          = "${local.company_name}_network_next_backend_artifacts"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  public_access_prevention = "enforced"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "relay_artifacts" {
  name          = "${local.company_name}_network_next_relay_artifacts"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_member" "relay_artifacts" {
  bucket = google_storage_bucket.relay_artifacts.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

resource "google_storage_bucket" "database_files" {
  name          = "${local.company_name}_network_next_database_files"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  public_access_prevention = "enforced"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "sql_files" {
  name          = "${local.company_name}_network_next_sql_files"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  public_access_prevention = "enforced"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "sdk_config" {
  name          = "${local.company_name}_network_next_sdk_config"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_member" "sdk_config" {
  bucket = google_storage_bucket.sdk_config.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

resource "google_storage_bucket" "terraform" {
  name          = "${local.company_name}_network_next_terraform"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_object" "amazon_txt" {
  name   = "amazon.txt"
  source = "../../config/amazon.txt"
  bucket = google_storage_bucket.sdk_config.name
}

resource "google_storage_bucket_object" "google_txt" {
  name   = "google.txt"
  source = "../../config/google.txt"
  bucket = google_storage_bucket.sdk_config.name
}

resource "google_storage_bucket_object" "multiplay_txt" {
  name   = "multiplay.txt"
  source = "../../config/multiplay.txt"
  bucket = google_storage_bucket.sdk_config.name
}

resource "google_storage_bucket_object" "akamai_txt" {
  name   = "akamai.txt"
  source = "../../config/akamai.txt"
  bucket = google_storage_bucket.sdk_config.name
}

resource "google_storage_bucket_object" "vultr_txt" {
  name   = "vultr.txt"
  source = "../../config/vultr.txt"
  bucket = google_storage_bucket.sdk_config.name
}

resource "google_storage_bucket_object" "dev_bin" {
  name   = "dev.bin"
  source = "../../envs/empty.bin"
  bucket = google_storage_bucket.database_files.name
}

resource "google_storage_bucket_object" "staging_bin" {
  name   = "staging.bin"
  source = "../../envs/staging.bin"
  bucket = google_storage_bucket.database_files.name
}

resource "google_storage_bucket_object" "prod_bin" {
  name   = "prod.bin"
  source = "../../envs/empty.bin"
  bucket = google_storage_bucket.database_files.name
}

resource "google_storage_bucket_object" "create_sql" {
  name   = "create.sql"
  source = "../../schemas/sql/create.sql"
  bucket = google_storage_bucket.sql_files.name
}

resource "google_storage_bucket_object" "destroy_sql" {
  name   = "destroy.sql"
  source = "../../schemas/sql/destroy.sql"
  bucket = google_storage_bucket.sql_files.name
}

resource "google_storage_bucket_object" "staging_sql" {
  name   = "staging.sql"
  source = "../../schemas/sql/staging.sql"
  bucket = google_storage_bucket.sql_files.name
}

resource "google_storage_bucket_iam_binding" "backend_artifacts" {
  bucket = google_storage_bucket.backend_artifacts.name
  role = "roles/storage.admin"
  members = [
    google_service_account.terraform.member,
  ]
}

resource "google_storage_bucket_iam_binding" "relay_artifacts" {
  bucket = google_storage_bucket.relay_artifacts.name
  role = "roles/storage.admin"
  members = [
    google_service_account.terraform.member,
  ]
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
