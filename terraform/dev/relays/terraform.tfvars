
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-debug-1.0.0"
relay_artifacts_bucket      = "alocasia_network_next_relay_artifacts"
relay_backend_public_key    = "D5W6IV4blvCzL+UCmlGmr0er4EC3nj7MTDPsgaIQg2U="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "d57/wLdTcI/S0XqCqsxNt2iLeFaszI5DkCy9/0fWIsOoaVPt141SWA=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "0/bHg4VPjOrB+Jp8kiyAkyPhrSnfOZi9jTNLFNbsTbS2e3MSeQdv7Q=="

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
