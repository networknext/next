
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

relay_backend_public_key    = "9EH++ESxyB7cI21gtxnvwrOJIXea8J8D0qBju7riKUw="

server_backend_public_key   = "QrFX6TpX5PtaToPm5LFtBK0eBU4x9WREYtFqkHUeZc0="

load_test_buyer_public_key  = "WC6mKDnJmlHPsrvfii0WFJXExS3cVRJK+Hv1wWX1gRuISdNtuklkYw=="
load_test_buyer_private_key = "WC6mKDnJmlGOqyEK8CCxCE58KXEyeCCJMSJOP9K/Bh34tUl5WbwZ1s+yu9+KLRYUlcTFLdxVEkr4e/XBZfWBG4hJ0226SWRj"

ip2location_bucket_name     = "test_network_next_local"
