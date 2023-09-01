
# dev relay variables

env                      = "dev"
vpn_address              = "45.79.157.168"
ssh_public_key_file      = "~/.ssh/id_rsa.pub"
ssh_private_key_file     = "~/.ssh/id_rsa"
relay_version            = "relay-debug-1.0.27"
relay_artifacts_bucket   = "network_next_relay_artifacts"
relay_public_key         = "S0Nr+Hh05vpCpggvUjTwqqC5FX+nuKN02q1K9aOiGQY="
relay_private_key        = "ei8wBWrWnnJOoI3dyQgajOcwfk1axAoKg0L5Xp9UCZw="
relay_backend_public_key = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
relay_backend_hostname   = "dev.losangelesfreewaysatnight.com"

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
