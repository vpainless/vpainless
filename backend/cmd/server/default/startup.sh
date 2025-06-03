#!/bin/bash
set -e;
apt update -y;
DEBIAN_FRONTEND=noninteractive apt upgrade -y;
apt install curl tmux -y;
bash -c "$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)" @ install --beta -u root;
ufw allow ssh;
ufw allow http;
ufw allow https;
ufw enable;
