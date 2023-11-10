
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-debug-1.0.28"
relay_artifacts_bucket      = "memento_network_next_relay_artifacts"
relay_backend_public_key    = "0Ad0T86N1X8uRF4XnusUbM19k6Srnk/BxihgZT0Pvlo="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "Df/gqFQMDkNp8GH28WjVEl4I6aB5KxsVipZKow0PhCmQIQIMEeUkOg=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "AN+VWuqgAQfd5d+QeT6apNf+Gbb9rqFBMtk5M+GoMvwS1Eqz764X/A=="

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
	"VULTR"  = "vultr"
}
