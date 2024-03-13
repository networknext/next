
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://consoles_network_next_backend_artifacts"
google_database_bucket      = "gs://consoles_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "QXFdH+crj7e/3fryUJ85PmBa7wkHvBGpGL1+WMLd/zM="

server_backend_public_key   = "lsJbdkx6OrfNRyN+orNxWHhVoNSHT/m0ii4WN8EvW4M="

load_test_buyer_public_key  = "5Vr+VZdUXckgQwHdPRftc/8IUWDL7ZftvBOzE/+Zpp+PIjSU0Kxmwg=="
load_test_buyer_private_key = "5Vr+VZdUXckPZsd89NGTmXASmmlHRuWiyVs7orAxRV6hDkvTc3VMtCBDAd09F+1z/whRYMvtl+28E7MT/5mmn48iNJTQrGbC"

ip2location_bucket_name     = "consoles_network_next_staging"

relay_public_key  = "DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk="
relay_private_key = "QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE="
