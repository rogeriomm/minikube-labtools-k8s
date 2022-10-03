#!/usr/bin/env zsh

source minikube-lib.sh

if [[ "$1" = "install" ]]; then
  sudo -v
  are_you_sure
  init
elif [[ "$1" = "postinstall" ]]; then
  sudo -v
  post_init
elif [[ "$1" = "argocd" ]]; then
  argocd_show_password
else
  echo "Invalid command"
fi
