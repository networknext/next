
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

test_buyer_public_key  = "zkaPRGcAuThG6poXMJ8di/yKzgZEbqyQ6Ky951reRq4sgCm83lV24g=="
test_buyer_private_key = "zkaPRGcAuTiYwqkwWEmWSrsxpcJzErC1mkBz3W0PlWdSynr/uuS4jUbqmhcwnx2L/IrOBkRurJDorL3nWt5GriyAKbzeVXbi"

raspberry_buyer_public_key  = "ZdjlnZkzaMiY0tloNOtfflxhAAfXpDzGHk9s3qtK991ue5mbOSzdWQ=="
raspberry_buyer_private_key = "ZdjlnZkzaMhWNxyIW/5uLn3wcuUXIvQoU5MA1WlOJaYIIfbWSwZneZjS2Wg0619+XGEAB9ekPMYeT2zeq0r33W57mZs5LN1Z"

ip2location_bucket_name     = "newyork_network_next_dev"

relay_backend_public_key    = "QvHkCNNjQos2A9s1ufDJilvanYgQXNtB5E/eb6M9PDc="

server_backend_public_key   = "qtXDPQZ4St9XihqsNs6hP8QuSCHpr/63aKIOJehTNSg="

test_server_tag             = "006" # increment this each time you want to deploy the test server
