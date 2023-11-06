
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

raspberry_buyer_public_key  = "emtKwwJhDkpGotL1Wxg4M1d4EU7DtjOwLYpd3uWjng+hZF4TLI3TkA=="
raspberry_buyer_private_key = "emtKwwJhDkqd+znEScKN3wMntPFOUzcG8e7w8A/pKQvBgmhY8gQUt0ai0vVbGDgzV3gRTsO2M7Atil3e5aOeD6FkXhMsjdOQ"

ip2location_bucket_name     = "mindful_network_next_prod"

relay_backend_public_key    = "4rreqvfvfwqBMwjp8XrtudIg467LPoLbTjXADH/Ed1M="

server_backend_public_key   = "Y/OGHBlzR5FcEsOx00dBr8g8B0UQg48ghENfiGdNECY="
