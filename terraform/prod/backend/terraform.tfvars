
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://solaris_network_next_backend_artifacts"
google_database_bucket      = "gs://solaris_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "wNC7TgbyyGAkKvnImeFBbB4vkNSXw0STe4qAAYic/duERThtam7RQA=="
raspberry_buyer_private_key = "wNC7TgbyyGB7gAEMry/JF9R16Z1Cz9W+N3vurxr9zYQB67p1AfCayiQq+ciZ4UFsHi+Q1JfDRJN7ioABiJz924RFOG1qbtFA"

ip2location_bucket_name     = "solaris_network_next_prod"

relay_backend_public_key    = "Nze8p0TuFD0OSCzgf4lsuhqdcmtIBv5crqfFGKSJJx0="

server_backend_public_key   = "v6S6WtuUBGapWDkWmeACP/MDZoyc27yzZoLvLQdY+m4="
