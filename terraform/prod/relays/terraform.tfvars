
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-release-1.0.0"
relay_artifacts_bucket      = "wilton_network_next_relay_artifacts"
relay_backend_public_key    = "jRUKe8S7x5s6Se537gkjpIHq4JWKyEd3MoQ7iQOqoTQ="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "ZkQ8nC3BTkFjR7ws8GhRz/zVUueQXcde1T0EL+j0/XPVh39S6IoVaw=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key  = "fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA=="

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
