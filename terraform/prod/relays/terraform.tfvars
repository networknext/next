
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.0"
relay_artifacts_bucket      = "dogfood_network_next_relay_artifacts"
relay_backend_public_key    = "k4AuJrGEbDFpI3quMgL0jOLEUmeeDmEZUTEQz7lI/G4="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "yAh9zcMJ+lXnHtZpRwq1DdkTK4Oh9yoPWqxy4JIQLU6FTEKxB5YBLQ=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "Lw4WO3aMP6VKgpMFYilu2KRcVwqyoMHodt3QFK+FXpR+c0+DGvITog=="

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
