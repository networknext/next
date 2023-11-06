
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-debug-1.0.28"
relay_artifacts_bucket      = "auspicious_network_next_relay_artifacts"
relay_backend_public_key    = "zRblD5fjEXW9je6q3NXeq8EEdVF3vhmLMotFUUu9ahA="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "bEXPecXWQSn5ciiHrX3lHT+J6/MxJjmoimOP/PktT3oW1Cd88Vnm8w=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "GLPP7IXT09HXFk8T1T8vBPC5Pe/YFSgWZtjPFRYFJ6klHS2b3ZNc5Q=="

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
