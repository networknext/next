
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "test_network_next_relay_artifacts"
relay_backend_public_key    = "KGhbQrYoDdZJJa1F/su0qTQw007twLPrVwgDs8ZRPyg="
relay_backend_url           = "relay-dev.spacecats.net"

raspberry_buyer_public_key  = "oeYmYtWJz/DhpGxyLZN8w8ukkdJ6yV3kT60sSJhX3LA2SxxfxSMW9A=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "WC6mKDnJmlHPsrvfii0WFJXExS3cVRJK+Hv1wWX1gRuISdNtuklkYw=="

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
