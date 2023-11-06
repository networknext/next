
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-debug-1.0.28"
relay_artifacts_bucket      = "mindful_network_next_relay_artifacts"
relay_backend_public_key    = "qrbwHhOxsJA8CKSuhICXCGbgb1WABxR1quREOF5xth4="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "S+f9/oRn2oRI1ED+ze0CQciHHeNJhSWoL6NEK598S2aFrFC5yyrYAg=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "8eONd6nhkfFw71ingGK9Q/4YDgKFM5x7odw0iYv8WEhjpdF2nEGYgA=="

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
