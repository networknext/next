#!/bin/bash

bucket=
artifact=
version=

print_usage() {
    printf "Usage: bootstrap.sh -b bucket_name -a artifact -v version\n\n"
    printf "b [string]\tBucket name on GCP Storage\n"
    printf "a [string]\tArtifact name on GCP Storage\n"

    printf "Example:\n\n"
    printf "> portal_bootstrap.sh -b gs://development_artifacts -a portal.dev.tar.gz \n"
}

while getopts 'b:a:h' flag; do
  case "${flag}" in
    b) bucket="${OPTARG}" ;;
    a) artifact="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

sudo apt-get update

# check for nginx and install if necessary
REQUIRED_PKG="nginx"
PKG_OK=$(dpkg-query -W --showformat='${Status}\n' $REQUIRED_PKG|grep "install ok installed")
echo Checking for $REQUIRED_PKG: $PKG_OK > /portal/dpkg.txt
if [ "" = "$PKG_OK" ]; then
  echo "No $REQUIRED_PKG. Setting up $REQUIRED_PKG." >> /portal/dpkg.txt
  sudo apt-get --yes install $REQUIRED_PKG
fi

# Copy the required files for the service from GCP Storage
gsutil cp "$bucket/$artifact" artifact.tar.gz

# Uncompress the artifact files into /app
tar -xvf artifact.tar.gz

# copy over the nginx config
rm -f /etc/nginx/sites-enabled/default
cp ./nginx/portal.nginx.conf /etc/nginx/sites-available/nn.portal
ln -s /etc/nginx/sites-available/nn.portal /etc/nginx/sites-enabled/nn.portal
cp ./nginx/digicert.conf /etc/nginx/snippets/
cp ./nginx/digicert.crt /etc/ssl/certs/
cp ./nginx/digicert.key /etc/ssl/private/
chmod 0644 /etc/ssl/private/digicert.key

systemctl restart nginx
