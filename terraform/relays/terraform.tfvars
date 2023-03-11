credentials              = "~/Documents/terraform-relays.json"
project                  = "relays-380114"
vpn_address              = "45.33.53.242"
ssh_public_key_file      = "~/.ssh/id_rsa.pub"
ssh_private_key_file     = "~/.ssh/id_rsa"
env                      = "dev"
relay_version            = "1.0.19"
relay_artifacts_bucket   = "network_next_relay_artifacts"
relay_public_key         = "S0Nr+Hh05vpCpggvUjTwqqC5FX+nuKN02q1K9aOiGQY="
relay_private_key        = "ei8wBWrWnnJOoI3dyQgajOcwfk1axAoKg0L5Xp9UCZw="
relay_backend_public_key = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
relay_backend_hostname   = "dev.losangelesfreewaysatnight.com"

amazon_relays = [
	{
		name = "amazon.virginia.1"
		zone = "us-east-1a"
		type = "a1.large"
		ami  = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-arm64-server-*"
	},
	{
		name = "amazon.virginia.2"
		zone = "us-east-1b"
		type = "a1.large"
		ami  = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-arm64-server-*"
	},
	{
		name = "amazon.virginia.3"
		zone = "us-east-1c"
		type = "m5a.large"
		ami  = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
	},
	{
		name = "amazon.virginia.4"
		zone = "us-east-1d"
		type = "a1.large"
		ami  = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-arm64-server-*"
	},
	{
		name = "amazon.virginia.5"
		zone = "us-east-1e"
		type = "m4.large"
		ami  = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
	},
	{
		name = "amazon.virginia.6"
		zone = "us-east-1f"
		type = "m5a.large"
		ami  = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
	},

]
