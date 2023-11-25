
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-dev.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"]
google_artifacts_bucket     = "gs://wilton_network_next_backend_artifacts"
google_database_bucket      = "gs://wilton_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

test_buyer_public_key  = "fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA=="
test_buyer_private_key = "fJ9R1DqVKetUVM2jP6hcMjDTkKlOwFbrtgeCUrpAh3epnaO07HBkm+t6D6S+oSQWpsABrnhzoFdyAGf5HYYozUaQWjxC2RRk"

raspberry_buyer_public_key  = "ZkQ8nC3BTkFjR7ws8GhRz/zVUueQXcde1T0EL+j0/XPVh39S6IoVaw=="
raspberry_buyer_private_key = "ZkQ8nC3BTkGK7DxXGLIPwIx00Jm3ep25n1OmwTZWqYJrKfLj2XmXVmNHvCzwaFHP/NVS55Bdx17VPQQv6PT9c9WHf1LoihVr"

ip2location_bucket_name     = "wilton_network_next_dev"

relay_backend_public_key    = "8jOvRZXo57q2kivlnm4nW9Ff6oi9fgHBnWoUJhz4PQQ="

server_backend_public_key   = "hc3/baZ4FYYaknk1heRK345FK3ZOSpfK1PiMmuIGb1w="

test_server_tag             = "002" # increment this when you need to redeploy the test server
