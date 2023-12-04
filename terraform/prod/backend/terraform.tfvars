
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

raspberry_buyer_public_key  = "/AUcbl5fuLxmPYEjtjBbVFnPJDlUuWrcntrVL5na6NsYzP2lsoOR5A=="
raspberry_buyer_private_key = "/AUcbl5fuLxR6BwDi1NUM41MAzeP8ZBYGNOY4FVBOqnA0nPn2Q77FGY9gSO2MFtUWc8kOVS5atye2tUvmdro2xjM/aWyg5Hk"

ip2location_bucket_name     = "newyork_network_next_prod"

relay_backend_public_key    = "NQiV7UaFevD4yA7xq6Qiy0nTQJ2WPr9g6AERXCYlln4="

server_backend_public_key   = "x9LEJTvjQ1kc+Vu9I8o6NmXyZbjCXnfs6upPcG+eEdU="

test_server_tag             = "001" # increment this when you need to redeploy the test server
