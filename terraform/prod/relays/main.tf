# ========================================================================================
#                                      PROD RELAYS
# ========================================================================================

locals {
  
  env                         = "prod"
  vpn_address                 = "45.79.157.168"
  ssh_public_key_file         = "~/secrets/next_ssh.pub"
  ssh_private_key_file        = "~/secrets/next_ssh"
  relay_version               = "relay-144"
  relay_artifacts_bucket      = "sloclap_network_next_relay_artifacts"
  relay_backend_public_key    = "TINP/TnYY/0W7JvLFlYGrB0MUw+b4aIrN20Vq7g5bhU="
  relay_backend_url           = "relay.virtualgo.net"

  raspberry_buyer_public_key  = "gtdzp3hCfJ9Y+6OOpsWoMChMXhXGDRnY7vkFdHwNqVW0bdp6jjTx6Q=="

  raspberry_datacenters = [
    "google.saopaulo.1",
    "google.saopaulo.2",
    "google.saopaulo.3",
  ]

  test_buyer_public_key       = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ=="

  test_datacenters = [
    "google.saopaulo.1",
    "google.saopaulo.2",
    "google.saopaulo.3",
  ]

  rematch_buyer_public_key    = "+FJSkN2kAua9KdP3/83gSkmIXARxcoB1vFKA6JAXsWf/0Syno+6T1A=="

  rematch_datacenters = [

    "latitude.saopaulo",
    "google.saopaulo.1",
    "google.saopaulo.2",
    "google.saopaulo.3",

    "i3d.dubai",
    "google.doha.1",
    "google.doha.2",
    "google.doha.3",

    "zenlayer.istanbul",
    "datapacket.istanbul",
    "gcore.istanbul",

    "velia.frankfurt",
    "uk2group.frankfurt",
    "gcore.frankfurt",
    "i3d.frankfurt",
    "datapacket.frankfurt",
    "ovh.frankfurt",
    "google.frankfurt.1",
    "google.frankfurt.2",
    "google.frankfurt.3",
    "amazon.frankfurt.1",
    "amazon.frankfurt.2",
    "amazon.frankfurt.3",

    "i3d.losangeles",
    "datapacket.losangeles",
    "hivelocity.losangeles",
    "google.losangeles.1",
    "google.losangeles.2",
    "google.losangeles.3",

    "uk2group.dallas",
    "serversdotcom.dallas",
    "google.dallas.1",
    "google.dallas.2",
    "google.dallas.3",

    "latitude.ashburn",
    "i3d.ashburn",
    "gcore.ashburn",
    "ovh.ashburn",
    "datapacket.ashburn",
    "google.virginia.1",
    "google.virginia.2",
    "google.virginia.3",

    "i3d.singapore",

    "google.tokyo.1",
    "google.tokyo.2",
    "google.tokyo.3",

    "serversaustralia.sydney",
    "google.sydney.1",
    "google.sydney.2",
    "google.sydney.3",

    "gcore.johannesburg",
    "google.johannesburg.1",
    "google.johannesburg.2",
    "google.johannesburg.3",
  ]

  sellers = {
    "Akamai" = "akamai"
    "Amazon" = "amazon"
    "Google" = "google"
    "Datapacket" = "datapacket"
    "i3D" = "i3d"
    "Oneqode" = "oneqode"
    "GCore" = "gcore"
    "Hivelocity" = "hivelocity"
    "ColoCrossing" = "colocrossing"
    "phoenixNAP" = "phoenixnap"
    "servers.com" = "serversdotcom"
    "Velia" = "velia"
    "Zenlayer" = "zenlayer"
    "Latitude" = "latitude"
    "Equinix" = "equinix"
    "Azure" = "azure"
    "UK2Group" = "uk2group"
    "OVH" = "ovh"
    "ServersAustralia" = "serversaustralia"
  }
}

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    networknext = {
      source = "networknext/networknext"
      version = "~> 5.0.14"
    }
  }
  backend "gcs" {
    bucket  = "sloclap_network_next_terraform"
    prefix  = "prod_relays"
  }
}

provider "networknext" {
  hostname = "https://api.virtualgo.net"
  api_key  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwicG9ydGFsIjp0cnVlLCJpc3MiOiJuZXh0IGtleWdlbiIsImlhdCI6MTc0OTczODE4OX0._3gzvR5D7mIILHXnujbGkqwK0jiDCUUXRRzTpjJv-gc"
}

