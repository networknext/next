
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

raspberry_buyer_public_key  = "GRlQrb34lQ3BqXhKMEbHPrxxxwe4MA7BT7OPklLGYDsb4aFgJL9Dzw=="
raspberry_buyer_private_key = "GRlQrb34lQ0dyJT1kIMpQyqhDphQuZaAy+cfWaiCBK8cj/2B2O7pKsGpeEowRsc+vHHHB7gwDsFPs4+SUsZgOxvhoWAkv0PP"

ip2location_bucket_name     = "test_network_next_local"

relay_backend_public_key    = "tpIgkNaOt6trqDssQghqxJQkP1uGwXwyWL5OoReHylM="

server_backend_public_key   = "aCc9C6tL5as8zUA2hYz+vmuIxW6v2lcy23ppfwxnWAQ="
