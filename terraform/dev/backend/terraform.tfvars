
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-dev.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"]
google_artifacts_bucket     = "gs://newyork_network_next_backend_artifacts"
google_database_bucket      = "gs://newyork_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

test_buyer_public_key  = "kLWeaPkL+EYZJDBguhajXE1V5yj5q2WY3I0ITQNN6TELp2J39hOQrA=="
test_buyer_private_key = "kLWeaPkL+Ea3Nrmwy30aYExru7gmp5/m5HWDY+M50do9cKO6stcsHRkkMGC6FqNcTVXnKPmrZZjcjQhNA03pMQunYnf2E5Cs"

raspberry_buyer_public_key  = "Oe96rqXLyoxaBc3ce8jo6UeoSNlWpq11Druv8kEO9bP0C2rQW5phkg=="
raspberry_buyer_private_key = "Oe96rqXLyozzVR7TFcwEj/9hcuIcq+RQxHC+Jjhs2p9vJZVp7MDCGloFzdx7yOjpR6hI2VamrXUOu6/yQQ71s/QLatBbmmGS"

ip2location_bucket_name     = "newyork_network_next_dev"

relay_backend_public_key    = "osFv1SPtMkhezNPuLbNbjp/F8ks5I1Y1QVqD0yLd+0o="

server_backend_public_key   = "MuPCD98jRSLzAU4VvgpSpsZmagg30P1M5oIUppGhAAA="

test_server_tag             = "006" # increment this each time you want to deploy the test server
