
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://mindful_network_next_backend_artifacts"
google_database_bucket      = "gs://mindful_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "h5E1IwxPLHjrJKXwnAY2J5MLbmLtm0+3SOirrUMZrhDC6KxN+fQQMw=="
raspberry_buyer_private_key = "h5E1IwxPLHh3KHkYz8eYs+YgtZ+ntKi8TzlofjANE4xFv1TLUf+jjuskpfCcBjYnkwtuYu2bT7dI6KutQxmuEMLorE359BAz"

ip2location_bucket_name     = "mindful_network_next_prod"

relay_backend_public_key    = "s14UkUu+hW5tDBSMPdk+y3CUaGZsocRZL0PR78RYmhE="

server_backend_public_key   = "ADYty/Hkx9w9T0c3jWkMuzyM6ttNU3oldw1CTWynwGc="
