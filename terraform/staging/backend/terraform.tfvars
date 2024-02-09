
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

relay_backend_public_key    = "BPtFFeLDzu/Am0PZ3kaJvk/9B5eGk3rd16EGrnm8jmg="

server_backend_public_key   = "SVKPWusPqa4KR5ll/EpnApPNZELsgcx7J/MKXHqvJcY="

load_test_buyer_public_key  = "9fuymsQpqwyyimT9iSSXJi4Dnf3XHM6XlYq0kIPOawhLHjN7TwDXAQ=="
load_test_buyer_private_key = "9fuymsQpqwxryaio6c8tvgI8YrsIh1Er4Z+oOxCbpTwBqfx1D2qZprKKZP2JJJcmLgOd/dcczpeVirSQg85rCEseM3tPANcB"

ip2location_bucket_name     = "newyork_network_next_staging"

relay_public_key  = "G4dLznxXv9DRvsuarADx5LRpG8h0iZ+YlAKv/vYhMGI="
relay_private_key = "FR9RAy+xo/Idvg9GWCOiP06TQHpn7zPhwM+TjPQYS1A="
