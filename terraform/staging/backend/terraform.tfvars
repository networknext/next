
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

relay_backend_public_key    = "ICqts7NXBFB63DYjKVzkLd7Jyz7aN1DeaVai1jP2bGU="

server_backend_public_key   = "om19C1IDP8bL7arIRaJ0r+YEaoXjRODuFBF/X1xxR/k="

load_test_buyer_public_key  = "WU+Fzt1uQN5Fj6iQ9MU2hm/i/deZRscQtD7DymIvdEpj8RIhxhfcXg=="
load_test_buyer_private_key = "WU+Fzt1uQN63cn1xN7pL2ZVpvsWQxYRoUut5IjfH/vxzi/p4j9l4iUWPqJD0xTaGb+L915lGxxC0PsPKYi90SmPxEiHGF9xe"

ip2location_bucket_name     = "test_network_next_local"
