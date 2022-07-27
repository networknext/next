#!/bin/bash

# Size can be anything and should be picked to fit each use case (more logs = more space)
sudo journalctl --vacuum-size 200M
sudo sed -i "s/\(.*SystemMaxUse= *\).*/\SystemMaxUse=200M/" /etc/systemd/journald.conf
sudo systemctl restart systemd-journald