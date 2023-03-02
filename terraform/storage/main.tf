# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

provider "google" {
  credentials = file("~/Documents/terraform.json")
  project = "heroic-grove-379322"
  region  = "us-central1"
  zone    = "us-central1-c"
}

# ----------------------------------------------------------------------------------------

resource "google_storage_bucket" "dev-artifacts" {
  name = "network_next_dev_artifacts"
  storage_class = "MULTI_REGIONAL"
  location = "US"
  force_destroy = true
}

resource "google_storage_bucket" "relay-artifacts" {
  name = "network_next_relay_artifacts"
  storage_class = "MULTI_REGIONAL"
  location = "US"
  force_destroy = true
}

# ----------------------------------------------------------------------------------------
