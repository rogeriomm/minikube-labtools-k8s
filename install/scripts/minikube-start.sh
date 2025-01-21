#!/usr/bin/env zsh

sudo -v

source minikube-lib.sh

clusters_start

sudo -v

echo "Checking Minikube DNS"

if ! dig @192.168.64.1 www.google.com; then
  echo 'Check MACOS firewall. Add "named" firewall rule and restart named'
  exit 2
fi

sudo -v

clusters_post_start

