
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

raspberry_buyer_public_key  = "iEDSDTO02hfruP1NR0QSuBEpXC3Q790/aGEhzyoGaEoYZmZ/0lrHew=="
raspberry_buyer_private_key = "iEDSDTO02hcDD6I1qjw/dMXEYJ+YbFzEANOh43799NVFuHN2AsLJUOu4/U1HRBK4ESlcLdDv3T9oYSHPKgZoShhmZn/SWsd7"

ip2location_bucket_name     = "mindful_network_next_prod"

relay_backend_public_key    = "H5CQPnVmPc7BDi2RKEhaS1RKEmLrGUDPEGLpSfdtdE8="

server_backend_public_key   = "2SDxhk3e1hH+Tqke6Q7EO7+FbBIFwquwnHL9F5SBFGE="
