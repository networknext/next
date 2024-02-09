
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-release-1.0.0"
relay_artifacts_bucket      = "newyork_network_next_relay_artifacts"
relay_backend_public_key    = "MzEsgGvPag3t+/ZA0NvwnUZdNedZGLySsm+baHPbAGY="
relay_backend_url           = "relay.virtualgo.net"

raspberry_buyer_public_key  = "+fQWj43aTZfihh967fcy9hmsBHgC50QEmJ2i3gJUyxInsyy86llnLQ=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key  = "9fuymsQpqwyyimT9iSSXJi4Dnf3XHM6XlYq0kIPOawhLHjN7TwDXAQ=="

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
