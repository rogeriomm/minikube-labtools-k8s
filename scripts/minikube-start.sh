#!/usr/bin/env zsh

minikube -p cluster2 start --embed-certs

minikube -p cluster  start --embed-certs
minikube -p cluster docker-env > ~/.minikube/docker-env
source ~/.minikube/docker-env

minikube profile cluster2

./minikube-init.sh postinstall

