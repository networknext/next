
extra = ""

vpn_address = "45.79.157.168"

google_service_account      = "terraform@development-394617.iam.gserviceaccount.com"
google_credentials          = "~/secrets/terraform-development.json"
google_project              = "development-394617"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_artifacts_bucket     = "gs://test_network_next_backend_artifacts"
google_database_bucket      = "gs://test_network_next_database_files"
google_machine_type         = "f1-micro"

cloudflare_api_token              = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id_api            = "eba5d882ea2aa23f92dfb50fbf7e3cf4" # -> virtualgo.net
cloudflare_zone_id_relay_backend  = "4690b11494ddc59f6fdf514780d24e0a" # -> losangelesfreewaysatnight.com
cloudflare_zone_id_server_backend = "d74db52c8b51a2b80a76f736713e3edd" # -> spacecats.net

relay_backend_public_key    = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
relay_backend_private_key   = "ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M="
server_backend_public_key   = "TGHKjEeHPtSgtZfDyuDPcQgtJTyRDtRvGSKvuiWWo0A="
server_backend_private_key  = "FXwFqzjGlIwUDwiq1N5Um5VUesdr4fP2hVV2cnJ+yARMYcqMR4c+1KC1l8PK4M9xCC0lPJEO1G8ZIq+6JZajQA=="
ping_key 					= "56MoxCiExN8NCq/+Zlt7mtTsiu+XXSqk8lOHUOm3I64="
api_private_key             = "this is the private key that generates API keys. make sure you change this value in production"
customer_public_key         = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
customer_private_key        = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

maxmind_license_key = "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"
