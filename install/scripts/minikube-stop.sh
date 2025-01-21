#!/usr/bin/env zsh

source minikube-lib.sh

rm -f "$MINIKUBE_HOME"/docker-env

minikube -p cluster stop 
minikube -p cluster2 stop 
