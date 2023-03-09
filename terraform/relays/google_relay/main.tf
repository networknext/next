# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "4.51.0"
    }
  }
}

# ----------------------------------------------------------------------------------------

variable "relay_name" { type = string }
variable "region" { type = string }
variable "zone" { type = string }
variable "machine_type" { type = string }
variable "vpn_address" { type = string }
variable "ssh_public_key_file" { type = string }
variable "env" { type = string }
variable "relay_version" { type = string }
variable "relay_artifacts_bucket" { type = string }
variable "relay_public_key" { type = string }
variable "relay_private_key" { type = string }
variable "relay_backend_hostname" { type = string }
variable "relay_backend_public_key" { type = string }

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "public" {
  name         = "${replace(var.relay_name, ".", "-")}-public"
  region       = var.region
  address_type = "EXTERNAL"
}

resource "google_compute_address" "internal" {
  name         = "${replace(var.relay_name, ".", "-")}-internal"
  region       = var.region
  address_type = "INTERNAL"
}

resource "google_compute_instance" "relay" {
  name         = "relay"
  zone         = var.zone
  machine_type = var.machine_type
  network_interface {
    network_ip = google_compute_address.internal.id
    network    = "default"
    subnetwork = "default"
    access_config {
      nat_ip = google_compute_address.public.address
    }
  }
  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }
  metadata = {
    ssh-keys = "ubuntu:${file(var.ssh_public_key_file)}"
  }

  metadata_startup_script = <<-EOF

# remove any old journalctl files to free up disk space (if necessary)

sudo journalctl --vacuum-size 10M

# clean up old packages from apt-get to free up disk space (if necessary)

sudo apt autoremove -y

# update installed packages

sudo apt update -y
sudo apt upgrade -y
sudo apt dist-upgrade -y
sudo apt autoremove -y

# install build essentials so we can build libsodium

sudo apt install build-essential -y

# install unattended upgrades so the relay keeps up to date with security fixes

sudo apt install unattended-upgrades -y

# only allow ssh from vpn address

echo sshd: ALL > hosts.deny
echo sshd: ${var.vpn_address} > hosts.allow
sudo mv hosts.deny /etc/hosts.deny
sudo mv hosts.allow /etc/hosts.allow

# make the relay command line prompt cool

sudo cat >> ~/.bashrc <<- EOM
export PS1="\[\033[36m\]${var.relay_name} [${var.env}] \[\033[00m\]\w # "
EOM

sudo echo "source ~/.bashrc" >> ~/.profile.sh

# build and install libsodium optimized for this relay

wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.18.tar.gz
tar -zxf libsodium-1.0.18.tar.gz
cd libsodium-1.0.18
./configure
make -j
sudo make install
ldconfig
cd ~

# download the relay binary and rename it to 'relay'

wget https://storage.googleapis.com/${var.relay_artifacts_bucket}/relay-${var.relay_version} --no-cache
sudo mv relay-${var.relay_version} relay
sudo chmod +x relay

# setup the relay environment file

sudo cat > relay.env <<- EOM
RELAY_NAME=${var.relay_name}
RELAY_PUBLIC_ADDRESS=${google_compute_address.public.address}:40000
RELAY_INTERNAL_ADDRESS=${google_compute_address.internal.address}:40000
RELAY_PUBLIC_KEY=${var.relay_public_key}
RELAY_PRIVATE_KEY=${var.relay_private_key}
RELAY_BACKEND_HOSTNAME=${var.relay_backend_hostname}
RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
EOM

# setup the relay service file

sudo cat > relay.service <<- EOM
[Unit]
Description=Network Next Relay
ConditionPathExists=/app/relay
After=network.target

[Service]
Type=simple
LimitNOFILE=1024
WorkingDirectory=/app
ExecStart=/app/relay
EnvironmentFile=/app/relay.env
Restart=on-failure
RestartSec=12

[Install]
WantedBy=multi-user.target
EOM

# move everything into the /app dir

sudo rm -rf /app
sudo mkdir /app
sudo mv relay /app/relay
sudo mv relay.env /app/relay.env
sudo mv relay.service /app/relay.service

# limit maximum journalctl logs to 200MB so we don't run out of disk space

sudo sed -i "s/\(.*SystemMaxUse= *\).*/\SystemMaxUse=200M/" /etc/systemd/journald.conf
sudo systemctl restart systemd-journald

# install the relay service so it starts on boot

sudo systemctl enable /app/relay.service

# start the relay service

sudo systemctl start relay

EOF
}

output "public_address" {
  description = "The public IP address of the google relay"
  value = google_compute_address.public.address
}

output "internal_address" {
  description = "The internal IP address of the google relay"
  value = google_compute_address.internal.address
}

# ----------------------------------------------------------------------------------------
