
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

test_buyer_public_key  = "9fuymsQpqwyyimT9iSSXJi4Dnf3XHM6XlYq0kIPOawhLHjN7TwDXAQ=="
test_buyer_private_key = "9fuymsQpqwxryaio6c8tvgI8YrsIh1Er4Z+oOxCbpTwBqfx1D2qZprKKZP2JJJcmLgOd/dcczpeVirSQg85rCEseM3tPANcB"

raspberry_buyer_public_key  = "+fQWj43aTZfihh967fcy9hmsBHgC50QEmJ2i3gJUyxInsyy86llnLQ=="
raspberry_buyer_private_key = "+fQWj43aTZd6Nled4vr0EIB5prNBA4yDQBwu/pSiakVv/clsZg/3vOKGH3rt9zL2GawEeALnRASYnaLeAlTLEiezLLzqWWct"

ip2location_bucket_name     = "newyork_network_next_dev"

relay_backend_public_key    = "kW884wxI4z2919AwG5Fa/4VtzCKDwx1omD/JxdhnfRM="

server_backend_public_key   = "PBZBu3EavRpTr1+6EJMTUCkwLj8DYgN11ky/gwimbyE="

test_server_tag             = "006" # increment this each time you want to deploy the test server
