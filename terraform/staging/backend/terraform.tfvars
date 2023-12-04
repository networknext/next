
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://newyork_network_next_backend_artifacts"
google_database_bucket      = "gs://newyork_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "WGpbJQHZ3cVGn9hAAcIUlsGCFQPuFWJZmWxlgOwEkBc="

server_backend_public_key   = "Rm2XKpNDjqtEhdYqclNweGUOGWpeNM3D6v4Q8V85EIA="

load_test_buyer_public_key  = "kLWeaPkL+EYZJDBguhajXE1V5yj5q2WY3I0ITQNN6TELp2J39hOQrA=="
load_test_buyer_private_key = "kLWeaPkL+Ea3Nrmwy30aYExru7gmp5/m5HWDY+M50do9cKO6stcsHRkkMGC6FqNcTVXnKPmrZZjcjQhNA03pMQunYnf2E5Cs"

ip2location_bucket_name     = "newyork_network_next_staging"

relay_public_key  = "a7LY0GIAmEdg6ntY1qTM8ke0p+EXmIYivtbK3cFaqA8="
relay_private_key = "6Wo1FMSCkWwAC87rNlpwIfWuk8t3XuLS5S6ayAv4cb4="
