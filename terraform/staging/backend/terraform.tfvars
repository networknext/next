
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://consoles_network_next_backend_artifacts"
google_database_bucket      = "gs://consoles_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "nyyozFWp3tjYhHqfHyGyn8UbGGUmw7r54sgsgyWiRjY="

server_backend_public_key   = "lA6G77f+t/W3HF2Flk/6IWUlvZ3knNrKjgko9xoqfZY="

load_test_buyer_public_key  = "zkaPRGcAuThG6poXMJ8di/yKzgZEbqyQ6Ky951reRq4sgCm83lV24g=="
load_test_buyer_private_key = "zkaPRGcAuTiYwqkwWEmWSrsxpcJzErC1mkBz3W0PlWdSynr/uuS4jUbqmhcwnx2L/IrOBkRurJDorL3nWt5GriyAKbzeVXbi"

ip2location_bucket_name     = "consoles_network_next_staging"

relay_public_key  = "9597P1ZapnmR5X9sTeOLRIE6ZCqGfOEiyJVq2Rb+bV0="
relay_private_key = "ykNSEqmbzjyz678XfDUnnItB63S1FyBQ7CafO7W1Fgo="
