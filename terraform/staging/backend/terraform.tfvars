
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://solaris_network_next_backend_artifacts"
google_database_bucket      = "gs://solaris_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "TGmsqemngRky2w5HvOWnXO2fa+mLtrEHnwDguiBEpwQ="

server_backend_public_key   = "62gHZrwcZHgaH1p3fov9oacBdK5Expd2tLxqTzdyFr8="

load_test_buyer_public_key  = "n3t//PqEF3yRw1ifWMlt9KMVNe7Q0nH+tgJNeaNDY0IP0/A33ytcEA=="
load_test_buyer_private_key = "n3t//PqEF3x+An7p/9F9SgxIsB6g5ul5foCRp4ktJvz0Z4u6aFQ8TpHDWJ9YyW30oxU17tDScf62Ak15o0NjQg/T8DffK1wQ"

ip2location_bucket_name     = "solaris_network_next_staging"

relay_public_key  = "S1zu8lZCkT2TbS+133zdh/1/9iTjmYsCDTRKMpDQq1U="
relay_private_key = "rVgz/JvjN0PM4GVdyely+iEimuV/VnMvp25xVpD7Ruk="
