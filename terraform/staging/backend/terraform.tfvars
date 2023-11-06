
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

relay_backend_public_key    = "YFQmIrt7CuH++jeDBewujNk9n3SUowisv0DeQabSWBE="

server_backend_public_key   = "McWWyZVbwqTkinwz3wktcEWsUm78dVUSZaiWcZ5JImE="

load_test_buyer_public_key  = "YdayThVTwjK5klktRiH1CX5dkvENGG29ObRaGbjsdwxMZh0awtB8uw=="
load_test_buyer_private_key = "YdayThVTwjJpL1fN7C2VEszvrMjkrvZR5K7SA8LagXuXuCeyseY/zbmSWS1GIfUJfl2S8Q0Ybb05tFoZuOx3DExmHRrC0Hy7"

ip2location_bucket_name     = "mindful_network_next_staging"

relay_public_key  = "6M3UU558N5GP7nEueVvDCGPyXXrZohJK2USI+Dv25T4="
relay_private_key = "TO4B9fx9x8X6Knwbh3pusFhzOhKFv6hX3ANoOU+bWi8="
