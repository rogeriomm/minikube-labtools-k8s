#!/usr/bin/env zsh

PATH=$PATH:/usr/local/bin
mac_ip=$(ifconfig bridge100 | grep "inet " | awk -F ' ' '{print $2}')

# On Minikube nodes run:
# sudo mkdir -p /home/${USERNAME}
# sudo mount -t 9p -o ro -o version=9p2000.L -o dfltuid=1000 -o dfltgid=1000 -o port=8888 ${mac_ip} /home/${USERNAME}

minikube -p cluster2 mount --options fscache,ro --kill=false --ip $mac_ip --port=8888 --9p-version='9p2000.L' \
    $HOME:/home/$USERNAME

