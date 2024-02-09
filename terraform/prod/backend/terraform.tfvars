
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://newyork_network_next_backend_artifacts"
google_database_bucket      = "gs://newyork_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "52M6enRvBROk/IX86wOdyZRCBkJMRiqlj6nEA/klgCAdsPW1Eqwv7Q=="
raspberry_buyer_private_key = "52M6enRvBRNaR5FIuSrO8b4QiCHQZxAkBUcSpCBighyet/qH3AW3ZKT8hfzrA53JlEIGQkxGKqWPqcQD+SWAIB2w9bUSrC/t"

ip2location_bucket_name     = "newyork_network_next_prod"

relay_backend_public_key    = "io936eLvF+jXGwthG5fTNX/eluIbgvZ76ZT+GH3j+D8="

server_backend_public_key   = "uLpTRU7CVeoWFceGobX284EgmvNE5pha/yrjQx2v5wI="

test_server_tag             = "001" # increment this when you need to redeploy the test server
