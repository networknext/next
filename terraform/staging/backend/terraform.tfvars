
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

relay_backend_public_key    = "DCEcNvVP06YMDOjq3SpNRj5CeZ7FkQUDpgIOTlUtTQQ="

server_backend_public_key   = "gX7FGAe+eOKWMTeaR6WvIQ81RZ0r6EGh5aYyl2xjQLE="

load_test_buyer_public_key  = "8eONd6nhkfFw71ingGK9Q/4YDgKFM5x7odw0iYv8WEhjpdF2nEGYgA=="
load_test_buyer_private_key = "8eONd6nhkfFG8hBpFnZQa+VNRRvGQn541iN/+AWXGPD2PkBF4sXNs3DvWKeAYr1D/hgOAoUznHuh3DSJi/xYSGOl0XacQZiA"

ip2location_bucket_name     = "mindful_network_next_staging"

relay_public_key  = "b6zQWGulFgsdNC7Mm42Mm3vqXsGFuFAJ5AiTosuuzTM="
relay_private_key = "hQJ5cFtANfTw35qPBiWCszbAMZ/df35rquwZab4eb4o="
