#!/usr/bin/env zsh

source minikube-lib.sh

minikube_check_config

sudo -v

clusters_start

sudo -v

echo "Checking Minikube DNS"

if ! dig @192.168.64.1 www.google.com; then
  echo 'Check MACOS firewall. Add "named" firewall rule and restart named: "sudo brew services restart bind"'
  exit 2
fi

sudo -v

clusters_post_start
