
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://wilton_network_next_backend_artifacts"
google_database_bucket      = "gs://wilton_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "ZkQ8nC3BTkFjR7ws8GhRz/zVUueQXcde1T0EL+j0/XPVh39S6IoVaw=="
raspberry_buyer_private_key = "ZkQ8nC3BTkGK7DxXGLIPwIx00Jm3ep25n1OmwTZWqYJrKfLj2XmXVmNHvCzwaFHP/NVS55Bdx17VPQQv6PT9c9WHf1LoihVr"

ip2location_bucket_name     = "wilton_network_next_prod"

relay_backend_public_key    = "jRUKe8S7x5s6Se537gkjpIHq4JWKyEd3MoQ7iQOqoTQ="

server_backend_public_key   = "//VvaQBqYPq1Qto6nct6PZe2BINccgeHDbohlyN5vYk="

test_server_tag             = "001" # increment this when you need to redeploy the test server
