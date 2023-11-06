
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://mindful_network_next_backend_artifacts"
google_database_bucket      = "gs://mindful_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "S+f9/oRn2oRI1ED+ze0CQciHHeNJhSWoL6NEK598S2aFrFC5yyrYAg=="
raspberry_buyer_private_key = "S+f9/oRn2oRpFWa7rTn8WCbFcMl02T2nQXiws2GCQFMkHBW1INhvIUjUQP7N7QJByIcd40mFJagvo0Qrn3xLZoWsULnLKtgC"

ip2location_bucket_name     = "mindful_network_next_prod"

relay_backend_public_key    = "e7l47nb1cCEvu6DD+lyJlsAmV5Sntv0fzDK50UaQmgA="

server_backend_public_key   = "SKb19G4Jmnt6dB/vAoAQstDd0IOC+N+RLeYPn1cstpw="
