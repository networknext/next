
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/secrets/next_ssh.pub"
ssh_private_key_file        = "~/secrets/next_ssh"
relay_version               = "relay-debug-1.0.0"
relay_artifacts_bucket      = "newyork_network_next_relay_artifacts"
relay_backend_public_key    = "osFv1SPtMkhezNPuLbNbjp/F8ks5I1Y1QVqD0yLd+0o="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "Oe96rqXLyoxaBc3ce8jo6UeoSNlWpq11Druv8kEO9bP0C2rQW5phkg=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key  = "kLWeaPkL+EYZJDBguhajXE1V5yj5q2WY3I0ITQNN6TELp2J39hOQrA=="

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
}
