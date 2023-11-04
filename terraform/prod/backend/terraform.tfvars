
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://test_network_next_terraform"
google_artifacts_bucket     = "gs://test_network_next_terraform"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "spacecats.net"

raspberry_buyer_public_key  = "yGzu/W1RB6CWnGtcCwjg+euwcpu2wE0P+9rf57dnjmqqfGieB3vPTQ=="
raspberry_buyer_private_key  = "yGzu/W1RB6CnMfaFZM+EV+fJoGEy1U+YjTqQUMRh5EGHhGOUyCPX/5aca1wLCOD567Bym7bATQ/72t/nt2eOaqp8aJ4He89N"

ip2location_bucket_name     = "test_network_next_local"

relay_backend_public_key    = "TxS2kZqLnYlaD1Alt8i3XEqi/KxoPiM44rtqdLDNVTE="

server_backend_public_key   = "81UP1bcHXA3qJGt89neareDvBUTsUhYdWP4HTm7++q0="
