
extra = ""

vpn_address = "45.79.157.168"

google_credentials          = "~/secrets/terraform-staging.json"
google_location             = "US"
google_region               = "us-central1"
google_zone                 = "us-central1-a"
google_zones                = ["us-central1-a", "us-central1-b", "us-central1-c"] 	# IMPORTANT: c3 family is only available in these zones, not us-central1-f
google_artifacts_bucket     = "gs://auspicious_network_next_backend_artifacts"
google_database_bucket      = "gs://auspicious_network_next_database_files"

cloudflare_api_token        = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id          = "eba5d882ea2aa23f92dfb50fbf7e3cf4"
cloudflare_domain           = "virtualgo.net"

relay_backend_public_key    = "E/ilXdLkMb7sPSC+TZy3V0fCmbWiUJAV3DmGL6rKU2s="

server_backend_public_key   = "l9XDdCTOg5jifg5Zoz6Xrk9fUrISAflKpmRW1CjcqVI="

load_test_buyer_public_key  = "GLPP7IXT09HXFk8T1T8vBPC5Pe/YFSgWZtjPFRYFJ6klHS2b3ZNc5Q=="
load_test_buyer_private_key = "GLPP7IXT09H/IkzKXra+698dNfpoumU13z9BnTssEsikw/GHEcWnPdcWTxPVPy8E8Lk979gVKBZm2M8VFgUnqSUdLZvdk1zl"

ip2location_bucket_name     = "auspicious_network_next_staging"

relay_public_key  = "Y3L2D92gGoH/GzIT/LMEHCFSF81BGfpHGZr+MOhtOik="
relay_private_key = "yg9OZ4NEvHGjfuuz/KDAT9lPWF13/ilVMljZojP51YM="
