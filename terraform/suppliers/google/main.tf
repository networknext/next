# ========================================================================================
#                                     GOOGLE CLOUD
# ========================================================================================

/*
    Google cloud is an awesome supplier with highly performant relays.

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

    See google.txt for more details, or schemas/sql/google.sql

    You can get an updated version of this list from: https://cloud.google.com/compute/docs/regions-zones
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

  # IMPORTANT: This map translates from the network next google datacenter names to google cloud zone and region
  # If you add any more datacenters from google cloud, then this map must be updated before you can use them.
  # You must also update update schemas/sql/google.sql then patch your database, as well as update config/google.txt
  # and deploy the config to your environment before the SDK will pick up the new datacenters in autodetect.

  datacenter_map = {

    google.taiwan.1 = {
      zone   = "asia-east1-a"
      region = "asia-east1"
    },

    google.taiwan.2 = {
      zone   = "asia-east1-b"
      region = "asia-east1"
    },

    google.taiwan.3 = {
      zone   = "asia-east1-c"
      region = "asia-east1"
    },

    google.hongkong.1 = {
      zone   = "asia-east2-a"
      region = "asia-east2"
    },

    google.hongkong.2 = {
      zone   = "asia-east2-b"
      region = "asia-east2"
    },

    google.hongkong.3 = {
      zone   = "asia-east2-c"
      region = "asia-east2"
    },

    google.tokyo.1 = {
      zone   = "asia-northeast1-a"
      region = "asia-northeast1"
    },

    google.tokyo.2 = {
      zone   = "asia-northeast1-b"
      region = "asia-northeast1"
    },

    google.tokyo.3 = {
      zone   = "asia-northeast1-c"
      region = "asia-northeast1"
    }

    google.osaka.1 = {
      zone   = "asia-northeast2-a"
      region = "asia-northeast2"
    },

    google.osaka.2 = {
      zone   = "asia-northeast2-b"
      region = "asia-northeast2"
    },

    google.osaka.3 = {
      zone   = "asia-northeast2-c"
      region = "asia-northeast2"
    },

    google.seoul.1 = {
      zone   = "asia-northeast3-a"
      region = "asia-northeast3"
    },

    google.seoul.2 = {
      zone   = "asia-northeast3-b"
      region = "asia-northeast3"
    },

    google.seoul.3 = {
      zone   = "asia-northeast3-c"
      region = "asia-northeast3"
    },

    google.mumbai.1 = {
      zone   = "asia-south1-a"
      region = "asia-south1"
    },

    google.mumbai.2 = {
      zone   = "asia-south1-b"
      region = "asia-south1"
    },

    google.mumbai.3 = {
      zone   = "asia-south1-c"
      region = "asia-south1"
    },

    google.delhi.1 = {
      zone   = "asia-south2-a"
      region = "asia-south2"
    },

    google.delhi.2 = {
      zone   = "asia-south2-b"
      region = "asia-south2"
    },

    google.delhi.3 = {
      zone   = "asia-south2-c"
      region = "asia-south2"
    },

    google.singapore.1 = {
      zone   = "asia-southeast1-a"
      region = "asia-southeast1"
    },

    google.singapore.2 = {
      zone   = "asia-southeast1-b"
      region = "asia-southeast1"
    },

    google.singapore.3 = {
      zone   = "asia-southeast1-c"
      region = "asia-southeast1"
    },

    google.jakarta.1 = {
      zone   = "asia-southeast2-a"
      region = "asia-southeast2"
    },

    google.jakarta.2 = {
      zone   = "asia-southeast2-b"
      region = "asia-southeast2"
    },

    google.jakarta.3 = {
      zone   = "asia-southeast2-c"
      region = "asia-southeast2"
    },

    google.sydney.1 = {
      zone   = "australia-southeast1-a"
      region = "australia-southeast1"
    },

    google.sydney.2 = {
      zone   = "australia-southeast1-b"
      region = "australia-southeast1"
    },

    google.sydney.3 = {
      zone   = "australia-southeast1-c"
      region = "australia-southeast1"
    },

    google.melbourne.1 = {
      zone   = "australia-southeast2-a"
      region = "australia-southeast2"
    },

    google.melbourne.2 = {
      zone   = "australia-southeast2-b"
      region = "australia-southeast2"
    },

    google.melbourne.3 = {
      zone   = "australia-southeast2-c"
      region = "australia-southeast2"
    },

  }

}

/*
europe-central2-a,google.warsaw.1
europe-central2-b,google.warsaw.2
europe-central2-c,google.warsaw.3
europe-north1-a,google.finland.1
europe-north1-b,google.finland.2
europe-north1-c,google.finland.3
europe-southwest1-a,google.madrid.1,
europe-southwest1-b,google.madrid.2,
europe-southwest1-c,google.madrid.3,
europe-west1-b,google.belgium.1
europe-west1-c,google.belgium.2
europe-west1-d,google.belgium.3
europe-west2-a,google.london.1
europe-west2-b,google.london.2
europe-west2-c,google.london.3
europe-west3-a,google.frankfurt.1
europe-west3-b,google.frankfurt.2
europe-west3-c,google.frankfurt.3
europe-west4-a,google.netherlands.1
europe-west4-b,google.netherlands.2
europe-west4-c,google.netherlands.3
europe-west6-a,google.zurich.1
europe-west6-b,google.zurich.2
europe-west6-c,google.zurich.3
europe-west8-a,google.milan.1
europe-west8-b,google.milan.2
europe-west8-c,google.milan.3
europe-west9-a,google.paris.1
europe-west9-b,google.paris.2
europe-west9-c,google.paris.3
me-west1-a,google.telaviv.1
me-west1-b,google.telaviv.2
me-west1-c,google.telaviv.3
northamerica-northeast1-a,google.montreal.1
northamerica-northeast1-b,google.montreal.2
northamerica-northeast1-c,google.montreal.3
northamerica-northeast2-a,google.toronto.1
northamerica-northeast2-b,google.toronto.2
northamerica-northeast2-c,google.toronto.3
southamerica-east1-a,google.saopaulo.1
southamerica-east1-b,google.saopaulo.2
southamerica-east1-c,google.saopaulo.3
southamerica-west1-a,google.santiago.1
southamerica-west1-b,google.santiago.2
southamerica-west1-c,google.santiago.3
us-central1-a,google.iowa.1
us-central1-b,google.iowa.2
us-central1-c,google.iowa.3
us-central1-f,google.iowa.4
us-east1-b,google.southcarolina.1
us-east1-c,google.southcarolina.2
us-east1-d,google.southcarolina.3
us-east4-a,google.virginia.1
us-east4-b,google.virginia.2
us-east4-c,google.virginia.3
us-east5-a,google.ohio.1
us-east5-b,google.ohio.2
us-east5-c,google.ohio.3
us-south1-a,google.dallas.1
us-south1-b,google.dallas.2
us-south1-c,google.dallas.3
us-west1-a,google.oregon.1
us-west1-b,google.oregon.2
us-west1-c,google.oregon.3
us-west2-a,google.losangeles.1
us-west2-b,google.losangeles.2
us-west2-c,google.losangeles.3
us-west3-a,google.saltlakecity.1
us-west3-b,google.saltlakecity.2
us-west3-c,google.saltlakecity.3
us-west4-a,google.lasvegas.1
us-west4-b,google.lasvegas.2
us-west4-c,google.lasvegas.3
*/

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