# ----------------------------------------------------------------------------------------

# =============
# GOOGLE RELAYS
# =============

locals {

  google_credentials = "~/secrets/terraform-prod-relays.json"
  google_project     = file("~/secrets/prod-relays-project-id.txt")
  google_relays = {

    "google.lasvegas.1" = {
      datacenter_name = "google.lasvegas.1"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.saltlakecity.1" = {
      datacenter_name = "google.saltlakecity.1"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.oregon.1" = {
      datacenter_name = "google.oregon.1"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.toronto.1" = {
      datacenter_name = "google.toronto.1"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.toronto.2" = {
      datacenter_name = "google.toronto.2"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.toronto.3" = {
      datacenter_name = "google.toronto.3"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.montreal.1" = {
      datacenter_name = "google.montreal.1"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.montreal.2" = {
      datacenter_name = "google.montreal.2"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.montreal.3" = {
      datacenter_name = "google.montreal.3"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.ohio.1" = {
      datacenter_name = "google.ohio.1"      
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.ohio.2" = {
      datacenter_name = "google.ohio.2"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.ohio.3" = {
      datacenter_name = "google.ohio.3"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.1" = {
      datacenter_name = "google.iowa.1"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.2" = {
      datacenter_name = "google.iowa.2"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.3" = {
      datacenter_name = "google.iowa.3"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.saopaulo.1" = {
      datacenter_name = "google.saopaulo.1"
      type            = "n2-highcpu-2" # "c2-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.saopaulo.2" = {
      datacenter_name = "google.saopaulo.2"
      type            = "n2-highcpu-2" # c2-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.saopaulo.3" = {
      datacenter_name = "google.saopaulo.3"
      type            = "n2-highcpu-2" # "c2-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.santiago.1" = {
      datacenter_name = "google.santiago.1"
      type            = "n2-highcpu-2" # "c2-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.santiago.2" = {
      datacenter_name = "google.santiago.2"
      type            = "c2-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.santiago.3" = {
      datacenter_name = "google.santiago.3"
      type            = "c2-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

    "google.frankfurt.1" = {
      datacenter_name = "google.frankfurt.1"
      type            = "n2-highcpu-2" # "c2-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.frankfurt.2" = {
      datacenter_name = "google.frankfurt.2"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.frankfurt.3" = {
      datacenter_name = "google.frankfurt.3"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.dallas.1" = {
      datacenter_name = "google.dallas.1"
      type            = "n2-highcpu-2" # "c3-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.dallas.2" = {
      datacenter_name = "google.dallas.2"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.dallas.3" = {
      datacenter_name = "google.dallas.3"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

  /*
    "google.losangeles.1" = {
      datacenter_name = "google.losangeles.1"
      type            = "n2-highcpu-2" # "c2-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
  */

    "google.losangeles.2" = {
      datacenter_name = "google.losangeles.2"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.losangeles.3" = {
      datacenter_name = "google.losangeles.3"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.virginia.1" = {
      datacenter_name = "google.virginia.1"
      type            = "n2-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.virginia.2" = {
      datacenter_name = "google.virginia.2"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.virginia.3" = {
      datacenter_name = "google.virginia.3"
      type            = "n2-highcpu-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.queretaro.1" = {
      datacenter_name = "google.queretaro.1"
      type            = "n4-highcpu-2" # "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.queretaro.2" = {
      datacenter_name = "google.queretaro.2"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.queretaro.3" = {
      datacenter_name = "google.queretaro.3"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

    "google.doha.1" = {
      datacenter_name = "google.doha.1"
      type            = "n2-highcpu-2" # "c3-highmem-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.doha.2" = {
      datacenter_name = "google.doha.2"
      type            = "n2-highcpu-2" # c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.doha.3" = {
      datacenter_name = "google.doha.3"
      type            = "n2-highcpu-2" # "e2-standard-8"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.dammam.1" = {
      datacenter_name = "google.dammam.1"
      type            = "n2-highcpu-2" # "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.dammam.2" = {
      datacenter_name = "google.dammam.2"
      type            = "c2-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.dammam.3" = {
      datacenter_name = "google.dammam.3"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

    "google.telaviv.1" = {
      datacenter_name = "google.telaviv.1"
      type            = "n2-highcpu-2" # "c4-highmem-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.telaviv.2" = {
      datacenter_name = "google.telaviv.2"
      type            = "c3-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.telaviv.3" = {
      datacenter_name = "google.telaviv.3"
      type            = "c3-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

    "google.paris.1" = {
      datacenter_name = "google.paris.1"
      type            = "n2-highcpu-2" # "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.paris.2" = {
      datacenter_name = "google.paris.2"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.paris.3" = {
      datacenter_name = "google.paris.3"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

    "google.netherlands.1" = {
      datacenter_name = "google.netherlands.1"
      type            = "n2-highcpu-2" # c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.netherlands.2" = {
      datacenter_name = "google.netherlands.2"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.netherlands.3" = {
      datacenter_name = "google.netherlands.2"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

    "google.milan.2" = {
      datacenter_name = "google.milan.2"
      type            = "n2-highcpu-2" # c4-standard-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.milan.3" = {
      datacenter_name = "google.milan.3"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

    "google.belgium.2" = {
      datacenter_name = "google.belgium.2"
      type            = "n2-highcpu-2" # "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.belgium.3" = {
      datacenter_name = "google.belgium.3"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.belgium.4" = {
      datacenter_name = "google.belgium.4"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

    "google.london.1" = {
      datacenter_name = "google.london.1"
      type            = "n2-highcpu-2" # "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.london.2" = {
      datacenter_name = "google.london.2"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.london.3" = {
      datacenter_name = "google.london.3"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

    "google.madrid.1" = {
      datacenter_name = "google.madrid.1"
      type            = "n2-highcpu-2" # c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    /*
    "google.madrid.2" = {
      datacenter_name = "google.madrid.2"
      type            = "c4-highcpu-4"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
    */

  }
}

module "google_relays" {
  env                 = "prod"
  relays              = local.google_relays
  project             = local.google_project
  credentials         = local.google_credentials
  source              = "../../sellers/google"
  vpn_address         = local.vpn_address
  ssh_public_key_file         = "~/secrets/next_ssh.pub"
}

# ----------------------------------------------------------------------------------------

# =============
# AMAZON RELAYS
# =============

locals {
  amazon_config      = ["~/.aws/config"]
  amazon_credentials = ["~/.aws/credentials"]
  amazon_profile     = "default"
}

module "amazon_relays" {

  # IMPORTANT: It is LITERALLY IMPOSSIBLE to work with multiple AWS regions programmatically in Terraform
  # So for AWS, see sellers/amazon.go for the set of prod relays -> amazon/generated.tf

  config              = local.amazon_config
  credentials         = local.amazon_credentials
  profile             = local.amazon_profile
  source              = "./amazon"
  vpn_address         = local.vpn_address
  ssh_public_key_file         = "~/secrets/next_ssh.pub"
}

# ----------------------------------------------------------------------------------------

# =============
# AKAMAI RELAYS
# =============

locals {

  akamai_relays = {

    "akamai.chicago" = {
      datacenter_name = "akamai.chicago"
      type            = "g7-premium-16"
      image           = "linode/ubuntu22.04"
    },

    "akamai.saopaulo" = {
      datacenter_name = "akamai.saopaulo"
      type            = "g7-premium-16"
      image           = "linode/ubuntu22.04"
    },

    /*
    "akamai.frankfurt.1" = {
      datacenter_name = "akamai.frankfurt.1"
      type            = "g6-dedicated-16"
      image           = "linode/ubuntu22.04"
    },
    */

    "akamai.frankfurt.2" = {
      datacenter_name = "akamai.frankfurt.2"
      type            = "g7-premium-16"
      image           = "linode/ubuntu22.04"
    },

    "akamai.dallas" = {
      datacenter_name = "akamai.dallas"
      type            = "g6-dedicated-16"
      image           = "linode/ubuntu22.04"
    },

    "akamai.washington" = {
      datacenter_name = "akamai.washington"
      type            = "g7-premium-16"
      image           = "linode/ubuntu22.04"
    },

    "akamai.losangeles" = {
      datacenter_name = "akamai.losangeles"
      type            = "g7-premium-16"
      image           = "linode/ubuntu22.04"
    },    

    /*
    "akamai.london" = {
      datacenter_name = "akamai.london"
      type            = "g7-premium-16"
      image           = "linode/ubuntu22.04"
    },
    */
  }
}

module "akamai_relays" {
  env                 = "prod"
  relays              = local.akamai_relays
  source              = "../../sellers/akamai"
  vpn_address         = local.vpn_address
  ssh_public_key_file         = "~/secrets/next_ssh.pub"
}

# ----------------------------------------------------------------------------------------

# ===============
# ZENLAYER RELAYS
# ===============

locals {

  zenlayer_relays = {

    "zenlayer.saopaulo" = {
      datacenter_name = "zenlayer.saopaulo"
      public_address  = "128.14.222.42"
    },

    "zenlayer.bogota" = {
      datacenter_name  = "zenlayer.bogota"
      public_address   = "107.151.192.238"
      internal_address = "10.131.48.2"
    },

    "zenlayer.lima" = {
      datacenter_name = "zenlayer.lima"
      public_address   = "198.44.168.157"
      internal_address = "10.131.64.2"
    },

    "zenlayer.riyadh" = {
      datacenter_name  = "zenlayer.riyadh"
      public_address   = "162.128.75.95"
      internal_address = "10.130.112.2"
    },

    "zenlayer.frankfurt" = {
      datacenter_name  = "zenlayer.frankfurt"
      public_address   = "193.118.46.254"
    },

    "zenlayer.frankfurt" = {
      datacenter_name  = "zenlayer.frankfurt"
      public_address   = "193.118.46.254"
    },

    "zenlayer.london.a" = {
      datacenter_name  = "zenlayer.london"
      public_address   = "98.98.142.118"
    },

    "zenlayer.istanbul" = {
      datacenter_name  = "zenlayer.istanbul"
      public_address   = "104.166.176.246"
    },

    "zenlayer.dubai" = {
      datacenter_name  = "zenlayer.dubai"
      public_address   = "193.118.56.2"
    },

    "zenlayer.ashburn" = {
      datacenter_name  = "zenlayer.ashburn"
      public_address   = "128.14.84.78"
    },

    "zenlayer.miami" = {
      datacenter_name  = "zenlayer.miami"
      public_address   = "128.14.216.30"
    },

    "zenlayer.dallas" = {
      datacenter_name  = "zenlayer.dallas"
      public_address   = "98.96.193.162"
    },

    "zenlayer.buenosaires" = {
      datacenter_name  = "zenlayer.buenosaires"
      public_address   = "98.98.173.226"
    },

    "zenlayer.losangeles.a" = {
      datacenter_name  = "zenlayer.losangeles"
      public_address   = "128.14.73.78"
    },

    "zenlayer.losangeles.b" = {
      datacenter_name  = "zenlayer.losangeles"
      public_address   = "128.14.73.82"
    },

  }
}

module "zenlayer_relays" {
  relays = local.zenlayer_relays
  source = "../../sellers/zenlayer"
}

# ----------------------------------------------------------------------------------------

# ============
# AZURE RELAYS
# ============

locals {

  azure_relays = {

    // ...
    
  }
}

module "azure_relays" {
  relays = local.azure_relays
  source = "../../sellers/azure"
}

# ----------------------------------------------------------------------------------------

# =================
# DATAPACKET RELAYS
# =================

locals {

  datapacket_relays = {

    "datapacket.saopaulo" = {
      datacenter_name = "datapacket.saopaulo"
      ssh_address = "79.127.137.166"
      public_address = "79.127.137.166"
    },

    "datapacket.miami" = {
      datacenter_name = "datapacket.miami"
      ssh_address = "152.233.22.15"
      public_address = "152.233.22.15"
    },

    "datapacket.losangeles" = {
      datacenter_name = "datapacket.losangeles"
      ssh_address = "79.127.232.205"
      public_address = "79.127.232.205"
    }

    "datapacket.dallas" = {
      datacenter_name = "datapacket.dallas"
      ssh_address = "212.102.40.186"
      public_address = "212.102.40.186"
    }

    "datapacket.istanbul" = {
      datacenter_name = "datapacket.istanbul"
      ssh_address = "169.150.215.57"
      public_address = "169.150.215.57"
    }

    "datapacket.ashburn" = {
      datacenter_name = "datapacket.ashburn"
      ssh_address = "79.127.223.38"
      public_address = "79.127.223.38"
    }

    "datapacket.frankfurt.a" = {
      datacenter_name = "datapacket.frankfurt"
      ssh_address = "89.222.124.57"
      public_address = "89.222.124.57"
    }

    "datapacket.frankfurt.b" = {
      datacenter_name = "datapacket.frankfurt"
      ssh_address = "89.222.124.31"
      public_address = "89.222.124.31"
    }

    "datapacket.santiago" = {
      datacenter_name = "datapacket.santiago"
      ssh_address = "79.127.209.147"
      public_address = "79.127.209.147"
    }

    "datapacket.lima" = {
      datacenter_name = "datapacket.lima"
      ssh_address = "79.127.252.114"
      public_address = "79.127.252.114"
    }

    "datapacket.bogota" = {
      datacenter_name = "datapacket.bogota"
      ssh_address = "169.150.228.9"
      public_address = "169.150.228.9"
    }

    "datapacket.london" = {
      datacenter_name = "datapacket.london"
      ssh_address = "138.199.51.243"
      public_address = "138.199.51.243"
    }

  }
}

module "datapacket_relays" {
  relays = local.datapacket_relays
  source = "../../sellers/datapacket"
}

# ----------------------------------------------------------------------------------------

# ==========
# I3D RELAYS
# ==========

locals {

  i3d_relays = {

    "i3d.saopaulo" = {
      datacenter_name = "i3d.saopaulo"
      public_address  = "185.50.104.109"
    },

    "i3d.ashburn.a" = {
      datacenter_name = "i3d.ashburn"
      public_address  = "162.244.55.38"
    },

    "i3d.ashburn.b" = {
      datacenter_name = "i3d.ashburn"
      public_address  = "162.244.55.42"
    },

    "i3d.dubai.a" = {
      datacenter_name = "i3d.dubai"
      public_address  = "185.179.202.102"
    },

    "i3d.dubai.b" = {
      datacenter_name = "i3d.dubai"
      public_address  = "185.179.202.104"
    },

    "i3d.frankfurt" = {
      datacenter_name = "i3d.frankfurt"
      public_address  = "188.122.68.101"
    },

    "i3d.losangeles" = {
      datacenter_name = "i3d.losangeles"
      public_address  = "162.245.204.237"
    },

    "i3d.dallas" = {
      datacenter_name = "i3d.dallas"
      public_address  = "138.128.136.133"
    },

  }
}

module "i3d_relays" {
  relays = local.i3d_relays
  source = "../../sellers/i3d"
}

# ----------------------------------------------------------------------------------------

# ===============
# LATITUDE RELAYS
# ===============

locals {

  latitude_relays = {

    "latitude.chicago" = {
      datacenter_name = "latitude.chicago"
      public_address  = "186.233.187.217"
    },

    "latitude.saopaulo.a" = {
      datacenter_name = "latitude.saopaulo"
      public_address  = "189.1.173.223"
    },

    "latitude.saopaulo.b" = {
      datacenter_name = "latitude.saopaulo"
      public_address  = "103.88.235.93"
    },

    "latitude.saopaulo.c" = {
      datacenter_name = "latitude.saopaulo"
      public_address  = "103.88.235.133"
    },

    "latitude.saopaulo.d" = {
      datacenter_name = "latitude.saopaulo"
      public_address  = "177.54.147.193"
    },

    "latitude.ashburn" = {
      datacenter_name = "latitude.ashburn"
      public_address  = "103.106.59.205"
    },

    "latitude.miami" = {
      datacenter_name = "latitude.miami"
      public_address  = "69.67.150.47"
    },

    "latitude.losangeles" = {
      datacenter_name = "latitude.losangeles"
      public_address  = "67.213.124.239"
    },

    "latitude.santiago" = {
      datacenter_name = "latitude.santiago"
      public_address  = "45.250.252.203"
    },

    "latitude.london" = {
      datacenter_name = "latitude.london"
      public_address  = "103.50.32.156"
    },

    "latitude.mexico" = {
      datacenter_name = "latitude.mexico"
      public_address  = "103.88.234.141"
    }

    "latitude.frankfurt" = {
      datacenter_name = "latitude.frankfurt"
      public_address  = "189.1.171.153"
    }

  }
}

module "latitude_relays" {
  relays = local.latitude_relays
  source = "../../sellers/latitude"
}

# ----------------------------------------------------------------------------------------

# ============
# GCORE RELAYS
# ============

locals {

  gcore_relays = {

    "gcore.saopaulo" = {
      datacenter_name = "gcore.saopaulo"
      public_address  = "92.38.150.8"
      ssh_user        = "root"
    },

    "gcore.frankfurt" = {
      datacenter_name = "gcore.frankfurt"
      public_address  = "93.114.56.87"
      ssh_user        = "ubuntu"
    },

    "gcore.istanbul.a" = {
      datacenter_name = "gcore.istanbul"
      public_address  = "213.156.152.90"
      ssh_user        = "ubuntu"
    },

    "gcore.istanbul.b" = {
      datacenter_name = "gcore.istanbul"
      public_address  = "213.156.152.107"
      ssh_user        = "ubuntu"    
    },

    "gcore.ashburn" = {
      datacenter_name = "gcore.ashburn"
      public_address  = "5.188.124.74"
      ssh_user        = "root"
    },

  }
}

module "gcore_relays" {
  relays = local.gcore_relays
  source = "../../sellers/gcore"
}

# ----------------------------------------------------------------------------------------



























# ----------------------------------------------------------------------------------------

# ==============
# ONEQODE RELAYS
# ==============

locals {

  oneqode_relays = {

    /*
    "oneqode.singapore" = {
      datacenter_name = "oneqode.singapore"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "oneqode_relays" {
  relays = local.oneqode_relays
  source = "../../sellers/oneqode"
}

# ----------------------------------------------------------------------------------------

# ===============
# UK2GROUP RELAYS
# ===============

locals {

  uk2group_relays = {

    "uk2group.frankfurt.a" = {
      datacenter_name = "uk2group.frankfurt"
      public_address  = "46.23.74.129"
    },

    "uk2group.frankfurt.b" = {
      datacenter_name = "uk2group.frankfurt"
      public_address  = "46.23.75.213"
    },

    "uk2group.frankfurt.c" = {
      datacenter_name = "uk2group.frankfurt"
      public_address  = "46.23.75.215"
    },

    "uk2group.dallas.a" = {
      datacenter_name = "uk2group.dallas"
      public_address  = "206.217.211.37"
    },

    "uk2group.dallas.b" = {
      datacenter_name = "uk2group.dallas"
      public_address  = "206.217.211.40"
    },

    "uk2group.dallas.c" = {
      datacenter_name = "uk2group.dallas"
      public_address  = "206.217.211.48"
    },
  }
}

module "uk2group_relays" {
  relays = local.uk2group_relays
  source = "../../sellers/uk2group"
}

# ----------------------------------------------------------------------------------------

# ==========
# OVH RELAYS
# ==========

locals {

  ovh_relays = {

    "ovh.ashburn" = {
      datacenter_name = "ovh.ashburn"
      public_address  = "40.160.32.142"
    },
    
    "ovh.frankfurt" = {
      datacenter_name = "ovh.frankfurt"
      public_address  = "51.89.11.27"
    },

  }
}

module "ovh_relays" {
  relays = local.ovh_relays
  source = "../../sellers/ovh"
}

# ----------------------------------------------------------------------------------------

# =================
# HIVELOCITY RELAYS
# =================

locals {

  hivelocity_relays = {

    "hivelocity.losangeles" = {
      datacenter_name = "hivelocity.losangeles"
      public_address  = "107.155.127.114"
    },

    "hivelocity.ashburn" = {
      datacenter_name = "hivelocity.ashburn"
      public_address  = "91.191.213.138"
    },

    "hivelocity.miami" = {
      datacenter_name = "hivelocity.miami"
      public_address  = "162.254.151.130"
    },

  }
}

module "hivelocity_relays" {
  relays = local.hivelocity_relays
  source = "../../sellers/hivelocity"
}

# ----------------------------------------------------------------------------------------

# ===================
# COLOCROSSING RELAYS
# ===================

locals {

  colocrossing_relays = {

    /*
    "colocrossing.chicago" = {
      datacenter_name = "colocrossing.chicago"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "colocrossing_relays" {
  relays = local.colocrossing_relays
  source = "../../sellers/colocrossing"
}

# ----------------------------------------------------------------------------------------

# =================
# PHOENIXNAP RELAYS
# =================

locals {

  phoenixnap_relays = {

    /*
    "phoenixnap.phoenix" = {
      datacenter_name = "phoenixnap.phoenix"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "phoenixnap_relays" {
  relays = local.phoenixnap_relays
  source = "../../sellers/phoenixnap"
}

# ----------------------------------------------------------------------------------------

# ==================
# SERVERS.COM RELAYS
# ==================

locals {

  serversdotcom_relays = {

    "serversdotcom.dallas" = {
      datacenter_name = "serversdotcom.dallas"
      public_address  = "64.58.117.12"
    },

  }
}

module "serversdotcom_relays" {
  relays = local.serversdotcom_relays
  source = "../../sellers/serversdotcom"
}

# ----------------------------------------------------------------------------------------

# =======================
# SERVER AUSTRALIA RELAYS
# =======================

locals {

  serversaustralia_relays = {

    /*
    "serversaustralia.sydney" = {
      datacenter_name = "serversaustralia.sydney"
      public_address  = ""
    },
    */

  }
}

module "serversaustralia_relays" {
  relays = local.serversaustralia_relays
  source = "../../sellers/serversaustralia"
}

# ----------------------------------------------------------------------------------------

# ============
# VELIA RELAYS
# ============

locals {

  velia_relays = {

    "velia.frankfurt.a" = {
      datacenter_name = "velia.frankfurt"
      public_address  = "37.61.208.81"
    },

    "velia.frankfurt.b" = {
      datacenter_name = "velia.frankfurt"
      public_address  = "37.61.218.237" 
    },

    "velia.frankfurt.c" = {
      datacenter_name = "velia.frankfurt"
      public_address  = "37.61.219.29"
    },

  }
}

module "velia_relays" {
  relays = local.velia_relays
  source = "../../sellers/velia"
}

# ----------------------------------------------------------------------------------------

# ==============
# EQUINIX RELAYS
# ==============

locals {

  equinix_relays = {

    /*
    "equinix.miami" = {
      datacenter_name = "equinix.miami"
      public_address  = "185.152.67.2"
    },
    */

  }
}

module "equinix_relays" {
  relays = local.equinix_relays
  source = "../../sellers/equinix"
}

# ----------------------------------------------------------------------------------------

# ========================
# INITIALIZE PROD DATABASE
# ========================

# Setup sellers, datacenters and relays in prod

locals {
  
  relay_names = sort(
    concat(
      keys(module.google_relays.relays),
      keys(module.amazon_relays.relays),
      keys(module.akamai_relays.relays),
      keys(module.datapacket_relays.relays),
      keys(module.i3d_relays.relays),
      keys(module.oneqode_relays.relays),
      keys(module.gcore_relays.relays),
      keys(module.hivelocity_relays.relays),
      keys(module.colocrossing_relays.relays),
      keys(module.phoenixnap_relays.relays),
      keys(module.serversdotcom_relays.relays),
      keys(module.velia_relays.relays),
      keys(module.zenlayer_relays.relays),
      keys(module.latitude_relays.relays),
      keys(module.equinix_relays.relays),
      keys(module.azure_relays.relays),
      keys(module.uk2group_relays.relays),
      keys(module.ovh_relays.relays),
      keys(module.serversaustralia_relays.relays),
    )
  )

  relays = merge(
    module.google_relays.relays,
    module.amazon_relays.relays,
    module.akamai_relays.relays,
    module.datapacket_relays.relays,
    module.i3d_relays.relays,
    module.oneqode_relays.relays,
    module.gcore_relays.relays,
    module.hivelocity_relays.relays,
    module.colocrossing_relays.relays,
    module.phoenixnap_relays.relays,
    module.serversdotcom_relays.relays,
    module.velia_relays.relays,
    module.zenlayer_relays.relays,
    module.latitude_relays.relays,
    module.equinix_relays.relays,
    module.azure_relays.relays,
    module.uk2group_relays.relays,
    module.ovh_relays.relays,
    module.serversaustralia_relays.relays,
  )

  datacenters = merge(
    module.google_relays.datacenters,
    module.amazon_relays.datacenters,
    module.akamai_relays.datacenters,
    module.datapacket_relays.datacenters,
    module.i3d_relays.datacenters,
    module.oneqode_relays.datacenters,
    module.gcore_relays.datacenters,
    module.hivelocity_relays.datacenters,
    module.colocrossing_relays.datacenters,
    module.phoenixnap_relays.datacenters,
    module.serversdotcom_relays.datacenters,
    module.velia_relays.datacenters,
    module.zenlayer_relays.datacenters,
    module.latitude_relays.datacenters,
    module.equinix_relays.datacenters,
    module.azure_relays.datacenters,
    module.uk2group_relays.datacenters,
    module.ovh_relays.datacenters,
    module.serversaustralia_relays.datacenters,
  )

  datacenter_names = distinct([for k, relay in local.relays : relay.datacenter_name])
}

resource "networknext_seller" sellers {
  for_each = local.sellers
  name     = each.key
  code     = each.value
}

locals {
  seller_map = {
    for seller in networknext_seller.sellers: 
      seller.code => seller
  }
}

resource "networknext_datacenter" datacenters {
  for_each = local.datacenters
  name = each.key
  seller_id = local.seller_map[each.value.seller_code].id
  latitude = each.value.latitude
  longitude = each.value.longitude
  native_name = each.value.native_name
}

locals {
  datacenter_map = {
    for datacenter in networknext_datacenter.datacenters:
      datacenter.name => datacenter
  }
}

resource "networknext_relay_keypair" relay_keypairs {
  for_each = local.relays
}

resource "networknext_relay" relays {
  for_each = local.relays
  name = each.key
  datacenter_id = local.datacenter_map[each.value.datacenter_name].id
  public_key_base64=networknext_relay_keypair.relay_keypairs[each.key].public_key_base64
  private_key_base64=networknext_relay_keypair.relay_keypairs[each.key].private_key_base64
  public_ip = each.value.public_ip
  public_port = each.value.public_port
  internal_ip = each.value.internal_ip
  internal_port = each.value.internal_port
  internal_group = each.value.internal_group
  ssh_ip = each.value.ssh_ip
  ssh_port = each.value.ssh_port
  ssh_user = each.value.ssh_user
  version = local.relay_version
  bandwidth_price = each.value.bandwidth_price
}

# ----------------------------------------------------------------------------------------

# Print out set of relays in the database

data "networknext_relays" relays {
  depends_on = [
    networknext_relay.relays,
  ]
}

locals {
  database_relays = {
    for k,v in networknext_relay.relays: k => v.public_ip
  }
}

output "database_relays" {
  value = local.database_relays
}

output "all_relays" {
  value = networknext_relay.relays
}

# ----------------------------------------------------------------------------------------

/*
# ===============
# RASPBERRY BUYER
# ===============

resource "networknext_route_shader" raspberry {
  name = "raspberry"
  force_next = true
  route_select_threshold = 300
  route_switch_threshold = 300
}

resource "networknext_buyer" raspberry {
  name = "Raspberry"
  code = "raspberry"
  debug = false
  live = true
  route_shader_id = networknext_route_shader.raspberry.id
  public_key_base64 = local.raspberry_buyer_public_key
}

resource "networknext_buyer_datacenter_settings" raspberry {
  for_each = toset(local.raspberry_datacenters)
  buyer_id = networknext_buyer.raspberry.id
  datacenter_id = networknext_datacenter.datacenters[each.value].id
  enable_acceleration = true
}

# ==========
# TEST BUYER
# ==========

resource "networknext_route_shader" test {
  name = "test"
  force_next = true
  latency_reduction_threshold = 1
  acceptable_latency = 0
  acceptable_packet_loss = 100
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
  route_select_threshold = 1
  route_switch_threshold = 10
}

resource "networknext_buyer" test {
  name = "Test"
  code = "test"
  debug = false
  live = true
  route_shader_id = networknext_route_shader.test.id
  public_key_base64 = local.test_buyer_public_key
}

resource "networknext_buyer_datacenter_settings" test {
  for_each = toset(local.test_datacenters)
  buyer_id = networknext_buyer.test.id
  datacenter_id = networknext_datacenter.datacenters[each.value].id
  enable_acceleration = true
}
*/

# =============
# REMATCH BUYER
# =============

resource "networknext_route_shader" rematch {
  name = "rematch"
  force_next = false
  latency_reduction_threshold = 10
  acceptable_latency = 20
  acceptable_packet_loss = 1.0
  bandwidth_envelope_up_kbps = 1024
  bandwidth_envelope_down_kbps = 1024
  route_select_threshold = 1
  route_switch_threshold = 5
}

resource "networknext_buyer" rematch {
  name = "REMATCH"
  code = "rematch"
  debug = false
  live = true
  route_shader_id = networknext_route_shader.rematch.id
  public_key_base64 = local.rematch_buyer_public_key
}

resource "networknext_buyer_datacenter_settings" rematch {
  for_each = toset(local.rematch_datacenters)
  buyer_id = networknext_buyer.rematch.id
  datacenter_id = networknext_datacenter.datacenters[each.value].id
  enable_acceleration = true
}

# ----------------------------------------------------------------------------------------
