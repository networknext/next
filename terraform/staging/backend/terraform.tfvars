
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

relay_backend_public_key    = "+qOB7SA7Bzg7tiN72lwEVJ2gVhaI6zhFqEkvG9SrJAE="

server_backend_public_key   = "XkE+xWyskCf8jq3tQ1L3im1224RgDhoKBbEp1+cc6lI="

load_test_buyer_public_key  = "XS1IFDqcGudRvHbJltGmfnsO78MRB4tiLHqzsHgFCbxqwgKzNesDwg=="
load_test_buyer_private_key = "XS1IFDqcGuf1cCsPYwRyeMJxjOcWSr7Tr91M+slYCDr4XVnXM8OCCFG8dsmW0aZ+ew7vwxEHi2IserOweAUJvGrCArM16wPC"

ip2location_bucket_name     = "auspicious_network_next_staging"
