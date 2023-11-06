
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://auspicious_network_next_backend_artifacts"
google_database_bucket      = "gs://auspicious_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "bEXPecXWQSn5ciiHrX3lHT+J6/MxJjmoimOP/PktT3oW1Cd88Vnm8w=="
raspberry_buyer_private_key = "bEXPecXWQSkYMyWse2Aj+fhbxuz2kpwUJGZjH6kJ7HL08zEqcEDhm/lyKIetfeUdP4nr8zEmOaiKY4/8+S1PehbUJ3zxWebz"

ip2location_bucket_name     = "auspicious_network_next_prod"

relay_backend_public_key    = "SL1l3HwQMeVf8C4FNmi1ne+piEUfii27uMES2jqPWFY="

server_backend_public_key   = "KZOUnBWHFcUGmqTGU5AhgYL2cRA+c5G/Khai9hpb+i8="
