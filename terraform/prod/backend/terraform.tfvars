
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://auspicious_network_next_terraform"
google_database_bucket      = "gs://auspicious_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "2QUAl0H5x2+BSNq/utL48vNQ5irBoKmLB7t9cANZVJmY2VVxf4KnUg=="
raspberry_buyer_private_key = "2QUAl0H5x2+0+TGHyjGRVwZfNZlTCDkgpSKergLaFhNZE7zK39CpmIFI2r+60vjy81DmKsGgqYsHu31wA1lUmZjZVXF/gqdS"

ip2location_bucket_name     = "auspicious_network_next_local"

relay_backend_public_key    = "PJ7o9YC/JDWY707D7mSWZneSXds5glKJFMROJtlb/BY="

server_backend_public_key   = "6XB35QZH1guyI68cMnqZdBYkdbYiib4D1kJshlgUe2w="
