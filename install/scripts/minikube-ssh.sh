#!/usr/bin/env zsh

source minikube-lib.sh

minikube profile cluster2

if [[ "$1" = "1" ]]; then
  minikube --node=cluster2 ssh
else
  minikube --node=cluster2-m0"$1" ssh
fi

