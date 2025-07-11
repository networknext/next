
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://sloclap_network_next_backend_artifacts"
google_database_bucket      = "gs://sloclap_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "6JMS4iG6mkY3w7AWzGGgN4H4ELYl2myiJgODXRhM8yo="

server_backend_public_key   = "vjhZOAGQQcLH/e+7m1OjB1lUp7noWxeP9rOHu5N3+yM="

load_test_buyer_public_key  = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ=="
load_test_buyer_private_key = "AzcqXbdP3TwX+9o9VfR7RcX2cq34UPdEsR2ztUnwxlTb/R49EiV5a2resciNEFLgF8bs6gpX08Bl8HStjtr4Ct2LMHMUB3a9"

ip2location_bucket_name     = "sloclap_network_next_staging"

relay_public_key  = "1nTj7bQmo8gfIDqG+o//GFsak/g1TRo4hl6XXw1JkyI="

relay_private_key = "cwvK44Pr5aHI3vE3siODS7CUgdPI/l1VwjVZ2FvEyAo="
