
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-release"
relay_artifacts_bucket      = "sloclap_network_next_relay_artifacts"
relay_backend_public_key    = "TINP/TnYY/0W7JvLFlYGrB0MUw+b4aIrN20Vq7g5bhU="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "gtdzp3hCfJ9Y+6OOpsWoMChMXhXGDRnY7vkFdHwNqVW0bdp6jjTx6Q=="

raspberry_datacenters = [
	"google.saopaulo.1",
	"google.saopaulo.2",
	"google.saopaulo.3",
	"unity.saopaulo.1",
	"unity.saopaulo.2",
	"unity.saopaulo.3",
]

test_buyer_public_key       = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ=="

test_datacenters = [
	"google.saopaulo.1",
	"google.saopaulo.2",
	"google.saopaulo.3",
	"unity.saopaulo.1",
	"unity.saopaulo.2",
	"unity.saopaulo.3",
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
	"servers.com" = "serversdotcom"
	"Velia" = "velia"
	"Zenlayer" = "zenlayer"
	"Latitude" = "latitude"
	"Equinix" = "equinix"
	"Unity" = "unity"
	"Azure" = "azure"
}
