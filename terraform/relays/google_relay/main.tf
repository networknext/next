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
variable "zone" { type = string }
variable "machine_type" { type = string }
variable "context" { type = map(string) }

# ----------------------------------------------------------------------------------------

locals {
  a = split("-", var.zone)[0]
  b = split("-", var.zone)[1]
  region = join("-", [local.a,local.b])
}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "public" {
  name         = "${replace(var.relay_name, ".", "-")}-public"
  region       = local.region
  address_type = "EXTERNAL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_address" "internal" {
  name         = "${replace(var.relay_name, ".", "-")}-internal"
  region       = local.region
  address_type = "INTERNAL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_instance" "relay" {
  name         = "${replace(var.relay_name, ".", "-")}"
  zone         = var.zone
  machine_type = var.machine_type
  network_interface {
    network_ip = google_compute_address.internal.address
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
    ssh-keys = "ubuntu:${file(var.context.ssh_public_key_file)}"
  }
  lifecycle {
    create_before_destroy = true
  }

  metadata_startup_script = <<-EOF

    # remove any old journalctl files to free up disk space (if necessary)

    journalctl --vacuum-size 10M

    # clean up old packages from apt-get to free up disk space (if necessary)

    apt autoremove -y

    # update installed packages

    apt update -y
    apt upgrade -y
    apt dist-upgrade -y
    apt autoremove -y

    # we need libcurl

    apt install libcurl3-gnutls -y
    ldconfig

    # install build essentials so we can build libsodium

    apt install build-essential -y

    # install unattended upgrades so the relay keeps up to date with security fixes

    apt install unattended-upgrades -y

    # only allow ssh from vpn address

    echo sshd: ALL > hosts.deny
    echo sshd: ${var.context.vpn_address} > hosts.allow
    mv hosts.deny /etc/hosts.deny
    mv hosts.allow /etc/hosts.allow

    # build and install libsodium optimized for this relay

    wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.18.tar.gz
    tar -zxf libsodium-1.0.18.tar.gz
    cd libsodium-1.0.18
    ./configure
    make -j
    make install
    ldconfig
    cd ~

    # download the relay binary and rename it to 'relay'

    wget https://storage.googleapis.com/${var.context.relay_artifacts_bucket}/relay-${var.context.relay_version} --no-cache
    mv relay-${var.context.relay_version} relay
    chmod +x relay

    # setup the relay environment file

    cat > relay.env <<- EOM
    RELAY_NAME=${var.relay_name}
    RELAY_PUBLIC_ADDRESS=${google_compute_address.public.address}:40000
    RELAY_INTERNAL_ADDRESS=${google_compute_address.internal.address}:40000
    RELAY_PUBLIC_KEY=${var.context.relay_public_key}
    RELAY_PRIVATE_KEY=${var.context.relay_private_key}
    RELAY_BACKEND_HOSTNAME=${var.context.relay_backend_hostname}
    RELAY_BACKEND_PUBLIC_KEY=${var.context.relay_backend_public_key}
    EOM

    # setup the relay service file

    cat > relay.service <<- EOM
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

    rm -rf /app
    mkdir /app
    mv relay /app/relay
    mv relay.env /app/relay.env
    mv relay.service /app/relay.service

    # limit maximum journalctl logs to 200MB so we don't run out of disk space

    sed -i "s/\(.*SystemMaxUse= *\).*/\SystemMaxUse=200M/" /etc/systemd/journald.conf
    systemctl restart systemd-journald

    # install the relay service so it starts on boot

    systemctl enable /app/relay.service

    # start the relay service

    systemctl start relay

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
