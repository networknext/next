# ==========================================================================
#                               AMAZON CLOUD
# ==========================================================================

/*
    Unfortunately, it is LITERALLY IMPOSSIBLE to work programmatically in
    terraform across multiple regions in AWS :(

    To work around this we use code generation.

    The set of amazon dev relays in dev are defined at the top of this file:

      config/amazon.go

    Make your edits in there, then run the amazon configuration tool:

      "run config-amazon"

    This generates the following files:

      config/amazon.txt
      schemas/sql/sellers/amazon.sql,
      terraform/dev/relays/amazon/generated.tf

    IMPORTANT: You need to enable some regions and zones manually in your AWS account.

    For more details see:

      https://docs.aws.amazon.com/general/latest/gr/rande-manage.html

    and

      https://aws.amazon.com/about-aws/global-infrastructure/localzones/locations/
*/

# --------------------------------------------------------------------------

variable "config" { type = list(string) }
variable "credentials" { type = list(string) }
variable "profile" { type = string }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# --------------------------------------------------------------------------

output "datacenters" {
  description = "Data for each amazon datacenter"
  value = local.datacenter_map
}

# --------------------------------------------------------------------------
