
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://auspicious_network_next_backend_artifacts"
google_database_bucket      = "gs://auspicious_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "uCy7rbmG396VzPqmZgj+MZoGPQgYPJ7+w72QBXR2fyc="

server_backend_public_key   = "RlsWe6kJIovRSaLbcOo5/fEsXSwrXtFbVIUM7DJAaAA="

load_test_buyer_public_key  = "GHV58rv7TXENRa7yp+0+M3fvoo1rpITj8ZW6tF1NBqFvhulAMGLorQ=="
load_test_buyer_private_key = "GHV58rv7TXENBEGJXojuMx3KrMX7b+uktlMa09WhFiRjKfet5vCvjQ1FrvKn7T4zd++ijWukhOPxlbq0XU0GoW+G6UAwYuit"

ip2location_bucket_name     = "auspicious_network_next_staging"
