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
#                                         STORAGE
# ----------------------------------------------------------------------------------------

locals {
  storage_services = [
    "storage.googleapis.com",
  ]
}

resource "google_project_service" "storage" {
  count    = length(local.storage_services)
  project  = google_project.storage.project_id
  service  = local.storage_services[count.index]
  timeouts {
    create = "30m"
    update = "40m"
  }
  disable_dependent_services = true
}

# create buckets

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

resource "google_storage_bucket" "terraform" {
  name          = "${local.company_name}_network_next_terraform"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
}

resource "google_storage_bucket" "dev" {
  name          = "${local.company_name}_network_next_dev"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  uniform_bucket_level_access = true
}

# create service accounts so semaphore can upload artifacts

resource "google_service_account" "terraform_storage" {
  project  = google_project.storage.project_id
  account_id   = "terraform-storage"
  display_name = "Terraform Service Account (Storage)"
}

resource "google_service_account_key" "terraform_storage" {
  service_account_id = google_service_account.terraform_storage.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}

# write the storage terraform service account key to "terraform-storage.json"

resource "local_file" "terraform_storage_json" {
    filename = "terraform-storage.json"
    content  =  base64decode(google_service_account_key.terraform_storage.private_key)
}

# setup bucket permissions

resource "google_storage_bucket_iam_member" "relay_artifacts" {
  bucket = google_storage_bucket.relay_artifacts.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

resource "google_storage_bucket_iam_member" "sdk_config" {
  bucket = google_storage_bucket.sdk_config.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

resource "google_storage_bucket_iam_member" "terraform_storage_object_admin" {
  bucket = google_storage_bucket.terraform.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.terraform_storage.member
  depends_on = [google_storage_bucket.terraform]
}

resource "google_storage_bucket_iam_member" "backend_artifacts_storage_object_admin" {
  bucket = google_storage_bucket.backend_artifacts.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.terraform_storage.member
  depends_on = [google_storage_bucket.terraform]
}

resource "google_storage_bucket_iam_member" "relay_artifacts_storage_object_admin" {
  bucket = google_storage_bucket.relay_artifacts.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.terraform_storage.member
  depends_on = [google_storage_bucket.terraform]
}

# upload config files read by the SDK

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

# upload database files necessary to bootstrap the envs

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

# upload sql files for initializing the postgres db

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

# ----------------------------------------------------------------------------------------
#                                           DEV
# ----------------------------------------------------------------------------------------

locals {
  dev_services = [
    "pubsub.googleapis.com",                    # google pubsub
    "storage.googleapis.com",                   # cloud storage
    "bigquery.googleapis.com",                  # bigquery
    "compute.googleapis.com",                   # compute engine
    "redis.googleapis.com",                     # redis
    "sqladmin.googleapis.com",                  # postgres
    "cloudresourcemanager.googleapis.com",      # cloud resource manager
    "servicenetworking.googleapis.com"          # service networking
  ]
}

resource "google_project_service" "dev" {
  count    = length(local.dev_services)
  project  = google_project.dev.project_id
  service  = local.dev_services[count.index]
  timeouts {
    create = "30m"
    update = "40m"
  }
  disable_dependent_services = true
}

# setup service accounts for dev

resource "google_service_account" "terraform_dev" {
  project  = google_project.dev.project_id
  account_id   = "terraform-dev"
  display_name = "Terraform Service Account (Development)"
}

resource "google_project_iam_member" "dev_terraform_editor" {
  project = google_project.dev.project_id
  role    = "roles/admin"
  member  = google_service_account.terraform_dev.member
}

resource "google_service_account_key" "terraform_dev" {
  service_account_id = google_service_account.terraform_dev.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}

resource "local_file" "terraform_dev_json" {
    filename = "terraform-dev.json"
    content  =  base64decode(google_service_account_key.terraform_dev.private_key)
}

resource "google_service_account" "dev_runtime" {
  project  = google_project.dev.project_id
  account_id   = "dev-runtime"
  display_name = "Development Runtime Service Account"
}

resource "google_storage_bucket_iam_member" "dev_runtime_backend_artifacts_storage_viewer" {
  bucket = google_storage_bucket.backend_artifacts.name
  role   = "roles/storage.objectViewer"
  member = google_service_account.dev_runtime.member
  depends_on = [google_storage_bucket.backend_artifacts]
}

resource "google_storage_bucket_iam_member" "dev_runtime_database_files_storage_admin" {
  bucket = google_storage_bucket.database_files.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.dev_runtime.member
  depends_on = [google_storage_bucket.database_files]
}

resource "google_storage_bucket_iam_member" "dev_runtime_dev_storage_admin" {
  bucket = google_storage_bucket.dev.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.dev_runtime.member
  depends_on = [google_storage_bucket.dev]
}

resource "google_storage_bucket_iam_member" "terraform_dev_object_admin" {
  bucket = google_storage_bucket.terraform.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.terraform_dev.member
  depends_on = [google_storage_bucket.terraform]
}

# write the dev project id to "dev-project.txt"

resource "local_file" "dev_project" {
  filename = "dev-project.txt"
  content  = replace(google_project.dev.id, "projects/", "")
}

# write the dev runtime service account email to "dev-runtime-service-account.txt"

resource "local_file" "dev_runtime_service_account" {
  filename = "dev-runtime-service-account.txt"
  content  =  google_service_account.dev_runtime.email
}

# give the existing internal pubsub service account bigquery admin permissions (otherwise, we can't create the pubsub subscriptions that write to bigquery...)

resource "google_project_iam_member" "pubsub_bigquery_admin" {
  project = google_project.dev.project_id
  role    = "roles/bigquery.admin"
  member  = "serviceAccount:service-${google_project.dev.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
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
