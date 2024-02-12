
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://consoles_network_next_backend_artifacts"
google_database_bucket      = "gs://consoles_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "ZdjlnZkzaMiY0tloNOtfflxhAAfXpDzGHk9s3qtK991ue5mbOSzdWQ=="
raspberry_buyer_private_key = "ZdjlnZkzaMhWNxyIW/5uLn3wcuUXIvQoU5MA1WlOJaYIIfbWSwZneZjS2Wg0619+XGEAB9ekPMYeT2zeq0r33W57mZs5LN1Z"

ip2location_bucket_name     = "consoles_network_next_prod"

relay_backend_public_key    = "n7yB7ag9URvrKAUFLJYxaKi/HWN+O16MEQoE/bbf9xM="

server_backend_public_key   = "Nb/JECiiftr9zuSlttiybGnjHzHlTBWE7JFwStPdIZ4="

test_server_tag             = "001" # increment this when you need to redeploy the test server
