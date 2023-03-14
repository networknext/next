# ==========================================================================
#                               AMAZON CLOUD
# ==========================================================================

/*
    Before deploying relays in AWS, run the amazon configuration tool:

      "run amazon-config"

    It generates: 

      config/amazon.txt
      schemas/sql/sellers/amazon.sql,
      terraform/amazon.tfvars
      terraform/suppliers/amazon/providers.tf
      terraform/suppliers/amazon/region/main.tf

    Some AWS regions need to be manually enabled on your account.

    For more details on enabling regions see: 

      https://docs.aws.amazon.com/general/latest/gr/rande-manage.html
*/

# --------------------------------------------------------------------------

variable "config" { type = list(string) }
variable "credentials" { type = list(string) }
variable "profile" { type = string }
variable "relays" { type = map(map(string)) }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }
variable "datacenter_map" { type = map(map(string)) }
variable "regions" { type = list(string) }

# --------------------------------------------------------------------------

module "region" {
  source = "./region"
  providers = {
    aws = aws.us-east-1
  }
  vpn_address = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
}

# --------------------------------------------------------------------------

# per-relay

# security_group_id = aws_security_group.allow_ssh_and_udp.id

# --------------------------------------------------------------------------

output "relays" {

  description = "Data for each amazon setup by Terraform"

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
        "amazon", 
        "127.0.0.1:40000",
        "127.0.0.1:40000",
        var.datacenter_map[v.datacenter_name].region,
        "127.0.0.1:22",
        "ubuntu",
      ]
    )
  }
}

# --------------------------------------------------------------------------
