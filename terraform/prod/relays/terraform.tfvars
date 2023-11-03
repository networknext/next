
# prod relay variables

env                         = "prod"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-release-1.0.28"
relay_artifacts_bucket      = "test_network_next_relay_artifacts"
relay_public_key            = "S0Nr+Hh05vpCpggvUjTwqqC5FX+nuKN02q1K9aOiGQY="
relay_private_key           = "ei8wBWrWnnJOoI3dyQgajOcwfk1axAoKg0L5Xp9UCZw="
relay_backend_public_key    = "VSIRTzslerQzR1CA1i9mo4HE3IcImWkRG60HnO3u0Sg="
relay_backend_url           = "relay-dev.spacecats.net"

raspberry_datacenters = [
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
