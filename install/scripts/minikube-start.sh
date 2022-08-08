#!/usr/bin/env zsh

sudo -v

source minikube-lib.sh

clusters_start

sudo -v

echo "Dig www.google.com"
dig @192.168.64.1 www.google.com
echo "Check MACOS firewall. Add named firewall rule and restart named"

sudo -v

clusters_post_start

