
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-debug-1.0.0"
relay_artifacts_bucket      = "newyork_network_next_relay_artifacts"
relay_backend_public_key    = "ujPMqkxj4AtXChJ0dfvCvuRiVQyKZjhy+PelZ4xT13k="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "52M6enRvBROk/IX86wOdyZRCBkJMRiqlj6nEA/klgCAdsPW1Eqwv7Q=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key  = "n7u8H8u17JkwfCTkyJ9W1nHzkxjYotCoErwIdpfyAU9uqMgV4SJEjQ=="

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
	"Datapacket" = "datapacket"
	"i3D" = "i3d"
	"Oneqode" = "oneqode"
	"GCore" = "gcore"
	"Hivelocity" = "hivelocity"
	"ColoCrossing" = "colocrossing"
	"phoenixNAP" = "phoenixnap"
	"INAP" = "inap"
	"servers.com" = "serversdotcom"
	"Velia" = "velia"
	"Zenlayer" = "zenlayer"
	"Stackpath" = "stackpath"
	"Latitude" = "latitude"
	"Equinix" = "equinix"
}
