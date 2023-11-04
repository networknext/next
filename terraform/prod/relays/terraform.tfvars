
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "test_network_next_relay_artifacts"
relay_backend_public_key    = "jhf5k2xPTwi2p41oYR/7VBSiTcNZcxF41m1PVNq9QRg="
relay_backend_url           = "relay-dev.spacecats.net"

raspberry_buyer_public_key  = "W1XqU0C/QJCBI7oFaortPtqCLhN3PcwxOd9laPLoy1faxW7e46kRTA=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "qfWdjLu3HjFti1XkneM4z+0+ZUDa03eb7iq/wA0G+Uu6heI5RhRDuQ=="

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
