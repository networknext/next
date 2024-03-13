
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-dev.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"]
google_artifacts_bucket     = "gs://consoles_network_next_backend_artifacts"
google_database_bucket      = "gs://consoles_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

test_buyer_public_key  = "5Vr+VZdUXckgQwHdPRftc/8IUWDL7ZftvBOzE/+Zpp+PIjSU0Kxmwg=="
test_buyer_private_key = "5Vr+VZdUXckPZsd89NGTmXASmmlHRuWiyVs7orAxRV6hDkvTc3VMtCBDAd09F+1z/whRYMvtl+28E7MT/5mmn48iNJTQrGbC"

raspberry_buyer_public_key  = "wXf5tMK7F+x/4AytQh/8xYj+mBGIC5lDInYYS4M6RgrzGWaYqWUTFg=="
raspberry_buyer_private_key = "wXf5tMK7F+wuYlVfn8LtaIk3qNpfgdWHaCfl2XA2dVSld99R8EDRWn/gDK1CH/zFiP6YEYgLmUMidhhLgzpGCvMZZpipZRMW"

ip2location_bucket_name     = "consoles_network_next_dev"

relay_backend_public_key    = "OBzjZnC9Dpr8W85ITOjsTfjoX9NJTux1vHzf2blXSkk="

server_backend_public_key   = "emXNUJn34nPyGKWEAs7z18VXL+M8uo0QAZFHEGipeds="

test_server_tag             = "007" # increment this each time you want to deploy the test server
