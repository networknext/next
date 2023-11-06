
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

relay_backend_public_key    = "VqbY7q4o4CJZIulr+NwhCg9LaoC+I8LhTQbXID2fnVU="

server_backend_public_key   = "jERuHBvsgci3w8MAkejuzzubTZ/kSW53/lMs/tPPdwQ="

load_test_buyer_public_key  = "G3TWXYHw0JiyXFxLSyi3wWezB8ISLW0l0GlJMP9JDgutWXnW5X/eOA=="
load_test_buyer_private_key = "G3TWXYHw0JhnLHUp5pNfjUIr3HJpz8Gz/jcMD1XoDTGR59Sdf28p5rJcXEtLKLfBZ7MHwhItbSXQaUkw/0kOC61Zedblf944"

ip2location_bucket_name     = "mindful_network_next_staging"

relay_public_key  = "ErmQ+VAwr+7HCcCYpz7r7yftZtU40Bw2AVXs0190Eks="
relay_private_key = "6fYy8jm3pZlQN44VOVTkQyQ1BFukXVFl91r9/bj+LDY="
