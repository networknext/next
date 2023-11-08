
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://memento_network_next_backend_artifacts"
google_database_bucket      = "gs://memento_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "zQzW72VIU9M2TH5qFCtIrFTB3A2OOql2S42bRm7ivzg="

server_backend_public_key   = "pbKKAPHlpn5+s48/fWzHtxIs0weT/SN5a7IASjtbSXM="

load_test_buyer_public_key  = "468kCDmMIj1QK5jnDzhQfoRr/vyix8zQGBV7nn82Cjww1ch3triSdA=="
load_test_buyer_private_key = "468kCDmMIj0t6IBjVtcEuxBTlkcvhr15fecY7HFI0mS7Lx6SBAtAW1ArmOcPOFB+hGv+/KLHzNAYFXuefzYKPDDVyHe2uJJ0"

ip2location_bucket_name     = "memento_network_next_staging"

relay_public_key  = "qQ4V/6qn++aVZVf1ZdSG0KOlyn3ZOZNpfKSNRYLSfgo="
relay_private_key = "ok3BjDFfFgkHctwKOKPLtFFfsUo6ieqtne8PgU6iPBA="
