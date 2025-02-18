
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://theodore_network_next_backend_artifacts"
google_database_bucket      = "gs://theodore_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "MXpPOBtw5H6SmK+obrpnpjFlYNxFdnUQScQfmEjcHQA="

server_backend_public_key   = "JGUhV7NTnwiuMYr6EIER82lHiyoAPuJSfvvsmCH+YwI="

load_test_buyer_public_key  = "yaL9uP7tOnc4mG0DMCzRkOs5lShqN0zzrIn6s9jgao1iIv1//3g/Yw=="
load_test_buyer_private_key = "yaL9uP7tOncF85rlqE3P/Kee/C945C3e57zslfjB3h7/agqRKuyGHDiYbQMwLNGQ6zmVKGo3TPOsifqz2OBqjWIi/X//eD9j"

ip2location_bucket_name     = "theodore_network_next_staging"

relay_public_key  = "+ONHHci1bizkWzi4MTt1E5b0p0M5Xe0PhUay5H5KIl4="

relay_private_key = "S0S/gyTx2v1vmgAyuyEx6wsOtG0p6Q6GfP3PEnswYTc="
