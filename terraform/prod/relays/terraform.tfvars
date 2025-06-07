
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-release-1.0.0"
relay_artifacts_bucket      = "next_network_next_relay_artifacts"
relay_backend_public_key    = "unH/Yxm0C6JCZ1dTGZH2BTBOFhGMcYsOEDURd9qY72w="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "YJHQ5FGeoveQMdSzLRmbOKFRVY6QeMyUX1c4kRA72anucqnPRBr8IA=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "OPsJ/biQrnQEgoJr2oo9zeJG9vVkOUpWklw2+O2nfyy1BljyFxrU8Q=="

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
