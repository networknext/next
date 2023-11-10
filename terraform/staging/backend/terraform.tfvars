
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://memento_network_next_backend_artifacts"
google_database_bucket      = "gs://memento_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "asuVKW6Z5C6CrQ9xxfhkZUlgiR+/m5E2Hvs5S8oWYnI="

server_backend_public_key   = "vYmvYMNJgjWzE4POciiZvlrdgIaO8jpgmq7b8eSNpqM="

load_test_buyer_public_key  = "AN+VWuqgAQfd5d+QeT6apNf+Gbb9rqFBMtk5M+GoMvwS1Eqz764X/A=="
load_test_buyer_private_key = "AN+VWuqgAQdhmzB4XOT89baswrIaX6WS7dTIW8U6deMZdoemQh9qoN3l35B5Ppqk1/4Ztv2uoUEy2Tkz4agy/BLUSrPvrhf8"

ip2location_bucket_name     = "memento_network_next_staging"

relay_public_key  = "swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA="
relay_private_key = "2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8="
