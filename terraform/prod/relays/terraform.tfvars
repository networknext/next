
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-release-1.0.0"
relay_artifacts_bucket      = "newyork_network_next_relay_artifacts"
relay_backend_public_key    = "NQiV7UaFevD4yA7xq6Qiy0nTQJ2WPr9g6AERXCYlln4="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "/AUcbl5fuLxmPYEjtjBbVFnPJDlUuWrcntrVL5na6NsYzP2lsoOR5A=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key  = "9qzGNONKAHTBaPsm+b9pPUgEvekv3iKZBdXJt7eSBePFkeWtoxpGig=="

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
}
