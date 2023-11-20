
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://dogfood_network_next_backend_artifacts"
google_database_bucket      = "gs://dogfood_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "IH5qMORR3p6E3gncDLtjMnoc//52ksTe1kbFG2npsio="

server_backend_public_key   = "9eLrvRIbKH8AtVy0OUKQqFzH4sLU5r0geSJidVxMMxU="

load_test_buyer_public_key  = "Lw4WO3aMP6VKgpMFYilu2KRcVwqyoMHodt3QFK+FXpR+c0+DGvITog=="
load_test_buyer_private_key = "Lw4WO3aMP6WkuuvjIi+pcY6D/qC8rwCU3AjSxHYRgbWGyneNvRlA/EqCkwViKW7YpFxXCrKgweh23dAUr4VelH5zT4Ma8hOi"

ip2location_bucket_name     = "dogfood_network_next_staging"

relay_public_key  = "/uUJOkyWefo3mjqseLDlexsPH9S9pD13OXrD8u6nIVw="
relay_private_key = "nqRXN42HoyZaPTrcCz69bu2BmjBheCpEDm21RPsW7CE="
