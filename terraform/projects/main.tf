# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0.0"
    }
  }
}

locals {
  org_id = "434699063105"
  billing_account = "018C15-D3C7AC-4722E8"
  company_name = "next"
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
  deletion_policy = "DELETE"
}

resource "google_project" "dev" {
  name            = "Development"
  project_id      = "dev-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
  deletion_policy = "DELETE"
}

resource "google_project" "dev_relays" {
  name            = "Development Relays"
  project_id      = "dev-relays-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
  deletion_policy = "DELETE"
}

resource "google_project" "staging" {
  name            = "Staging"
  project_id      = "staging-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
  deletion_policy = "DELETE"
}

resource "google_project" "prod" {
  name            = "Production"
  project_id      = "prod-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
  deletion_policy = "DELETE"
}

resource "google_project" "prod_relays" {
  name            = "Production Relays"
  project_id      = "prod-relays-${random_id.postfix.hex}"
  org_id          = local.org_id
  billing_account = local.billing_account
  deletion_policy = "DELETE"
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
    create = "5m"
    update = "5m"
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

resource "google_storage_bucket" "staging" {
  name          = "${local.company_name}_network_next_staging"
  project       = google_project.storage.project_id
  location      = "US"
  force_destroy = true
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "prod" {
  name          = "${local.company_name}_network_next_prod"
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

# make relay artifacts and sdk config publicly accessible

resource "google_storage_bucket_iam_member" "relay_artifacts_public_access" {
  bucket = google_storage_bucket.relay_artifacts.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

resource "google_storage_bucket_iam_member" "sdk_config_public_access" {
  bucket = google_storage_bucket.sdk_config.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# give the terraform storage service account storage admin permission over the project

resource "google_project_iam_member" "terraform_storage_admin_admin" {
  project = google_project.storage.project_id
  role    = "roles/storage.objectAdmin"
  member = google_service_account.terraform_storage.member
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

resource "google_storage_bucket_object" "akamai_txt" {
  name   = "akamai.txt"
  source = "../../config/akamai.txt"
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
    create = "5m"
    update = "5m"
  }
  disable_dependent_services = true
}

# setup service accounts for dev

resource "google_service_account" "terraform_dev" {
  project  = google_project.dev.project_id
  account_id   = "terraform-dev"
  display_name = "Terraform Service Account (Development)"
}

resource "google_project_iam_member" "dev_terraform_admin" {
  project = google_project.dev.project_id
  role    = "roles/admin"
  member  = google_service_account.terraform_dev.member
}

resource "google_service_account_key" "terraform_dev" {
  service_account_id = google_service_account.terraform_dev.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}

resource "google_service_account" "dev_runtime" {
  project  = google_project.dev.project_id
  account_id   = "dev-runtime"
  display_name = "Development Runtime Service Account"
}

resource "google_project_iam_member" "dev_terraform_runtime_compute_viewer" {
  project = google_project.dev.project_id
  role    = "roles/compute.viewer"
  member  = google_service_account.dev_runtime.member
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

# write the terraform dev key to "terraform-dev.json"

resource "local_file" "terraform_dev_json" {
    filename = "terraform-dev.json"
    content  =  base64decode(google_service_account_key.terraform_dev.private_key)
}

# write the dev project id to "dev-project-id.txt"

resource "local_file" "dev_project_id" {
  filename = "dev-project-id.txt"
  content  = replace(google_project.dev.id, "projects/", "")
}

# write the dev project number to "dev-project-number.txt"

resource "local_file" "dev_project_number" {
  filename = "dev-project-number.txt"
  content  = google_project.dev.number
}

# write the dev runtime service account email to "dev-runtime-service-account.txt"

resource "local_file" "dev_runtime_service_account" {
  filename = "dev-runtime-service-account.txt"
  content  =  google_service_account.dev_runtime.email
}

# give the dev runtime service account permission to publish pubsub messages

resource "google_project_iam_member" "dev_pubsub_publish" {
  project = google_project.dev.project_id
  role    = "roles/pubsub.publisher"
  member  = google_service_account.dev_runtime.member
}

# ----------------------------------------------------------------------------------------
#                                        DEV RELAYS
# ----------------------------------------------------------------------------------------

resource "google_project_service" "dev_relays" {
  project  = google_project.dev_relays.project_id
  service  = "compute.googleapis.com"
  timeouts {
    create = "5m"
    update = "5m"
  }
  disable_dependent_services = true
  provisioner "local-exec" {
    command = <<EOF
for i in {1..5}; do
  sleep $i
  if gcloud services list --project="${google_project.dev_relays.project_id}" | grep "compute.googleapis.com"; then
    exit 0
  fi
done
echo "Service was not enabled after 15s"
exit 1
EOF
  }
}

# create "relays" network in the project. we don't use "default" network because some google cloud orgs disable default network creation for security reasons

resource "google_compute_network" "dev_relays" {
  name                    = "relays"
  project                 = google_project.dev_relays.project_id
  auto_create_subnetworks = true
  depends_on              = [google_project_service.dev_relays]
}

# setup service account for dev relays

resource "google_service_account" "terraform_dev_relays" {
  project      = google_project.dev_relays.project_id
  account_id   = "terraform-dev-relays"
  display_name = "Terraform Service Account (Development Relays)"
  depends_on   = [google_project_service.dev_relays]
}

resource "google_project_iam_member" "dev_relays_terraform_admin" {
  project    = google_project.dev_relays.project_id
  role       = "roles/admin"
  member     = google_service_account.terraform_dev_relays.member
  depends_on = [google_project_service.dev_relays]
}

resource "google_service_account_key" "terraform_dev_relays" {
  service_account_id = google_service_account.terraform_dev_relays.name
  public_key_type    = "TYPE_X509_PEM_FILE"
  depends_on         = [google_project_service.dev_relays]
}

# write the terraform dev relays private key to "terraform-dev-relays.json"

resource "local_file" "terraform_dev_relays_json" {
  filename   = "terraform-dev-relays.json"
  content    =  base64decode(google_service_account_key.terraform_dev_relays.private_key)
  depends_on = [google_project_service.dev_relays]
}

# write the dev relays project id to "dev-relays-project-id.txt"

resource "local_file" "dev_relays_project_id" {
  filename   = "dev-relays-project-id.txt"
  content    = replace(google_project.dev_relays.id, "projects/", "")
  depends_on = [google_project_service.dev_relays]
}

# write the dev relays project number to "dev-relays-project-number.txt"

resource "local_file" "dev_relays_project_number" {
  filename   = "dev-relays-project-number.txt"
  content    = google_project.dev_relays.number
  depends_on = [google_project_service.dev_relays]
}

# ----------------------------------------------------------------------------------------
#                                         STAGING
# ----------------------------------------------------------------------------------------

locals {
  staging_services = [
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

resource "google_project_service" "staging" {
  count    = length(local.staging_services)
  project  = google_project.staging.project_id
  service  = local.staging_services[count.index]
  timeouts {
    create = "5m"
    update = "5m"
  }
  disable_dependent_services = true
}

# setup service accounts for staging

resource "google_service_account" "terraform_staging" {
  project  = google_project.staging.project_id
  account_id   = "terraform-staging"
  display_name = "Terraform Service Account (Staging)"
}

resource "google_project_iam_member" "staging_terraform_admin" {
  project = google_project.staging.project_id
  role    = "roles/admin"
  member  = google_service_account.terraform_staging.member
}

resource "google_service_account_key" "terraform_staging" {
  service_account_id = google_service_account.terraform_staging.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}

resource "google_service_account" "staging_runtime" {
  project  = google_project.staging.project_id
  account_id   = "staging-runtime"
  display_name = "Staging Runtime Service Account"
}

resource "google_project_iam_member" "staging_terraform_runtime_compute_viewer" {
  project = google_project.staging.project_id
  role    = "roles/compute.viewer"
  member  = google_service_account.staging_runtime.member
}

resource "google_storage_bucket_iam_member" "staging_runtime_backend_artifacts_storage_viewer" {
  bucket = google_storage_bucket.backend_artifacts.name
  role   = "roles/storage.objectViewer"
  member = google_service_account.staging_runtime.member
  depends_on = [google_storage_bucket.backend_artifacts]
}

resource "google_storage_bucket_iam_member" "staging_runtime_database_files_storage_admin" {
  bucket = google_storage_bucket.database_files.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.staging_runtime.member
  depends_on = [google_storage_bucket.database_files]
}

resource "google_storage_bucket_iam_member" "staging_runtime_staging_storage_admin" {
  bucket = google_storage_bucket.staging.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.staging_runtime.member
  depends_on = [google_storage_bucket.staging]
}

resource "google_storage_bucket_iam_member" "terraform_staging_object_admin" {
  bucket = google_storage_bucket.terraform.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.terraform_staging.member
  depends_on = [google_storage_bucket.terraform]
}

# write the terraform staging private key to "terraform-staging.json"

resource "local_file" "terraform_staging_json" {
    filename = "terraform-staging.json"
    content  =  base64decode(google_service_account_key.terraform_staging.private_key)
}

# write the staging project id to "staging-project-id.txt"

resource "local_file" "staging_project" {
  filename = "staging-project-id.txt"
  content  = replace(google_project.staging.id, "projects/", "")
}

# write the staging project number to "staging-project-number.txt"

resource "local_file" "staging_project_number" {
  filename = "staging-project-number.txt"
  content  = google_project.staging.number
}

# write the staging runtime service account email to "staging-runtime-service-account.txt"

resource "local_file" "staging_runtime_service_account" {
  filename = "staging-runtime-service-account.txt"
  content  =  google_service_account.staging_runtime.email
}

# give the staging runtime service account permission to publish pubsub messages

resource "google_project_iam_member" "staging_pubsub_publish" {
  project = google_project.staging.project_id
  role    = "roles/pubsub.publisher"
  member  = google_service_account.staging_runtime.member
}

# ----------------------------------------------------------------------------------------
#                                          PROD
# ----------------------------------------------------------------------------------------

locals {
  prod_services = [
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

resource "google_project_service" "prod" {
  count    = length(local.prod_services)
  project  = google_project.prod.project_id
  service  = local.prod_services[count.index]
  timeouts {
    create = "5m"
    update = "5m"
  }
  disable_dependent_services = true
}

# setup service accounts for prod

resource "google_service_account" "terraform_prod" {
  project  = google_project.prod.project_id
  account_id   = "terraform-prod"
  display_name = "Terraform Service Account (Production)"
}

resource "google_project_iam_member" "prod_terraform_admin" {
  project = google_project.prod.project_id
  role    = "roles/admin"
  member  = google_service_account.terraform_prod.member
}

resource "google_service_account_key" "terraform_prod" {
  service_account_id = google_service_account.terraform_prod.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}

resource "google_service_account" "prod_runtime" {
  project  = google_project.prod.project_id
  account_id   = "prod-runtime"
  display_name = "Production Runtime Service Account"
}

resource "google_project_iam_member" "prod_terraform_runtime_compute_viewer" {
  project = google_project.prod.project_id
  role    = "roles/compute.viewer"
  member  = google_service_account.prod_runtime.member
}

resource "google_storage_bucket_iam_member" "prod_runtime_backend_artifacts_storage_viewer" {
  bucket = google_storage_bucket.backend_artifacts.name
  role   = "roles/storage.objectViewer"
  member = google_service_account.prod_runtime.member
  depends_on = [google_storage_bucket.backend_artifacts]
}

resource "google_storage_bucket_iam_member" "prod_runtime_database_files_storage_admin" {
  bucket = google_storage_bucket.database_files.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.prod_runtime.member
  depends_on = [google_storage_bucket.database_files]
}

resource "google_storage_bucket_iam_member" "prod_runtime_prod_storage_admin" {
  bucket = google_storage_bucket.prod.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.prod_runtime.member
  depends_on = [google_storage_bucket.prod]
}

resource "google_storage_bucket_iam_member" "terraform_prod_object_admin" {
  bucket = google_storage_bucket.terraform.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.terraform_prod.member
  depends_on = [google_storage_bucket.terraform]
}

# write the terraform prod private key to "terraform-prod.json"

resource "local_file" "terraform_prod_json" {
    filename = "terraform-prod.json"
    content  =  base64decode(google_service_account_key.terraform_prod.private_key)
}

# write the prod project id to "prod-project-id.txt"

resource "local_file" "prod_project_id" {
  filename = "prod-project-id.txt"
  content  = replace(google_project.prod.id, "projects/", "")
}

# write the prod project number to "prod-project-number.txt"

resource "local_file" "prod_project_number" {
  filename = "prod-project-number.txt"
  content  = google_project.prod.number
}

# write the prod runtime service account email to "prod-runtime-service-account.txt"

resource "local_file" "prod_runtime_service_account" {
  filename = "prod-runtime-service-account.txt"
  content  =  google_service_account.prod_runtime.email
}

# give the prod runtime service account permission to publish pubsub messages

resource "google_project_iam_member" "prod_pubsub_publish" {
  project = google_project.prod.project_id
  role    = "roles/pubsub.publisher"
  member  = google_service_account.prod_runtime.member
}

# ----------------------------------------------------------------------------------------
#                                       PROD RELAYS
# ----------------------------------------------------------------------------------------

resource "google_project_service" "prod_relays" {
  project  = google_project.prod_relays.project_id
  service  = "compute.googleapis.com"
  timeouts {
    create = "5m"
    update = "5m"
  }
  disable_dependent_services = true
  provisioner "local-exec" {
    command = <<EOF
for i in {1..5}; do
  sleep $i
  if gcloud services list --project="${google_project.prod_relays.project_id}" | grep "compute.googleapis.com"; then
    exit 0
  fi
done
echo "Service was not enabled after 15s"
exit 1
EOF
  }
}

# create "relays" network in the project. we don't use "default" network because some google cloud orgs disable default network creation for security reasons

resource "google_compute_network" "prod_relays" {
  name                    = "relays"
  project                 = google_project.prod_relays.project_id
  auto_create_subnetworks = true
  depends_on              = [google_project_service.prod_relays]
}

# setup service account for prod relays

resource "google_service_account" "terraform_prod_relays" {
  project      = google_project.prod_relays.project_id
  account_id   = "terraform-prod-relays"
  display_name = "Terraform Service Account (Production Relays)"
  depends_on   = [google_project_service.prod_relays]
}

resource "google_project_iam_member" "prod_relays_terraform_admin" {
  project    = google_project.prod_relays.project_id
  role       = "roles/admin"
  member     = google_service_account.terraform_prod_relays.member
  depends_on = [google_project_service.prod_relays]
}

resource "google_service_account_key" "terraform_prod_relays" {
  service_account_id = google_service_account.terraform_prod_relays.name
  public_key_type    = "TYPE_X509_PEM_FILE"
  depends_on         = [google_project_service.prod_relays]
}

# write the prod relays private key to "terraform-prod-relays.json"

resource "local_file" "terraform_prod_relays_json" {
  filename   = "terraform-prod-relays.json"
  content    =  base64decode(google_service_account_key.terraform_prod_relays.private_key)
  depends_on = [google_project_service.prod_relays]
}

# write the prod relays project id to "prod-relays-project-id.txt"

resource "local_file" "prod_relays_project_id" {
  filename   = "prod-relays-project-id.txt"
  content    = replace(google_project.prod_relays.id, "projects/", "")
  depends_on = [google_project_service.prod_relays]
}

# write the prod relays project number to "prod-relays-project-number.txt"

resource "local_file" "prod_relays_project_number" {
  filename   = "prod-relays-project-number.txt"
  content    = google_project.prod_relays.number
  depends_on = [google_project_service.prod_relays]
}

# ----------------------------------------------------------------------------------------
