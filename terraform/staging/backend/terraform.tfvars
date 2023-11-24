
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://wilton_network_next_backend_artifacts"
google_database_bucket      = "gs://wilton_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "qauO0L0Z8ACLvROGdgTTThYgLnGzYuYdPU3dYUsCHng="

server_backend_public_key   = "OK1lLHe129zXAPQdEyl6kyguFncVlwiJByFun09qYxI="

load_test_buyer_public_key  = "fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA=="
load_test_buyer_private_key = "fJ9R1DqVKetUVM2jP6hcMjDTkKlOwFbrtgeCUrpAh3epnaO07HBkm+t6D6S+oSQWpsABrnhzoFdyAGf5HYYozUaQWjxC2RRk"

ip2location_bucket_name     = "wilton_network_next_staging"

relay_public_key  = "ayyX2+oaE4FJjoEHGnAWTQ6EeO829If64UEcshgm6xA="
relay_private_key = "I7cfCSX8Kq62YeFaSd4CTpNwCVr+VlQxcb9+wUukXpk="
