
env                      = "dev"
vpn_address              = "45.33.53.242"
ssh_public_key_file      = "~/.ssh/id_rsa.pub"
ssh_private_key_file     = "~/.ssh/id_rsa"
relay_version            = "1.0.19"
relay_artifacts_bucket   = "network_next_relay_artifacts"
relay_public_key         = "S0Nr+Hh05vpCpggvUjTwqqC5FX+nuKN02q1K9aOiGQY="
relay_private_key        = "ei8wBWrWnnJOoI3dyQgajOcwfk1axAoKg0L5Xp9UCZw="
relay_backend_public_key = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
relay_backend_hostname   = "dev.losangelesfreewaysatnight.com"

# ----------------------------------------------------------------------

google_credentials = "~/secrets/terraform-relays.json"
google_project     = "relays-380114"

google_relays = [

/*
	# us-central1: IOWA
	{
		name   = "google.iowa.1"
		zone   = "us-central1-a"
		region = "us-central1"
		type   = "n1-standard-2"
		image  = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
	},
	{
		name   = "google.iowa.2"
		zone   = "us-central1-b"
		region = "us-central1"
		type   = "n1-standard-2"
		image  = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
	},
	{
		name   = "google.iowa.3"
		zone   = "us-central1-c"
		region = "us-central1"
		type   = "n1-standard-2"
		image  = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
	},
	{
		name   = "google.iowa.4"
		zone   = "us-central1-f"
		region = "us-central1"
		type   = "n1-standard-2"
		image  = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
	},

	# us-west2: OREGON
	{
		name   = "google.oregon.1"
		zone   = "us-west1-a"
		region = "us-west1"
		type   = "n1-standard-2"
		image  = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
	},
	{
		name   = "google.oregon.2"
		zone   = "us-west1-b"
		region = "us-west1"
		type   = "n1-standard-2"
		image  = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
	},

	# us-west3: SALT LAKE CITY
	{
		name   = "google.saltlakecity.1"
		zone   = "us-west3-a"
		region = "us-west3"
		type   = "n1-standard-2"
		image  = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
	}
*/
]

# ----------------------------------------------------------------------

amazon_relays = [

/*
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
*/
]

# ----------------------------------------------------------------------

akamai_relays = [

/*
	# us-east: NEW YORK
	{
		name   = "akamai.newyork"
		region = "us-east"
		type   = "g6-dedicated-2"
		image  = "linode/ubuntu22.04"
	},

	# us-east: ATLANTA
	{
		name   = "akamai.atlanta"
		region = "us-southeast"
		type   = "g6-dedicated-2"
		image  = "linode/ubuntu22.04"
	},

	# us-west: FREMONT
	{
		name   = "akamai.fremont"
		region = "us-west"
		type   = "g6-dedicated-2"
		image  = "linode/ubuntu22.04"
	},

	# eu-central: FRANKFURT
	{
		name   = "akamai.frankfurt"
		region = "eu-central"
		type   = "g6-dedicated-2"
		image  = "linode/ubuntu22.04"
	},
*/
]

# ----------------------------------------------------------------------

vultr_relays = [

/*
	# sea: SEATTLE
	{
		name   = "vultr.seattle"
		region = "sea"
		plan   = "vc2-1c-1gb"
		os     = "Ubuntu 22.04 LTS x64"
	},

	# ord: CHICAGO
	{
		name   = "vultr.chicago"
		region = "ord"
		plan   = "vc2-1c-1gb"
		os     = "Ubuntu 22.04 LTS x64"
	},
*/
]

# ----------------------------------------------------------------------

latitude_relays = [

	# VIRGINIA
	{
		name   = "latitude.virginia"
		site   = "ASH"
		plan   = "c1-tiny-x86"
		os     = "ubuntu_22_04_x64_lts"
	},
]

# ----------------------------------------------------------------------

equinix_project_id = "22b594e2-512c-42fb-881f-7a94da7d2737"

equinix_relays = [

/*
	# DALLAS
	{
		name   = "equinix.dallas.1"
		zone   = "da11"
		plan   = "c3.small.x86"
		os     = "ubuntu_22_04"
	},
	{
		name   = "equinix.dallas.2"
		zone   = "da6"
		plan   = "c3.medium.x86"
		os     = "ubuntu_22_04"
	},

	# CHICAGO
	{
		name   = "equinix.chicago.1"
		zone   = "ch3"
		plan   = "c3.small.x86"
		os     = "ubuntu_22_04"
	},
*/
]

# see: https://deploy.equinix.com/developers/docs/metal/locations/metros/

# ----------------------------------------------------------------------

hivelocity_relays = [

/*
	# DALLAS
	{
		name   = "hivelocity.dallas.1"
		zone   = "DAL1"
		os     = "Ubuntu 20.x"
	},

	# TAMPA
	{
		name   = "hivelocity.tampa.1"
		zone   = "TPA1"
		os     = "Ubuntu 20.x"
	},
	{
		name   = "hivelocity.tampa.2"
		zone   = "TPA2"
		os     = "Ubuntu 20.x"
	},
*/
]

// see: https://developers.hivelocity.net/docs/facilities

# ----------------------------------------------------------------------

gcore_project_id = "1"

gcore_relays = [

/*
	# DALLAS
	{
		name   = "hivelocity.dallas.1"
		zone   = "DAL1"
		os     = "Ubuntu 20.x"
	},
*/
]

// see: https://apidocs.gcore.com/cloud

# ----------------------------------------------------------------------

phoenixnap_client_id = "ANUS"

phoenixnap_relays = [

/*
	# DALLAS
	{
		name   = "phoenixnap.phoenix"
		zone   = "PHX"
		type   = "s1.c1.medium"
		os     = "ubuntu/jammy"
	},
*/
]

# ----------------------------------------------------------------------

bare_metal_relays = [
	{
		relay_name       = "datapacket.losangeles.a"
		datacenter_name  = "datapacket.losangeles"
		supplier_name    = "datapacket"
		public_address   = "127.0.0.1:40000"
		internal_address = "0.0.0.0"
		internal_group   = ""
		ssh_address      = "127.0.0.1"
		ssh_user         = "ubuntu"
	},
]

# ----------------------------------------------------------------------
