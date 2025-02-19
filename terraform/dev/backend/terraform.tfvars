
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-dev.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"]
google_artifacts_bucket     = "gs://theodore_network_next_backend_artifacts"
google_database_bucket      = "gs://theodore_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

test_buyer_public_key       = "yaL9uP7tOnc4mG0DMCzRkOs5lShqN0zzrIn6s9jgao1iIv1//3g/Yw=="
test_buyer_private_key      = "yaL9uP7tOncF85rlqE3P/Kee/C945C3e57zslfjB3h7/agqRKuyGHDiYbQMwLNGQ6zmVKGo3TPOsifqz2OBqjWIi/X//eD9j"

raspberry_buyer_public_key  = "fUJEyRqdR0X/63g6jCCnXgetsWuzG74zqg5xB4KeJMFK7EyBaXnb1A=="
raspberry_buyer_private_key = "fUJEyRqdR0UE0yFmg9tHAE6flSw2TnVD7+jA3q6etkbMudO84dyujf/reDqMIKdeB62xa7MbvjOqDnEHgp4kwUrsTIFpedvU"

ip2location_bucket_name     = "theodore_network_next_dev"

relay_backend_public_key    = "2UrvlOyXfQk+F3QZhXrP36kecqlLaSo28+eIubVDS2Y="

server_backend_public_key   = "FOl98B81AtRxlmisjkL5ROaVaV3bC6v3/hv+wubm6hs="

test_server_region          = "us-central1"
test_server_zone            = "us-central1-a"
test_server_tag             = "007" # increment this each time you want to deploy the test server

disable_backend             = false
disable_raspberry           = false
disable_ip2location         = false
