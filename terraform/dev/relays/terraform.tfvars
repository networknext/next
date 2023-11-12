
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-debug-1.0.28"
relay_artifacts_bucket      = "solaris_network_next_relay_artifacts"
relay_backend_public_key    = "5zf4rJXulcTmkceuqsR44NHgEhr3QdL+2wRFJkyFtVU="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "wNC7TgbyyGAkKvnImeFBbB4vkNSXw0STe4qAAYic/duERThtam7RQA=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "n3t//PqEF3yRw1ifWMlt9KMVNe7Q0nH+tgJNeaNDY0IP0/A33ytcEA=="

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
