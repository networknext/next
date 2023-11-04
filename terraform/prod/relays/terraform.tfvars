
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "test_network_next_relay_artifacts"
relay_backend_public_key    = "r8mruPQXKeUtWTRaQNzv3RYgNidRaV8U3p3u8azYeRQ="
relay_backend_url           = "relay-dev.spacecats.net"

raspberry_buyer_public_key  = "+2qlLDtiOJv+U+FYrs+7yLNq0oVkI/Qm7G2TI8lca/iCH7bTdEGzmw=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "H+P06Co6wglahRssa1oUzYbF9FN+QLzvF3T3vIyn5uRh3NSxptgJww=="

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
