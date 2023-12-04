
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://newyork_network_next_backend_artifacts"
google_database_bucket      = "gs://newyork_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "Oe96rqXLyoxaBc3ce8jo6UeoSNlWpq11Druv8kEO9bP0C2rQW5phkg=="
raspberry_buyer_private_key = "Oe96rqXLyozzVR7TFcwEj/9hcuIcq+RQxHC+Jjhs2p9vJZVp7MDCGloFzdx7yOjpR6hI2VamrXUOu6/yQQ71s/QLatBbmmGS"

ip2location_bucket_name     = "newyork_network_next_prod"

relay_backend_public_key    = "z28h4XSf7MsUFVdMtLMITDxZX2Y6+1oyKZjoRkQsXjc="

server_backend_public_key   = "tWi9O0YZuXIfYwbxx78yPISVCx6utFpiHpeeHCVN9Xs="

test_server_tag             = "001" # increment this when you need to redeploy the test server
