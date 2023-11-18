
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://alocasia_network_next_backend_artifacts"
google_database_bucket      = "gs://alocasia_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "bZcUhYFYkrLqvbvB1eYpyrcZBH6Qfr/WMZExUnu4wo5AUhXuZQQxmQ=="
raspberry_buyer_private_key = "bZcUhYFYkrK5HxET5Cy0F2qtbNuITF4Q4temJ/A54A6xf1ylXcrrNOq9u8HV5inKtxkEfpB+v9YxkTFSe7jCjkBSFe5lBDGZ"

ip2location_bucket_name     = "alocasia_network_next_prod"

relay_backend_public_key    = "I+7H0XwZ4liC5yEdPQv8LYCOL+Y57+qfKKOG2iqJYCY="

server_backend_public_key   = "jot/6lToP2sX71lepqHSkqgrsjOElCJWSyqbIfKsGvk="
