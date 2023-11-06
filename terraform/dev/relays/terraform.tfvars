
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-debug-1.0.28"
relay_artifacts_bucket      = "mindful_network_next_relay_artifacts"
relay_backend_public_key    = "EZSfiQVowpl8Uqi5Jpg81+PZWqwVnoN7OzBPLIPbNgU="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "emtKwwJhDkpGotL1Wxg4M1d4EU7DtjOwLYpd3uWjng+hZF4TLI3TkA=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "YdayThVTwjK5klktRiH1CX5dkvENGG29ObRaGbjsdwxMZh0awtB8uw=="

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
