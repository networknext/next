
# dev relay variables

env                         = "dev"
vpn_address                 = "45.79.157.168"
ssh_public_key_file         = "~/.ssh/id_rsa.pub"
ssh_private_key_file        = "~/.ssh/id_rsa"
relay_version               = "relay-debug-1.0.28"
relay_artifacts_bucket      = "auspicious_network_next_relay_artifacts"
relay_backend_public_key    = "aKYacsuKannqYYbILeBPqpdIkLTp68tL1NuwuqLziwo="
relay_backend_url           = "relay-dev.virtualgo.net"

raspberry_buyer_public_key  = "g/WKV8lUxG/bFhGkh7lnoZBLqfgUWn2Lx0/Zoqg1KIxWE16r/qEPAw=="

raspberry_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]

test_buyer_public_key       = "XS1IFDqcGudRvHbJltGmfnsO78MRB4tiLHqzsHgFCbxqwgKzNesDwg=="

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
