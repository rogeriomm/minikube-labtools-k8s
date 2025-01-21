#!/usr/bin/env zsh

PATH=$PATH:/usr/local/bin
mac_ip=$(ifconfig bridge100 | grep "inet " | awk -F ' ' '{print $2}')
minikube -p cluster2 mount --ip $mac_ip $HOME/git/dataops/bdmm/labtools-k8s:/labtools-k8s
