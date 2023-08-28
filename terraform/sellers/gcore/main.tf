# ----------------------------------------------------------------------------------------

terraform {
  required_version = ">= 0.13.0"
  required_providers {
    gcore = {
      source  = "G-Core/gcorelabs"
      version = "0.3.18"
    }
  }
}

provider gcore {
  permanent_api_token = file("~/secrets/terraform-gcore.txt")
}

# ----------------------------------------------------------------------------------------

variable "relays" { type = list(map(string)) }
variable "project" { type = string }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# ----------------------------------------------------------------------------------------

/*
resource "gcore_keypair" "relay" {
  sshkey_name = "relay"
  project_id = var.project
  public_key = file(var.ssh_public_key_file)
}
*/

# ----------------------------------------------------------------------------------------
