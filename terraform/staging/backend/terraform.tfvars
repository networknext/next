
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://auspicious_network_next_backend_artifacts"
google_database_bucket      = "gs://auspicious_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "xvCMS55jcVYVOzjrhdyokFVugbgCiij1nSRA2qQE3Tw="

server_backend_public_key   = "UKiSft2agzQt75fRuwrvt3Cqaw25i6imvhQpRHctQtk="

load_test_buyer_public_key  = "0xfkba1joh0OB0lnGgMWT84n/9ob5KbbBZTpPaKSKZYv96IV1JyzBg=="
load_test_buyer_private_key = "0xfkba1joh1qT1NXRyNteMhtDbfkNJKVwKxO/T/kUf7DFA9g0Zbdsw4HSWcaAxZPzif/2hvkptsFlOk9opIpli/3ohXUnLMG"

ip2location_bucket_name     = "auspicious_network_next_staging"

relay_public_key  = "lFA+GivH+HnvoxFRiVDQZxd7WnKeuh5FUtsQf13/Mz0="
relay_private_key = "Yxhh6xYA3skmlWvZuN44Eeyvj9U2v+a6OsWUqsyqK0U="
