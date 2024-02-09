
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

raspberry_buyer_public_key  = "+fQWj43aTZfihh967fcy9hmsBHgC50QEmJ2i3gJUyxInsyy86llnLQ=="
raspberry_buyer_private_key = "+fQWj43aTZd6Nled4vr0EIB5prNBA4yDQBwu/pSiakVv/clsZg/3vOKGH3rt9zL2GawEeALnRASYnaLeAlTLEiezLLzqWWct"

ip2location_bucket_name     = "newyork_network_next_prod"

relay_backend_public_key    = "MzEsgGvPag3t+/ZA0NvwnUZdNedZGLySsm+baHPbAGY="

server_backend_public_key   = "g+U2x1juGaFJ2LzcQ6eyTCwJhSxQgR4AYgA27T/QslA="

test_server_tag             = "001" # increment this when you need to redeploy the test server
