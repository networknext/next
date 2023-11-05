
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "test_network_next_relay_artifacts"
relay_backend_public_key    = "1ohrSutpU2m9loSsgw0T+A07pN2c356UKDG3B+V+tjg="
relay_backend_url           = "relay-dev.spacecats.net"

raspberry_buyer_public_key  = "blxyyvzQSRYo6Nfu5l7Ah/RbUB2FAx+AiRzcXFa4fwkj8tYLS3f5+Q=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "WU+Fzt1uQN5Fj6iQ9MU2hm/i/deZRscQtD7DymIvdEpj8RIhxhfcXg=="

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
