
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://memento_network_next_backend_artifacts"
google_database_bucket      = "gs://memento_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "BsLPRMwxGJfGvuxGSQQ4SKOy88JqbSTFGxPkJJkIQWaa5h5+W6xXqg=="
raspberry_buyer_private_key = "BsLPRMwxGJeCVSYT7UaFiBKAHsO+eQw6O0dr0nxdlclKHBOc2f5Keca+7EZJBDhIo7LzwmptJMUbE+QkmQhBZprmHn5brFeq"

ip2location_bucket_name     = "memento_network_next_prod"

relay_backend_public_key    = "56MmkT6UdHiI+dN+dBmljQuY7Nk0I+McGvEFnVLVZWE="

server_backend_public_key   = "5ECTj46rH9JAI0GSJb6H4/x3fpmIBLffQGraglr/KV8="
