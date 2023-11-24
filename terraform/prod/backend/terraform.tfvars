
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://dogfood_network_next_backend_artifacts"
google_database_bucket      = "gs://dogfood_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "RnNjUqoCkBjnssc4M49RBfA6Fs2FjbsEm6iOUNoma/0FX6Scl8Xumw=="
raspberry_buyer_private_key = "RnNjUqoCkBh+y+mbM6GbyvGvI2nrwCsDQilTfgpII/E/9Q7ypf5Qyueyxzgzj1EF8DoWzYWNuwSbqI5Q2iZr/QVfpJyXxe6b"

ip2location_bucket_name     = "dogfood_network_next_prod"

relay_backend_public_key    = "mifCRKudlMmSZ8IYOb9Iusx3xksbN8PLCfHz0SToI1s="

server_backend_public_key   = "tCtufQgbC5Ao9vcQaVEsftQqqCM+Zmz1FG6XiELuPtg="

test_server_tag             = "001" # increment this when you need to redeploy the test server
