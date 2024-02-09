
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-dev.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"]
google_artifacts_bucket     = "gs://newyork_network_next_backend_artifacts"
google_database_bucket      = "gs://newyork_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

test_buyer_public_key  = "n7u8H8u17JkwfCTkyJ9W1nHzkxjYotCoErwIdpfyAU9uqMgV4SJEjQ=="
test_buyer_private_key = "n7u8H8u17JkUE5QCcC6Qmi1Wgs4ATkawgxwNX98q2YbzTNZojCJnXTB8JOTIn1bWcfOTGNii0KgSvAh2l/IBT26oyBXhIkSN"

raspberry_buyer_public_key  = "52M6enRvBROk/IX86wOdyZRCBkJMRiqlj6nEA/klgCAdsPW1Eqwv7Q=="
raspberry_buyer_private_key = "52M6enRvBRNaR5FIuSrO8b4QiCHQZxAkBUcSpCBighyet/qH3AW3ZKT8hfzrA53JlEIGQkxGKqWPqcQD+SWAIB2w9bUSrC/t"

ip2location_bucket_name     = "newyork_network_next_dev"

relay_backend_public_key    = "ujPMqkxj4AtXChJ0dfvCvuRiVQyKZjhy+PelZ4xT13k="

server_backend_public_key   = "0TZc59EuAmfMuHlcNvYx7s4IH4ERMsZGj56wUaQcgVM="

test_server_tag             = "006" # increment this each time you want to deploy the test server
