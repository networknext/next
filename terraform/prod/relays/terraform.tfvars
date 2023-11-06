
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "mindful_network_next_relay_artifacts"
relay_backend_public_key    = "ogy0lpZgLxogkp1M5SmOCx8PtbIFFHyGHCefOe11+WU="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "T73BsScGIU7bUVHCD50I064wMRUBZGRqG8rH0moYKbawCRuD2nPUWA=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "9bR0O5cFm8VPNU+tb5neR83LW8g/uLMBmHn/tNuUFUvCE5FB8daA5w=="

test_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

sellers = {
	"Akamai" = "akamai"
	"Amazon" = "amazon"
	"Google" = "google"
	"VULTR"  = "vultr"
}
