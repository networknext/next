
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-release-1.0.0"
relay_artifacts_bucket      = "theodore_network_next_relay_artifacts"
relay_backend_public_key    = "Xg+WVwYouISwX/2h3Slu8knq5W/d+6AID0aF/Vatfg0="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "fUJEyRqdR0X/63g6jCCnXgetsWuzG74zqg5xB4KeJMFK7EyBaXnb1A=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key  = "yaL9uP7tOnc4mG0DMCzRkOs5lShqN0zzrIn6s9jgao1iIv1//3g/Yw=="

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
	"servers.com" = "serversdotcom"
	"Velia" = "velia"
	"Zenlayer" = "zenlayer"
	"Latitude" = "latitude"
	"Equinix" = "equinix"
}
