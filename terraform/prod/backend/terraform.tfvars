
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://memento_network_next_backend_artifacts"
google_database_bucket      = "gs://memento_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "Df/gqFQMDkNp8GH28WjVEl4I6aB5KxsVipZKow0PhCmQIQIMEeUkOg=="
raspberry_buyer_private_key = "Df/gqFQMDkMA8NLq4ZwJhdcufKBIjVRTQXylBQfpg1QF/BfW3Gr982nwYfbxaNUSXgjpoHkrGxWKlkqjDQ+EKZAhAgwR5SQ6"

ip2location_bucket_name     = "memento_network_next_prod"

relay_backend_public_key    = "uKVEGbP78uSop+rHQb+wsKARub6w2Jh8lCVsWQfUeQY="

server_backend_public_key   = "TIBUzpMOE+FC6f3yAtFFmqrpg4WkfcIhBEWXbfwJEvw="
