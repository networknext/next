#!/bin/bash

# For a reasonably chatty service, 200M will be about 2 days worth of logging (relay pusher with debug on does about 100M a day)
# For VMs and MIGs, the commands in the script should be added to each machine / templates' startup script. For relays, the commands will need to be run on install.
# The link below gives other possible values (K, M, G, etc) to tailor how many days worth of logs to save:
# https://www.freedesktop.org/software/systemd/man/journald.conf.html

# Size can be anything and should be picked to fit each use case (more logs = more space)
sudo journalctl --vacuum-size 200M
sudo sed -i "s/\(.*SystemMaxUse= *\).*/\SystemMaxUse=200M/" /etc/systemd/journald.conf
sudo systemctl restart systemd-journald