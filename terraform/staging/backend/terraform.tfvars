
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

relay_backend_public_key    = "oST/G6zwdIrdMK1dJxhelrrK1K5zpEr108HnBA/19Sg="

server_backend_public_key   = "PwVfQrUfc0fq6bKvUrJ+7JgvjSm5d2k8L2uoShobHOo="

load_test_buyer_public_key  = "n7u8H8u17JkwfCTkyJ9W1nHzkxjYotCoErwIdpfyAU9uqMgV4SJEjQ=="
load_test_buyer_private_key = "n7u8H8u17JkUE5QCcC6Qmi1Wgs4ATkawgxwNX98q2YbzTNZojCJnXTB8JOTIn1bWcfOTGNii0KgSvAh2l/IBT26oyBXhIkSN"

ip2location_bucket_name     = "newyork_network_next_staging"

relay_public_key  = "DpXODs3Asg2rPvHU6+7GHNzUBbD9wqwIcvp7tc3gRlU="
relay_private_key = "FD7bDX16pyHmQPGNaQCsEz+mLo0Q1GLVqSLsE+DsmRM="
