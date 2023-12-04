
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-dev.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"]
google_artifacts_bucket     = "gs://newyork_network_next_backend_artifacts"
google_database_bucket      = "gs://newyork_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

test_buyer_public_key  = "9qzGNONKAHTBaPsm+b9pPUgEvekv3iKZBdXJt7eSBePFkeWtoxpGig=="
test_buyer_private_key = "9qzGNONKAHTmQiE2vRhmIL7cB+GB2v0HrRNozHefaNCTHNow8WNqj8Fo+yb5v2k9SAS96S/eIpkF1cm3t5IF48WR5a2jGkaK"

raspberry_buyer_public_key  = "/AUcbl5fuLxmPYEjtjBbVFnPJDlUuWrcntrVL5na6NsYzP2lsoOR5A=="
raspberry_buyer_private_key = "/AUcbl5fuLxR6BwDi1NUM41MAzeP8ZBYGNOY4FVBOqnA0nPn2Q77FGY9gSO2MFtUWc8kOVS5atye2tUvmdro2xjM/aWyg5Hk"

ip2location_bucket_name     = "newyork_network_next_dev"

relay_backend_public_key    = "bKjCNngZ1H+XJppN6MymZ9UoTgewgOsLeAMAOiiWuws="

server_backend_public_key   = "mUHMKszMnQxeN9oC/etGMSpUEcIg7hJCupWaxLLLl2Y="

test_server_tag             = "006" # increment this each time you want to deploy the test server
