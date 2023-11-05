
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://test_network_next_terraform"
google_artifacts_bucket     = "gs://test_network_next_terraform"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "spacecats.net"

raspberry_buyer_public_key  = "blxyyvzQSRYo6Nfu5l7Ah/RbUB2FAx+AiRzcXFa4fwkj8tYLS3f5+Q=="
raspberry_buyer_private_key = "blxyyvzQSRYfV/Gkq/cCiAGY6aL8LmIMeEFvg8F2sak9sJAaeHE0lCjo1+7mXsCH9FtQHYUDH4CJHNxcVrh/CSPy1gtLd/n5"

ip2location_bucket_name     = "test_network_next_local"

relay_backend_public_key    = "1ohrSutpU2m9loSsgw0T+A07pN2c356UKDG3B+V+tjg="

server_backend_public_key   = "h6y6F8i/BkmxbmeyDW7I4hencNkoEpQ8b+UWw5jac+I="
