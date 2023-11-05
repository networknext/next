
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "auspicious_network_next_relay_artifacts"
relay_backend_public_key    = "chicU7Aw/LB/QuJze4IEzrcgqmgiJf4i0pqrBB3j9Ws="
relay_backend_url           = "relay-dev.spacecats.net"

raspberry_buyer_public_key  = "W0dcu/jQ0xDV6wrlmsc8A+I5+TM7+pwDhYP5wu0HUcY/JzKfMsz1XQ=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "N/d4A0szNUg42dQ8VLUhy6GvBW5e/84T1ncLB86TAww37+0ZBZ6b/g=="

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
