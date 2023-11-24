
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-dev.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"]
google_artifacts_bucket     = "gs://dogfood_network_next_backend_artifacts"
google_database_bucket      = "gs://dogfood_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

test_buyer_public_key       = "pcPuGqxFraa3IHicy0U91xwo3+pQbpdrJfQr9OBeVrrNcQvvNn3dJQ=="
test_buyer_private_key      = "pcPuGqxFraZh5lLf6S07LjSImiNsbHOdkMXN5hRfnS+Z7fFJECVZH7cgeJzLRT3XHCjf6lBul2sl9Cv04F5Wus1xC+82fd0l"

raspberry_buyer_public_key  = "RnNjUqoCkBjnssc4M49RBfA6Fs2FjbsEm6iOUNoma/0FX6Scl8Xumw=="
raspberry_buyer_private_key = "RnNjUqoCkBh+y+mbM6GbyvGvI2nrwCsDQilTfgpII/E/9Q7ypf5Qyueyxzgzj1EF8DoWzYWNuwSbqI5Q2iZr/QVfpJyXxe6b"

ip2location_bucket_name     = "dogfood_network_next_dev"

relay_backend_public_key    = "MExWIRGzL7YTfhmEM5xaGX/acijNs/WJoaFKpk9utUs="

server_backend_public_key   = "TZFLBJxO0OB+IhM1Tumdb7ClYpnOrdRdEmFJ5mGolAM="

test_server_tag             = "001" # increment this when you need to redeploy the test server
