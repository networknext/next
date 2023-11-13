
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://alocasia_network_next_backend_artifacts"
google_database_bucket      = "gs://alocasia_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "d57/wLdTcI/S0XqCqsxNt2iLeFaszI5DkCy9/0fWIsOoaVPt141SWA=="
raspberry_buyer_private_key = "d57/wLdTcI9RSzgbcnZwhsXgvpxuzc5DuCwZtXPREd4UGR1a1nywC9LReoKqzE23aIt4VqzMjkOQLL3/R9Yiw6hpU+3XjVJY"

ip2location_bucket_name     = "alocasia_network_next_prod"

relay_backend_public_key    = "8ZCxIJpt8Y65NbinaIj/vyDB+lGV5BOn9TNcKTo0GhQ="

server_backend_public_key   = "5ygH1wHVdWrO+uezE+HPlPScPtvNivfV5JHqwuXYGyc="
