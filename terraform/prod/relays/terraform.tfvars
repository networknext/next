
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "auspicious_network_next_relay_artifacts"
relay_backend_public_key    = "JDgqPf2/Bw3MtiGo8MwfDLIHs4PMj52lB1a7Ib4a8HA="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "3YtuzFtzwIoxVH92ocwhGR+JgfVuw+VLVSFZv2Oqp0aMreX1Bb0R4w=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "0xfkba1joh0OB0lnGgMWT84n/9ob5KbbBZTpPaKSKZYv96IV1JyzBg=="

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
