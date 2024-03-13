
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-release-1.0.0"
relay_artifacts_bucket      = "xdp_network_next_relay_artifacts"
relay_backend_public_key    = "JWQL5GKNSqBJSbc6Ybga9oxFheKgLEmNjZwyTI/nJg0="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "wXf5tMK7F+x/4AytQh/8xYj+mBGIC5lDInYYS4M6RgrzGWaYqWUTFg=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key  = "5Vr+VZdUXckgQwHdPRftc/8IUWDL7ZftvBOzE/+Zpp+PIjSU0Kxmwg=="

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
