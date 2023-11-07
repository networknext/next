
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://mindful_network_next_backend_artifacts"
google_database_bucket      = "gs://mindful_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "NdYOjRh3Z0f+T2u7V6uIsYgjtO7xW7kVQc8GvLY+Pwu9UJCPg30ZAQ=="
raspberry_buyer_private_key = "NdYOjRh3Z0dt1p5pJlNXHujIERNdHtJF/EtLarUn/oLvSB0lt65mCP5Pa7tXq4ixiCO07vFbuRVBzwa8tj4/C71QkI+DfRkB"

ip2location_bucket_name     = "mindful_network_next_prod"

relay_backend_public_key    = "U4JzQ3p3EWa5kDkHg0Oac9vYBeGW1Ac/H0SZyfeSAGo="

server_backend_public_key   = "7fhJnEKpyZs60OGioZoUyy90D6xSdiEnHbA0OpRrukE="
