
tag = "dev-007"

extra = ""

vpn_address = "45.79.157.168"

google_service_account      = "terraform@development-394617.iam.gserviceaccount.com"
google_credentials          = "~/secrets/terraform-development.json"
google_project              = "development-394617"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_artifacts_bucket     = "gs://test_network_next_dev_artifacts"
google_machine_type         = "f1-micro"

cloudflare_api_token              = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id_api            = "eba5d882ea2aa23f92dfb50fbf7e3cf4" # -> virtualgo.net
cloudflare_zone_id_relay_backend  = "4690b11494ddc59f6fdf514780d24e0a" # -> losangelesfreewaysatnight.com
cloudflare_zone_id_server_backend = "d74db52c8b51a2b80a76f736713e3edd" # -> spacecats.net
