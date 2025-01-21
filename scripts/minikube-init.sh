#!/usr/bin/env zsh

source minikube-lib.sh

if [[ "$1" = "install" ]]; then
  sudo echo -n
  are_you_sure
  init
  post_init
elif [[ "$1" = "postinstall" ]]; then
  sudo echo
  post_init
elif [[ "$1" = "argocd" ]]; then
  argocd_show_password
else
  echo "Invalid command"
fi
