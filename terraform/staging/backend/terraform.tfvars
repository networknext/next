
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://helsinki_network_next_backend_artifacts"
google_database_bucket      = "gs://helsinki_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "ad602Vfc4yG43hVChLdpdnVcPP/IoDWWqFmxX3S8yVo="

server_backend_public_key   = "JP+Nd0vwHzApOOtHWuqiqvB8OztqD78ScRZioLlTCik="

load_test_buyer_public_key  = "9qzGNONKAHTBaPsm+b9pPUgEvekv3iKZBdXJt7eSBePFkeWtoxpGig=="
load_test_buyer_private_key = "9qzGNONKAHTmQiE2vRhmIL7cB+GB2v0HrRNozHefaNCTHNow8WNqj8Fo+yb5v2k9SAS96S/eIpkF1cm3t5IF48WR5a2jGkaK"

ip2location_bucket_name     = "helsinki_network_next_staging"

relay_public_key  = "02YtwLT5RTPlEjxEe0oo/0EP3OFLOJdWLA5jxz3J5VY="
relay_private_key = "JB2hC7sEaj2ujpthoOWyEqKAqsBzQgrutUBopPShiuM="
