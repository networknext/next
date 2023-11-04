
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://test_network_next_terraform"
google_artifacts_bucket     = "gs://test_network_next_terraform"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "spacecats.net"

relay_backend_public_key    = "p5rFaSWJpOyVJ0MEO0NKF2vxuonpXj8WUjrmTiibA30="

server_backend_public_key   = "uuPPdQok9xNEgCzoRXQr9mlOXn9Lkoz0GvlCH55iI5k="

load_test_buyer_public_key  = "qfWdjLu3HjFti1XkneM4z+0+ZUDa03eb7iq/wA0G+Uu6heI5RhRDuQ=="
load_test_buyer_private_key = "qfWdjLu3HjEGCYdDGdBuUgY8IeF5W30Yom06m0Ko23IoTdjqG4RT2G2LVeSd4zjP7T5lQNrTd5vuKr/ADQb5S7qF4jlGFEO5"

ip2location_bucket_name     = "test_network_next_local"
