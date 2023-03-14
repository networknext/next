# ========================================================================================
#                                     GOOGLE CLOUD
# ========================================================================================

/*
    Here is the full set of google cloud datacenters, as of March 13, 2023:

      google.taiwan.1
      google.taiwan.2
      google.taiwan.3
      google.hongkong.1
      google.hongkong.2
      google.hongkong.3
      google.tokyo.1
      google.tokyo.2
      google.tokyo.3
      google.osaka.1
      google.osaka.2
      google.osaka.3
      google.seoul.1
      google.seoul.2
      google.seoul.3
      google.mumbai.1
      google.mumbai.2
      google.mumbai.3
      google.delhi.1
      google.delhi.2
      google.delhi.3
      google.singapore.1
      google.singapore.2
      google.singapore.3
      google.jakarta.1
      google.jakarta.2
      google.jakarta.3
      google.sydney.1
      google.sydney.2
      google.sydney.3
      google.melbourne.1
      google.melbourne.2
      google.melbourne.3
      google.warsaw.1
      google.warsaw.2
      google.warsaw.3
      google.finland.1
      google.finland.2
      google.finland.3
      google.madrid.1,
      google.madrid.2,
      google.madrid.3,
      google.belgium.1
      google.belgium.2
      google.belgium.3
      google.london.1
      google.london.2
      google.london.3
      google.frankfurt.1
      google.frankfurt.2
      google.frankfurt.3
      google.netherlands.1
      google.netherlands.2
      google.netherlands.3
      google.zurich.1
      google.zurich.2
      google.zurich.3
      google.milan.1
      google.milan.2
      google.milan.3
      google.paris.1
      google.paris.2
      google.paris.3
      google.telaviv.1
      google.telaviv.2
      google.telaviv.3
      google.montreal.1
      google.montreal.2
      google.montreal.3
      google.toronto.1
      google.toronto.2
      google.toronto.3
      google.saopaulo.1
      google.saopaulo.2
      google.saopaulo.3
      google.santiago.1
      google.santiago.2
      google.santiago.3
      google.iowa.1
      google.iowa.2
      google.iowa.3
      google.iowa.4
      google.southcarolina.1
      google.southcarolina.2
      google.southcarolina.3
      google.virginia.1
      google.virginia.2
      google.virginia.3
      google.ohio.1
      google.ohio.2
      google.ohio.3
      google.dallas.1
      google.dallas.2
      google.dallas.3
      google.oregon.1
      google.oregon.2
      google.oregon.3
      google.losangeles.1
      google.losangeles.2
      google.losangeles.3
      google.saltlakecity.1
      google.saltlakecity.2
      google.saltlakecity.3
      google.lasvegas.1
      google.lasvegas.2
      google.lasvegas.3

    See config/google.txt for more details, or schemas/sql/google.sql

    You can derive an updated version of this list from: https://cloud.google.com/compute/docs/regions-zones
*/

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
  credentials = file(var.credentials)
  project     = var.project
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }
variable "project" { type = string }
variable "credentials" { type = string }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# ----------------------------------------------------------------------------------------

data "google_compute_network" "default" {
  name = "default"
}

resource "google_compute_firewall" "google_allow_ssh" {
  name          = "allow-ssh"
  project       = var.project
  direction     = "INGRESS"
  network       = "default"
  source_ranges = [var.vpn_address]
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}

resource "google_compute_firewall" "google_allow_udp" {
  name          = "allow-udp"
  project       = var.project
  direction     = "INGRESS"
  network       = "default"
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
    ports    = ["40000"]
  }
}

# ----------------------------------------------------------------------------------------

