
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-debug-1.0.28"
relay_artifacts_bucket      = "mindful_network_next_relay_artifacts"
relay_backend_public_key    = "BR2SmTvTAq7ItyoYWqkSpEKc0UlLPIjyePTk76dbxl8="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "iEDSDTO02hfruP1NR0QSuBEpXC3Q790/aGEhzyoGaEoYZmZ/0lrHew=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "2Zlc0dVlkH7YTpxrhf6nsXkDjtwLUF9dFBpUGI4YowsbeYOsg3l9kQ=="

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
