
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://mindful_network_next_backend_artifacts"
google_database_bucket      = "gs://mindful_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "SM7+AwZxoGhxQuEabjNLxAewRHOg1AwAEI14ppI3u2c="

server_backend_public_key   = "Qsv2loePkUmQeHnWQ+qnWLDVpj3xNYRFmCduT8OxRj4="

load_test_buyer_public_key  = "2Zlc0dVlkH7YTpxrhf6nsXkDjtwLUF9dFBpUGI4YowsbeYOsg3l9kQ=="
load_test_buyer_private_key = "2Zlc0dVlkH7EcOeBYPpLAN2NabD8G/txrKh4vDC0okLqmzZX1JSq+NhOnGuF/qexeQOO3AtQX10UGlQYjhijCxt5g6yDeX2R"

ip2location_bucket_name     = "mindful_network_next_staging"

relay_public_key  = "XW18jSTt64uK+XGqOxAlY4zY01TCRKCKSTrNIT/ZOFM="
relay_private_key = "nG2dJRaobo42SbyO59xxd4QGcRIWNrslnFKkK9kh0GY="
