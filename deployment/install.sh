#!/bin/bash

set -euo pipefail

DUCKDNS_DOMAIN="${DUCKDNS_DOMAIN:-}"
DUCKDNS_TOKEN="${DUCKDNS_TOKEN:-}"
CERT_INSTALL_DIR="/etc/ssl/letsencrypt"
CERT_DOMAIN="${DUCKDNS_DOMAIN}.duckdns.org"

# Ensure DUCKDNS_TOKEN is set
if [[ -z "${DUCKDNS_TOKEN}" ]]; then
  echo "❌ DUCKDNS_TOKEN is not set. Please export it and try again."
  exit 1
fi

# Ensure DUCKDNS_TOKEN is set
if [[ -z "${DUCKDNS_DOMAIN}" ]]; then
  echo "❌ DUCKDNS_DOMAIN is not set. Please export it and try again."
  exit 1
fi


function install_dependencies() {
  echo "📦 Installing dependencies..."
  sudo apt update -y
  sudo DEBIAN_FRONTEND=noninteractive apt upgrade -y
  sudo apt install -y curl tmux git
  echo "✅ Dependencies installed"
}

function install_docker() {
	if command -v docker &> /dev/null; then
    echo "✅ Docker is already installed: $(docker --version)"
  else
    echo "🐳 Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    echo "✅ Docker installed: $(docker --version)"
  fi
}

function cert_exists() {
  [[ -f "${CERT_INSTALL_DIR}/fullchain.pem" ]] && [[ -f "${CERT_INSTALL_DIR}/privkey.pem" ]]
}

function setup_certificates() {
  echo "🔐 Setting up SSL certificates..."

  if cert_exists; then
    echo "✅ Valid certificates already exist at ${CERT_INSTALL_DIR}. Skipping issuance."
    return
  fi

  echo "🌐 Updating DuckDNS IP..."
  curl -k "https://www.duckdns.org/update?domains=${DUCKDNS_DOMAIN}&token=${DUCKDNS_TOKEN}&ip=" -o /tmp/duck.log

  echo "📥 Installing acme.sh..."
  curl https://get.acme.sh | sh
  export PATH=~/.acme.sh:$PATH
  export DuckDNS_Token="${DUCKDNS_TOKEN}"

  echo "📄 Issuing certificate for ${CERT_DOMAIN}..."
  ~/.acme.sh/acme.sh --issue \
    --dns dns_duckdns \
    -d "${CERT_DOMAIN}" \
    --keylength ec-256 \
    --force

  echo "📂 Installing certs to ${CERT_INSTALL_DIR}..."
  sudo mkdir -p "${CERT_INSTALL_DIR}"

  ~/.acme.sh/acme.sh --install-cert \
    -d "${CERT_DOMAIN}" \
    --ecc \
    --key-file "${CERT_INSTALL_DIR}/privkey.pem" \
    --fullchain-file "${CERT_INSTALL_DIR}/fullchain.pem"

  echo "🔁 Enabling auto-renewal..."
  ~/.acme.sh/acme.sh --upgrade --auto-upgrade

  echo "✅ Certificates installed at ${CERT_INSTALL_DIR}"
}

function run_vpainless() {
  echo "🚀 Running Vpainless..."
  echo "☁️ Cloning Vpainless Repo"
	git clone https://github.com/vpainless/vpainless.git
  cd vpainless

  echo "🔧 Building Docker images..."
  docker build --network=host --build-arg PROD="https://${CERT_DOMAIN}" -f ./frontend/Dockerfile -t vpainless-front:latest .
  docker build --network=host -t vpainless-server:latest ./backend
  echo "✅ Docker images built"

	cd deployment
  mkdir -p data
  cp -r ../backend/internal/pkg/db/migrations .

  echo "🔑 Generating SSH keys..."
  ssh-keygen -b 4096 -f key -N ""

  echo "⚙️ Configuring nginx.conf..."
  sed -i "s|{DUCKDNS_DOMAIN}|${DUCKDNS_DOMAIN}|g" nginx.conf

  echo "📦 Starting Docker Compose..."
  docker compose up -d
  echo "✅ Vpainless is up and running!"
}

install_dependencies
install_docker
setup_certificates
run_vpainless
