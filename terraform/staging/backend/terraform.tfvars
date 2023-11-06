
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://mindful_network_next_backend_artifacts"
google_database_bucket      = "gs://mindful_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "e8yrTIUsW5dlQf5J5zvzDVQ9z9JA5dhuSaIU2SANQmw="

server_backend_public_key   = "M2A5b3FTOcQw57NYxEABxDbBQao7yafCPLJZ0SCQ87o="

load_test_buyer_public_key  = "9bR0O5cFm8VPNU+tb5neR83LW8g/uLMBmHn/tNuUFUvCE5FB8daA5w=="
load_test_buyer_private_key = "9bR0O5cFm8U42EN+ewsdvGksKcUxrGN41oCOX2GPA5wOeYgI4idsZE81T61vmd5HzctbyD+4swGYef+025QVS8ITkUHx1oDn"

ip2location_bucket_name     = "mindful_network_next_staging"

relay_public_key  = "bs6r40PQCSz0f2wgLSJ/Kv2Y2ouR94CJnnO7g5LZ8j4="
relay_private_key = "1YOQ3O+SrwdfmxxBxvqSTIpb+/Y5gzoqGRkvvFduYVQ="
