# ==========================================================================
#                               AMAZON CLOUD
# ==========================================================================

/*
    Unfortunately, it is LITERALLY IMPOSSIBLE to work programmatically in
    terraform across multiple regions in AWS :(

    To work around this, the set of amazon relays are specified in:

      tools/amazon_config/amazon_config.go

    And they are written to the "generated.tf" file in this directory.

    To get started with amazon relays, look in the amazon_config.go file

    Make your edits in there...

    Then run the amazon configuration tool:

      "run amazon-config"

    The config tool generates the following files:

      config/amazon.txt
      schemas/sql/sellers/amazon.sql,
      terraform/dev/relays/amazon/generated.tf
      terraform/staging/relays/amazon/generated.tf
      terraform/prod/relays/amazon/generated.tf

    Also, some regions need to be manually enabled on your account.

    For more details on enabling regions see: 

      https://docs.aws.amazon.com/general/latest/gr/rande-manage.html

    There are also local zones in AWS which must be manually enabled here:

      https://aws.amazon.com/about-aws/global-infrastructure/localzones/locations/
*/

# --------------------------------------------------------------------------

variable "config" { type = list(string) }
variable "credentials" { type = list(string) }
variable "profile" { type = string }
variable "ssh_public_key_file" { type = string }
variable "vpn_address" { type = string }

# --------------------------------------------------------------------------