locals {

  /*
      IMPORTANT: This map translates from the network next google datacenter names to google cloud zone and region.
      
      If want you add more datacenters from google cloud, then you must:

          1. Update the list of supported google datacenters at the top of this file

          2. Add an entry in the map below for the new datacenter

          3. Update schemas/sql/sellers/google.sql then update your postgres database

          4. Update config/google.txt then "Deploy Config" config to your environment via semaphore

      Please be extremely careful making these changes!
  */

  datacenter_map = {

    "google.taiwan.1" = {
      zone   = "asia-east1-a"
      region = "asia-east1"
    },

    "google.taiwan.2" = {
      zone   = "asia-east1-b"
      region = "asia-east1"
    },

    "google.taiwan.3" = {
      zone   = "asia-east1-c"
      region = "asia-east1"
    },

    "google.hongkong.1" = {
      zone   = "asia-east2-a"
      region = "asia-east2"
    },

    "google.hongkong.2" = {
      zone   = "asia-east2-b"
      region = "asia-east2"
    },

    "google.hongkong.3" = {
      zone   = "asia-east2-c"
      region = "asia-east2"
    },

    "google.tokyo.1" = {
      zone   = "asia-northeast1-a"
      region = "asia-northeast1"
    },

    "google.tokyo.2" = {
      zone   = "asia-northeast1-b"
      region = "asia-northeast1"
    },

    "google.tokyo.3" = {
      zone   = "asia-northeast1-c"
      region = "asia-northeast1"
    }

    "google.osaka.1" = {
      zone   = "asia-northeast2-a"
      region = "asia-northeast2"
    },

    "google.osaka.2" = {
      zone   = "asia-northeast2-b"
      region = "asia-northeast2"
    },

    "google.osaka.3" = {
      zone   = "asia-northeast2-c"
      region = "asia-northeast2"
    },

    "google.seoul.1" = {
      zone   = "asia-northeast3-a"
      region = "asia-northeast3"
    },

    "google.seoul.2" = {
      zone   = "asia-northeast3-b"
      region = "asia-northeast3"
    },

    "google.seoul.3" = {
      zone   = "asia-northeast3-c"
      region = "asia-northeast3"
    },

    "google.mumbai.1" = {
      zone   = "asia-south1-a"
      region = "asia-south1"
    },

    "google.mumbai.2" = {
      zone   = "asia-south1-b"
      region = "asia-south1"
    },

    "google.mumbai.3" = {
      zone   = "asia-south1-c"
      region = "asia-south1"
    },

    "google.delhi.1" = {
      zone   = "asia-south2-a"
      region = "asia-south2"
    },

    "google.delhi.2" = {
      zone   = "asia-south2-b"
      region = "asia-south2"
    },

    "google.delhi.3" = {
      zone   = "asia-south2-c"
      region = "asia-south2"
    },

    "google.singapore.1" = {
      zone   = "asia-southeast1-a"
      region = "asia-southeast1"
    },

    "google.singapore.2" = {
      zone   = "asia-southeast1-b"
      region = "asia-southeast1"
    },

    "google.singapore.3" = {
      zone   = "asia-southeast1-c"
      region = "asia-southeast1"
    },

    "google.jakarta.1" = {
      zone   = "asia-southeast2-a"
      region = "asia-southeast2"
    },

    "google.jakarta.2" = {
      zone   = "asia-southeast2-b"
      region = "asia-southeast2"
    },

    "google.jakarta.3" = {
      zone   = "asia-southeast2-c"
      region = "asia-southeast2"
    },

    "google.sydney.1" = {
      zone   = "australia-southeast1-a"
      region = "australia-southeast1"
    },

    "google.sydney.2" = {
      zone   = "australia-southeast1-b"
      region = "australia-southeast1"
    },

    "google.sydney.3" = {
      zone   = "australia-southeast1-c"
      region = "australia-southeast1"
    },

    "google.melbourne.1" = {
      zone   = "australia-southeast2-a"
      region = "australia-southeast2"
    },

    "google.melbourne.2" = {
      zone   = "australia-southeast2-b"
      region = "australia-southeast2"
    },

    "google.melbourne.3" = {
      zone   = "australia-southeast2-c"
      region = "australia-southeast2"
    },

    "google.warsaw.1" = {
      zone   = "europe-central2-a"
      region = "europe-central2"
    },

    "google.warsaw.2" = {
      zone   = "europe-central2-b"
      region = "europe-central2"
    },

    "google.warsaw.3" = {
      zone   = "europe-central2-c"
      region = "europe-central2"
    },

    "google.finland.1" = {
      zone   = "europe-north1-a"
      region = "europe-north1"
    },

    "google.finland.2" = {
      zone   = "europe-north1-b"
      region = "europe-north1"
    },

    "google.finland.3" = {
      zone   = "europe-north1-c"
      region = "europe-north1"
    },

    "google.madrid.1" = {
      zone   = "europe-southwest1-a"
      region = "europe-southwest1"
    },

    "google.madrid.2" = {
      zone   = "europe-southwest1-b"
      region = "europe-southwest1"
    },

    "google.madrid.3" = {
      zone   = "europe-southwest1-c"
      region = "europe-southwest1"
    },

    "google.belgium.1" = {
      zone   = "europe-west1-a"
      region = "europe-west1"
    },

    "google.belgium.2" = {
      zone   = "europe-west1-b"
      region = "europe-west1"
    },

    "google.belgium.3" = {
      zone   = "europe-west1-c"
      region = "europe-west1"
    },

    "google.london.1" = {
      zone   = "europe-west2-a"
      region = "europe-west2"
    },

    "google.london.2" = {
      zone   = "europe-west2-b"
      region = "europe-west2"
    },

    "google.london.3" = {
      zone   = "europe-west2-c"
      region = "europe-west2"
    },

    "google.frankfurt.1" = {
      zone   = "europe-west3-a"
      region = "europe-west3"
    },

    "google.frankfurt.2" = {
      zone   = "europe-west3-b"
      region = "europe-west3"
    },

    "google.frankfurt.3" = {
      zone   = "europe-west3-c"
      region = "europe-west3"
    },

    "google.netherlands.1" = {
      zone   = "europe-west4-a"
      region = "europe-west4"
    },

    "google.netherlands.2" = {
      zone   = "europe-west4-b"
      region = "europe-west4"
    },

    "google.netherlands.3" = {
      zone   = "europe-west4-c"
      region = "europe-west4"
    },

    "google.zurich.1" = {
      zone   = "europe-west6-a"
      region = "europe-west6"
    },

    "google.zurich.2" = {
      zone   = "europe-west6-b"
      region = "europe-west6"
    },

    "google.zurich.3" = {
      zone   = "europe-west6-c"
      region = "europe-west6"
    },

    "google.milan.1" = {
      zone   = "europe-west8-a"
      region = "europe-west"
    },

    "google.milan.2" = {
      zone   = "europe-west8-b"
      region = "europe-west"
    },

    "google.milan.3" = {
      zone   = "europe-west8-c"
      region = "europe-west"
    },

    "google.paris.1" = {
      zone   = "europe-west9-a"
      region = "europe-west"
    },

    "google.paris.2" = {
      zone   = "europe-west9-b"
      region = "europe-west"
    },

    "google.paris.3" = {
      zone   = "europe-west9-c"
      region = "europe-west"
    },

    "google.telaviv.1" = {
      zone   = "me-west1-a"
      region = "me-west1"
    },

    "google.telaviv.2" = {
      zone   = "me-west1-b"
      region = "me-west1"
    },

    "google.telaviv.3" = {
      zone   = "me-west1-c"
      region = "me-west1"
    },

    "google.montreal.1" = {
      zone   = "northamerica-northeast1-a"
      region = "northamerica-northeast1"
    },

    "google.montreal.2" = {
      zone   = "northamerica-northeast1-b"
      region = "northamerica-northeast1"
    },

    "google.montreal.3" = {
      zone   = "northamerica-northeast1-c"
      region = "northamerica-northeast1"
    },

    "google.toronto.1" = {
      zone   = "northamerica-northeast2-a"
      region = "northamerica-northeast2"
    },

    "google.toronto.2" = {
      zone   = "northamerica-northeast2-b"
      region = "northamerica-northeast2"
    },

    "google.toronto.3" = {
      zone   = "northamerica-northeast2-c"
      region = "northamerica-northeast2"
    },

    "google.saopaulo.1" = {
      zone   = "southamerica-east1-a"
      region = "southamerica-east1"
    },

    "google.saopaulo.2" = {
      zone   = "southamerica-east1-b"
      region = "southamerica-east1"
    },

    "google.saopaulo.3" = {
      zone   = "southamerica-east1-c"
      region = "southamerica-east1"
    },

    "google.santiago.1" = {
      zone   = "southamerica-west1-a"
      region = "southamerica-west1"
    },

    "google.santiago.2" = {
      zone   = "southamerica-west1-b"
      region = "southamerica-west1"
    },

    "google.santiago.3" = {
      zone   = "southamerica-west1-c"
      region = "southamerica-west1"
    },

    "google.iowa.1" = {
      zone   = "us-central1-a"
      region = "us-central1"
    },

    "google.iowa.2" = {
      zone   = "us-central1-b"
      region = "us-central1"
    },

    "google.iowa.3" = {
      zone   = "us-central1-c"
      region = "us-central1"
    },

    "google.iowa.4" = {
      zone   = "us-central1-f"
      region = "us-central1"
    },

    "google.southcarolina.1" = {
      zone   = "us-east1-b"
      region = "us-east1"
    },

    "google.southcarolina.2" = {
      zone   = "us-east1-c"
      region = "us-east1"
    },

    "google.southcarolina.3" = {
      zone   = "us-east1-d"
      region = "us-east1"
    },

    "google.virginia.1" = {
      zone   = "us-east4-a"
      region = "us-east4"
    },

    "google.virginia.2" = {
      zone   = "us-east4-b"
      region = "us-east4"
    },

    "google.virginia.3" = {
      zone   = "us-east4-c"
      region = "us-east4"
    },

    "google.ohio.1" = {
      zone   = "us-east5-a"
      region = "us-east5"
    },

    "google.ohio.2" = {
      zone   = "us-east5-b"
      region = "us-east5"
    },

    "google.ohio.3" = {
      zone   = "us-east5-c"
      region = "us-east5"
    },

    "google.dallas.1" = {
      zone   = "us-south1-a"
      region = "us-south1"
    },

    "google.dallas.2" = {
      zone   = "us-south1-b"
      region = "us-south1"
    },

    "google.dallas.3" = {
      zone   = "us-south1-c"
      region = "us-south1"
    },

    "google.oregon.1" = {
      zone   = "us-west1-a"
      region = "us-west1"
    },

    "google.oregon.2" = {
      zone   = "us-west1-b"
      region = "us-west1"
    },

    "google.oregon.3" = {
      zone   = "us-west1-c"
      region = "us-west1"
    },

    "google.losangeles.1" = {
      zone   = "us-west2-a"
      region = "us-west2"
    },

    "google.losangeles.2" = {
      zone   = "us-west2-b"
      region = "us-west2"
    },

    "google.losangeles.3" = {
      zone   = "us-west2-c"
      region = "us-west2"
    },

    "google.saltlakecity.1" = {
      zone   = "us-west3-a"
      region = "us-west3"
    },

    "google.saltlakecity.2" = {
      zone   = "us-west3-b"
      region = "us-west3"
    },

    "google.saltlakecity.3" = {
      zone   = "us-west3-c"
      region = "us-west3"
    },

    "google.lasvegas.1" = {
      zone   = "us-west4-a"
      region = "us-west4"
    },

    "google.lasvegas.2" = {
      zone   = "us-west4-b"
      region = "us-west4"
    },

    "google.lasvegas.3" = {
      zone   = "us-west4-c"
      region = "us-west4"
    },

  }

}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "public" {
  for_each     = var.relays
  name         = "${replace(each.key, ".", "-")}-public"
  region       = local.datacenter_map[each.value.datacenter_name].region
  address_type = "EXTERNAL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_address" "internal" {
  for_each     = var.relays
  name         = "${replace(each.key, ".", "-")}-internal"
  region       = local.datacenter_map[each.value.datacenter_name].region
  address_type = "INTERNAL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_instance" "relay" {
  for_each     = var.relays
  name         = "${replace(each.key, ".", "-")}"
  zone         = local.datacenter_map[each.value.datacenter_name].zone
  machine_type = each.value.type
  network_interface {
    network_ip = google_compute_address.internal[each.key].address
    network    = "default"
    subnetwork = "default"
    access_config {
      nat_ip = google_compute_address.public[each.key].address
    }
  }
  boot_disk {
    initialize_params {
      image = each.value.image
    }
  }
  metadata = {
    ssh-keys = "ubuntu:${file(var.ssh_public_key_file)}"
  }
  lifecycle {
    create_before_destroy = true
  }
  metadata_startup_script = file("./setup_relay.sh")
}

# ----------------------------------------------------------------------------------------

output "relays" {

  description = "Data for each bare metal relay setup by Terraform"

  value = {
    for k, v in var.relays : k => zipmap( 
      [
        "relay_name", 
        "native_name",
        "datacenter_name",
        "supplier_name", 
        "public_address", 
        "internal_address", 
        "internal_group", 
        "ssh_address", 
        "ssh_user",
      ], 
      [
        k,
        local.datacenter_map[v.datacenter_name].zone,
        v.datacenter_name,
        "google", 
        "${google_compute_address.public[k].address}:40000",
        "${google_compute_address.internal[k].address}:40000",
        "", 
        "${google_compute_address.public[k].address}:22",
        "ubuntu",
      ]
    )
  }
}

# ----------------------------------------------------------------------------------------
