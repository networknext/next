# ----------------------------------------------------------------------------------------

/*
    Do not setup and try out this provider as part of a dev enviroment, or you purchase a monthly commit for a VM server each time you recreate the relays :)

    Plans are bare metal and available on a monthly basis. This provider currently does not do VMs.

    Available sites as of March 13, 2023:

      BGT    -> latitude.bogota
      ASH    -> latitude.virginia
      CH1    -> latitude.chicago
      BUE    -> latitude.buenosaires
      LON    -> latitude.london
      MEX    -> latitude.mexico.1
      MEX2   -> latitude.mexico.2
      TY6    -> latitude.tokyo.1
      TY8    -> latitude.tokyo.2
      SAN    -> latitude.santiago.1
      SAN2   -> latitude.santiago.2
      SYD    -> latitude.sydney
      NY2    -> latitude.newyork
      MI1    -> latitude.miami
      LA2    -> latitude.losangeles
      DAL2   -> latitude.dallas
      MH1    -> latitude.saopaulo.1
      SP2    -> latitude.saopaulo.2

  To update this list, run:

    curl --request GET \
     --url https://api.latitude.sh/regions \
     --header 'accept: application/json' \
     -H "Authorization: Bearer <API_KEY>"

  Available plans as of March 13, 2023:

    c1-tiny-x86
    c1-medium-x86
    c1-large-x86
    c2-small-x86
    c2-medium-x86
    c2-large-x86
    c3-medium-x86
    c3-large-x86
    c3-small-x86
    m3-large-x86
    s3-large-x86

    etc...

  Not all plans are available in each location. Plans are really just an inventory of different types of bare metal servers available in each location.

  To update this list, run:

    curl --request GET \
     --url https://api.latitude.sh/plans \
     --header 'accept: application/json' \
     -H "Authorization: Bearer <API_KEY>"

*/

# --------------------------------------------------------------------------

terraform {
  required_providers {
    latitudesh = {
      source  = "latitudesh/latitudesh"
      version = "~> 0.2.1"
    }
  }
}

provider "latitudesh" {
  auth_token = file("~/secrets/terraform-latitude.txt")
}

# ----------------------------------------------------------------------------------------

variable "project_name" { type = string }
variable "project_description" { type = string }
variable "project_environment" { type = string }
variable "relays" { type = list(map(string)) }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# ----------------------------------------------------------------------------------------

resource "latitudesh_project" "relay" {
  name        = var.project_name
  description = var.project_description
  environment = var.project_environment
}

resource "latitudesh_ssh_key" "relay" {
  name       = "relay"
  project    = latitudesh_project.relay.id
  public_key = file(var.ssh_public_key_file)
}

resource "latitudesh_user_data" "relay" {
  description = "Setup relay"
  project = latitudesh_project.relay.id
  content = base64encode(replace(file("./setup_relay.sh"), "$VPN_ADDRESS", var.vpn_address))
}

resource "latitudesh_server" "relay" {
  count            = length(var.relays)
  hostname         = var.relays[count.index].name
  site             = var.relays[count.index].site
  operating_system = var.relays[count.index].os
  plan             = var.relays[count.index].plan
  project          = latitudesh_project.relay.id
  ssh_keys         = [latitudesh_ssh_key.relay.id]
  user_data        = latitudesh_user_data.relay.id
}

output "relays" {
  description = "Data for each latitude relay setup by Terraform"
  value = [for i, v in var.relays : zipmap(["relay_name", "public_address", "internal_address"], [var.relays[i].name, latitudesh_server.relay[i].primary_ip_v4, "0.0.0.0"])]
}

# ----------------------------------------------------------------------------------------
