
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.0"
relay_artifacts_bucket      = "alocasia_network_next_relay_artifacts"
relay_backend_public_key    = "I+7H0XwZ4liC5yEdPQv8LYCOL+Y57+qfKKOG2iqJYCY="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "bZcUhYFYkrLqvbvB1eYpyrcZBH6Qfr/WMZExUnu4wo5AUhXuZQQxmQ=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "hc/OIfQ3E+IkueTZ8kW5Y0e/Y54luqQushYWjzyfYV5JxYc1RNI5lg=="

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
