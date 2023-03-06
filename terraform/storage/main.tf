# ----------------------------------------------------------------------------------------

variable "credentials" { type = string }
variable "project" { type = string }
variable "location" { type = string }
variable "region" { type = string }
variable "zone" { type = string }
variable "dev_artifacts" { type = string }
variable "relay_artifacts" { type = string }
variable "sdk_config" { type = string }

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

provider "google" {
  credentials = file(var.credentials)
  project     = var.project
  region      = var.region
  zone        = var.zone
}

# ----------------------------------------------------------------------------------------

resource "google_storage_bucket" "dev-artifacts" {
  name                        = var.dev_artifacts
  storage_class               = "MULTI_REGIONAL"
  location                    = var.location
  public_access_prevention    = "enforced"
  uniform_bucket_level_access = true
  force_destroy               = true
}

# ----------------------------------------------------------------------------------------

resource "google_storage_bucket" "relay-artifacts" {
  name                        = var.relay_artifacts
  storage_class               = "MULTI_REGIONAL"
  location                    = var.location
  uniform_bucket_level_access = true
  force_destroy               = true
}

resource "google_storage_bucket_iam_member" "relay-artifacts" {
  bucket   = google_storage_bucket.relay-artifacts.name
  role     = "roles/storage.objectViewer"
  member   = "allUsers"
}

# ----------------------------------------------------------------------------------------

resource "google_storage_bucket" "sdk-config" {
  name                        = var.sdk_config
  storage_class               = "MULTI_REGIONAL"
  location                    = var.location
  uniform_bucket_level_access = true
  force_destroy               = true
}

resource "google_storage_bucket_iam_member" "sdk-config" {
  bucket   = google_storage_bucket.sdk-config.name
  role     = "roles/storage.objectViewer"
  member   = "allUsers"
}

# ----------------------------------------------------------------------------------------
