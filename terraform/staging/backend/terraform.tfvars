
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://auspicious_network_next_terraform"
google_artifacts_bucket     = "gs://auspicious_network_next_terraform"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "spacecats.net"

relay_backend_public_key    = "ZUJYHyNI/ljYI2Xx0XQM33p16QqjnGxrtN10B/4Q7Hs="

server_backend_public_key   = "fR9dJ5gwMfsBI3k+3wvTGI97dBYMhryQk4rplx+HFiE="

load_test_buyer_public_key  = "N/d4A0szNUg42dQ8VLUhy6GvBW5e/84T1ncLB86TAww37+0ZBZ6b/g=="
load_test_buyer_private_key = "N/d4A0szNUgB2PHwkWdKC8jdw7mzNjo0YRA7Tkqk5W8lF/m57ygHLjjZ1DxUtSHLoa8Fbl7/zhPWdwsHzpMDDDfv7RkFnpv+"

ip2location_bucket_name     = "auspicious_network_next_local"
