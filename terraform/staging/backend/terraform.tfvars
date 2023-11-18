
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

relay_backend_public_key    = "s9RsnStfXjDsWZHmuIaIiMPbJRs2UJ6BOO564PzXQWI="

server_backend_public_key   = "xNsPk3gxRc9zUQ7qBm+GfT8OeFcuWwxZ/NCC80hnd5s="

load_test_buyer_public_key  = "hc/OIfQ3E+IkueTZ8kW5Y0e/Y54luqQushYWjzyfYV5JxYc1RNI5lg=="
load_test_buyer_private_key = "hc/OIfQ3E+LvtzNBp5sUI8+xw0RfOLdC+4pd05EMH8aFubWS9ydoeiS55NnyRbljR79jniW6pC6yFhaPPJ9hXknFhzVE0jmW"

ip2location_bucket_name     = "alocasia_network_next_staging"

relay_public_key  = "YUU0dPo+w2Yt+L8Q5EYa6oC6Ml6tC6V6gCTzF4kQ81A="
relay_private_key = "QXpvbsevqS9qzmJ1FmmJwMiAc4fiIahVAveRzSbSQfs="
