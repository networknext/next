
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://alocasia_network_next_backend_artifacts"
google_database_bucket      = "gs://alocasia_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "Lgk5XKY+3+fQR/yNMd6TbuZToY8RpWoL1e6c8vuIEyI="

server_backend_public_key   = "I3qjdX8b04RzfmYzXzJ9HBRCYpgD7Obqvh15Wp2zvow="

load_test_buyer_public_key  = "0/bHg4VPjOrB+Jp8kiyAkyPhrSnfOZi9jTNLFNbsTbS2e3MSeQdv7Q=="
load_test_buyer_private_key = "0/bHg4VPjOpyxYcBNBImCmapkW4Yj57Bott2ApCtORB6dfM8uv44RsH4mnySLICTI+GtKd85mL2NM0sU1uxNtLZ7cxJ5B2/t"

ip2location_bucket_name     = "alocasia_network_next_staging"

relay_public_key  = "JNH7qYX8mOPYy4enyN9ozjCL+0tCwACaiChfH3oP0Ek="
relay_private_key = "8qlMIoJNMxeLMJaj97E95vEAZhLRc6cmK/CtI3p3N7w="
