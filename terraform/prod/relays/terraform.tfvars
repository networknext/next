
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "mindful_network_next_relay_artifacts"
relay_backend_public_key    = "s14UkUu+hW5tDBSMPdk+y3CUaGZsocRZL0PR78RYmhE="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "h5E1IwxPLHjrJKXwnAY2J5MLbmLtm0+3SOirrUMZrhDC6KxN+fQQMw=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "SwsfaCgkMj602mH2dlzbz6+FME9LOHMybpoAVlTcfbJorvFoc9bRDQ=="

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
