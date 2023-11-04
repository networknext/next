
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://test_network_next_terraform"
google_artifacts_bucket     = "gs://test_network_next_terraform"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "spacecats.net"

relay_backend_public_key    = "kW1PyT1o9zd1oEN10oK5+RvtLcyLL9QqmY+MY1kpejU="

server_backend_public_key   = "r9gTi2Jd7aoGajNEsHUKaS32IuZ1Fr5Zj76trkslm34="

load_test_buyer_public_key  = "+rtOkfU/2mddf741YddaW9IXchGQGv8qQLnSe1nnVBaP68USps7BbQ=="
load_test_buyer_private_key = "+rtOkfU/2mdoHh583zWbJgGEHMq4fdrWzR0P9FiO+WC79McuPJaGBl1/vjVh11pb0hdyEZAa/ypAudJ7WedUFo/rxRKmzsFt"

ip2location_bucket_name     = "test_network_next_local"
