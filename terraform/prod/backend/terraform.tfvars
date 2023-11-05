
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://auspicious_network_next_backend_artifacts"
google_database_bucket      = "gs://auspicious_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "g/WKV8lUxG/bFhGkh7lnoZBLqfgUWn2Lx0/Zoqg1KIxWE16r/qEPAw=="
raspberry_buyer_private_key = "g/WKV8lUxG/q3RNYgYtcpLSOD5ugfEYDWlpJxnGeh7gzNS4yKbx/TNsWEaSHuWehkEup+BRafYvHT9miqDUojFYTXqv+oQ8D"

ip2location_bucket_name     = "auspicious_network_next_prod"

relay_backend_public_key    = "0qUll8AG7K846S2I05bBk3l0Ak67w9fXC0OIcxlotRs="

server_backend_public_key   = "AHqxqWTA7+30LhbV0KHK5IEXd7DCHtOiKeqaZrJrONE="
