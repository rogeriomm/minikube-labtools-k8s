#!/usr/bin/env zsh

source minikube-lib.sh

# On Minikube nodes run:
# sudo mkdir -p /home/${USERNAME}
# sudo mount -t 9p -o ro -o version=9p2000.L -o dfltuid=1000 -o dfltgid=1000 -o port=8888 ${mac_ip} /home/${USERNAME}

minikube -p cluster2 mount --options fscache,ro --kill=false --ip $(minikube_get_host_ip) --port=8888 --9p-version='9p2000.L' \
    $HOME:$HOME




