
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "auspicious_network_next_relay_artifacts"
relay_backend_public_key    = "PJ7o9YC/JDWY707D7mSWZneSXds5glKJFMROJtlb/BY="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "2QUAl0H5x2+BSNq/utL48vNQ5irBoKmLB7t9cANZVJmY2VVxf4KnUg=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "GHV58rv7TXENRa7yp+0+M3fvoo1rpITj8ZW6tF1NBqFvhulAMGLorQ=="

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
