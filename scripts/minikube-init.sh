#!/usr/bin/env zsh

create_cluster1()
{
  PROFILE=cluster

  minikube -p $PROFILE config set cpus 14
  minikube -p $PROFILE config set memory 16g
  minikube -p $PROFILE config set disk-size 100g
  minikube -p $PROFILE config view

  minikube -p $PROFILE start --nodes 1 --driver='hyperkit' --insecure-registry "192.168.64.0/24,10.0.0.0/8"

  minikube -p $PROFILE addons enable registry

  minikube -p cluster docker-env > ~/.minikube/docker-env

  source ~/.minikube/docker-env
}

create_cluster2()
{
  PROFILE=cluster2

  minikube -p $PROFILE config set cpus 28
  minikube -p $PROFILE config set memory 80g
  minikube -p $PROFILE config set disk-size 100g
  minikube -p $PROFILE config view

  minikube -p $PROFILE start --nodes 2 --driver='hyperkit' --insecure-registry "192.168.64.0/24,10.0.0.0/8"

  minikube -p $PROFILE addons enable ingress
  minikube -p $PROFILE addons enable ingress-dns
  minikube -p $PROFILE addons disable registry
  minikube -p $PROFILE addons disable registry-aliases
  minikube -p $PROFILE addons enable dashboard
  minikube -p $PROFILE addons enable metrics-server
  minikube -p $PROFILE addons disable registry-creds

  minikube -p $PROFILE addons list
}

create_argocd()
{
  kubectl create namespace argocd
  kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
}

# Create cluster with 1 node
create_cluster1

# Create cluster with 2 nodes
create_cluster2

# Set current minikube profile
minikube profile cluster2

# Setup Argocd
create_argocd
