
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "test_network_next_relay_artifacts"
relay_backend_public_key    = "tpIgkNaOt6trqDssQghqxJQkP1uGwXwyWL5OoReHylM="
relay_backend_url           = "relay-dev.spacecats.net"

raspberry_buyer_public_key  = "GRlQrb34lQ3BqXhKMEbHPrxxxwe4MA7BT7OPklLGYDsb4aFgJL9Dzw=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "+rtOkfU/2mddf741YddaW9IXchGQGv8qQLnSe1nnVBaP68USps7BbQ=="

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
