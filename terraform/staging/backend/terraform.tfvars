
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://mindful_network_next_backend_artifacts"
google_database_bucket      = "gs://mindful_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "sEP3o+p/Xw/spPEEZtKNGJme/TwqJXlXCsdP5ggTozI="

server_backend_public_key   = "NV0yjuPuUQAtvy6qzyT48AkGANGYzibCAAqyFbPJ9ig="

load_test_buyer_public_key  = "SwsfaCgkMj602mH2dlzbz6+FME9LOHMybpoAVlTcfbJorvFoc9bRDQ=="
load_test_buyer_private_key = "SwsfaCgkMj7G2/JQVe3aAX9Nx9kTO770W0ndXp0X8ORsToniC9hzSbTaYfZ2XNvPr4UwT0s4czJumgBWVNx9smiu8Whz1tEN"

ip2location_bucket_name     = "mindful_network_next_staging"

relay_public_key  = "xWCx2sztuBjmPM4Pevwff1XhdaGzhc+mE2PSyuR7oyI="
relay_private_key = "Siz1YyCV5oNmbv58EOO3k+0hCLj9rG0EvPZhVVgPP3k="
