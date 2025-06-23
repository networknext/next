
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://next_network_next_backend_artifacts"
google_database_bucket      = "gs://next_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

test_buyer_public_key       = "OPsJ/biQrnQEgoJr2oo9zeJG9vVkOUpWklw2+O2nfyy1BljyFxrU8Q=="
test_buyer_private_key      = "OPsJ/biQrnQWRDrCHrOYPpYR/aRkRJA3IhJKx1ZZu95p59UokTB6/gSCgmvaij3N4kb29WQ5SlaSXDb47ad/LLUGWPIXGtTx"

raspberry_region            = "us-central1"
raspberry_zones             = ["us-central1-a"]
raspberry_buyer_public_key  = "YJHQ5FGeoveQMdSzLRmbOKFRVY6QeMyUX1c4kRA72anucqnPRBr8IA=="
raspberry_buyer_private_key = "YJHQ5FGeovcEX0r2r9P1LH+8tnnA3oQ7RSV6f/r0ihkbC2MOLVVbsJAx1LMtGZs4oVFVjpB4zJRfVziREDvZqe5yqc9EGvwg"

ip2location_bucket_name     = "next_network_next_prod"

relay_backend_public_key    = "unH/Yxm0C6JCZ1dTGZH2BTBOFhGMcYsOEDURd9qY72w="

server_backend_public_key   = "Uycn3KibCfXJo1uM+NNWgCySRzM2Ti3bhvom9XBkxfE="

test_server_region          = "us-central1"
test_server_zone            = "us-central1-a"
test_server_tag             = "001" # increment this each time you want to deploy the test server

disable_backend             = false
disable_raspberry           = false
disable_ip2location         = true
