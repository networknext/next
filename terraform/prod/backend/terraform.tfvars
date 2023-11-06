
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-prod.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://mindful_network_next_backend_artifacts"
google_database_bucket      = "gs://mindful_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

raspberry_buyer_public_key  = "P5CJSvzIaPhVX3JAMzfVR4qOnM/JoaT/K9RZj+VKqFiqgAKI6HIv6w=="
raspberry_buyer_private_key = "P5CJSvzIaPh1dzdUzOR+Gthnxv4lQolsjF07CQlmp2R+SzBpZcIrKFVfckAzN9VHio6cz8mhpP8r1FmP5UqoWKqAAojoci/r"

ip2location_bucket_name     = "mindful_network_next_prod"

relay_backend_public_key    = "lVABhCBQdbBUSsXyu5KY13SdYdZrfSuIqavtO9Bvl0c="

server_backend_public_key   = "QMdTdN2CnOqrfSdTsZgzQAMp5TSt3idHhLl9BoOnxks="
