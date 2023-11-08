
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-debug-1.0.28"
relay_artifacts_bucket      = "memento_network_next_relay_artifacts"
relay_backend_public_key    = "TNBiEpx62+2zJ18IdOt3yjpnnJWjUQVV/7dmXWMMVRY="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "BsLPRMwxGJfGvuxGSQQ4SKOy88JqbSTFGxPkJJkIQWaa5h5+W6xXqg=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "468kCDmMIj1QK5jnDzhQfoRr/vyix8zQGBV7nn82Cjww1ch3triSdA=="

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
