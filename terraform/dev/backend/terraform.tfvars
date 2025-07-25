
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-dev.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"]
google_artifacts_bucket     = "gs://next_network_next_backend_artifacts"
google_database_bucket      = "gs://next_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

test_buyer_public_key       = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ=="
test_buyer_private_key      = "AzcqXbdP3TwX+9o9VfR7RcX2cq34UPdEsR2ztUnwxlTb/R49EiV5a2resciNEFLgF8bs6gpX08Bl8HStjtr4Ct2LMHMUB3a9"

raspberry_region            = "us-central1"
raspberry_zones             = ["us-central1-a"]
raspberry_buyer_public_key  = "gtdzp3hCfJ9Y+6OOpsWoMChMXhXGDRnY7vkFdHwNqVW0bdp6jjTx6Q=="
raspberry_buyer_private_key = "gtdzp3hCfJ+Xl4L4PsLbaBlzLeIogMkmzArY3r19jSenj1t4TAQKGlj7o46mxagwKExeFcYNGdju+QV0fA2pVbRt2nqONPHp"

ip2location_bucket_name     = "next_network_next_dev"

relay_backend_public_key    = "Z+9puZkCkV03nm4yO49ySF+H181jAlWVy7JPGMlk10I="

server_backend_public_key   = "0dvRVqU+krtetlEosEdPN+IxVsNqi7/+Hi6gVjSwSl0="

test_server_region          = "us-central1"
test_server_zone            = "us-central1-a"
test_server_tag             = "001" # increment this each time you want to deploy the test server

disable_backend             = true
disable_raspberry           = true
disable_ip2location         = true
