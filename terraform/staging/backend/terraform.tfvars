
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://dogfood_network_next_backend_artifacts"
google_database_bucket      = "gs://dogfood_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "yjPZFTl4/sUDMTFfDRAw3HXdQTGA0ECt30kPWZ/8XQo="

server_backend_public_key   = "WpNR7iCRHat8z9/V1nSSN5Ycmm83wHlkrg5MtF3714I="

load_test_buyer_public_key  = "pcPuGqxFraa3IHicy0U91xwo3+pQbpdrJfQr9OBeVrrNcQvvNn3dJQ=="
load_test_buyer_private_key = "pcPuGqxFraZh5lLf6S07LjSImiNsbHOdkMXN5hRfnS+Z7fFJECVZH7cgeJzLRT3XHCjf6lBul2sl9Cv04F5Wus1xC+82fd0l"

ip2location_bucket_name     = "dogfood_network_next_staging"

relay_public_key  = "VS5NZAEnCG9l9XjCk042/gStOnsmdsxHPKx3u6UeUiI="
relay_private_key = "u6apUtFATnUTVcU4cO6M0g/kbNFk7qj2rCSfXMcpR0Q="
