
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-debug-1.0.0"
relay_artifacts_bucket      = "dogfood_network_next_relay_artifacts"
relay_backend_public_key    = "MExWIRGzL7YTfhmEM5xaGX/acijNs/WJoaFKpk9utUs="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "RnNjUqoCkBjnssc4M49RBfA6Fs2FjbsEm6iOUNoma/0FX6Scl8Xumw=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key  = "pcPuGqxFraa3IHicy0U91xwo3+pQbpdrJfQr9OBeVrrNcQvvNn3dJQ=="

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
