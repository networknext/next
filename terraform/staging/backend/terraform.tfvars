
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://next_network_next_backend_artifacts"
google_database_bucket      = "gs://next_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "GwcGxScFxwc6pipXazWOQlz3aTVaNseThiNYtIyd10k="

server_backend_public_key   = "P3/d3jmPGQof1jjmF7/1aBd6ytd3bJ8rBp4qdkPlr/M="

load_test_buyer_public_key  = "OPsJ/biQrnQEgoJr2oo9zeJG9vVkOUpWklw2+O2nfyy1BljyFxrU8Q=="
load_test_buyer_private_key = "OPsJ/biQrnQWRDrCHrOYPpYR/aRkRJA3IhJKx1ZZu95p59UokTB6/gSCgmvaij3N4kb29WQ5SlaSXDb47ad/LLUGWPIXGtTx"

ip2location_bucket_name     = "next_network_next_staging"

relay_public_key  = "peLF27fnP8pXz6AqgH6SM7s90iCOgEI+2rjGrACgGCU="

relay_private_key = "ACQytjHVJca67Tp5RFCe9f/IKEwQLCxjr8xSymqu09E="
