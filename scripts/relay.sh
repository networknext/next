cd ~
rm -f relay*
wget https://storage.googleapis.com/relay_artifacts/relay-reference-1.0.4 -O relay --no-cache
chmod +x relay
./relay version
sudo mv relay /app/relay
sudo systemctl enable /app/relay.service
sudo systemctl start relay && sudo journalctl -fu relay -n 100

cd ~
rm -f relay*
wget https://storage.googleapis.com/relay_artifacts/relay-reference-1.0.4 -O relay --no-cache
chmod +x relay
./relay version
sudo mv relay /app/relay
sudo systemctl stop relay
sudo systemctl start relay && sudo journalctl -fu relay -n 100
