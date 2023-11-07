
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://mindful_network_next_backend_artifacts"
google_database_bucket      = "gs://mindful_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "uee7+ElivuPCl9PcRu6cZ/wJFeeQ75IFBdzwSlxOq2g="

server_backend_public_key   = "DYG4nzvI1M48qplMHvkv5/ONuYNnKs6QH/RrOUBtphA="

load_test_buyer_public_key  = "iqzfU7267tSb9vmCQIQJR2nZ06HuQcn57fxaxT/fquTLMvFBqLJshA=="
load_test_buyer_private_key = "iqzfU7267tQhG7wRljx4jMU998LhA0lxSMj+quRpUleR+BSlNN+rjJv2+YJAhAlHadnToe5Byfnt/FrFP9+q5Msy8UGosmyE"

ip2location_bucket_name     = "mindful_network_next_staging"

relay_public_key  = "smlC9vXuLQepLM1KYMIfYBczEWbvX5jItU/Kj+RK3Gw="
relay_private_key = "QD1t8t1wEBOTOhQXzuxuHi9ight+2gZHslwPMOokBak="
