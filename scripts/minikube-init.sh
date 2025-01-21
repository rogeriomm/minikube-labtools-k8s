#!/usr/bin/env zsh

cluster1_create()
{
  PROFILE=cluster

  minikube -p $PROFILE config set cpus 14
  minikube -p $PROFILE config set memory 16g
  minikube -p $PROFILE config set disk-size 100g
  minikube -p $PROFILE config view

  minikube -p $PROFILE start --nodes 1 --driver='hyperkit' --insecure-registry "192.168.64.0/24,10.0.0.0/8"

  minikube -p cluster docker-env > ~/.minikube/docker-env

  source ~/.minikube/docker-env
}

cluster2_create()
{
  PROFILE=cluster2

  minikube -p $PROFILE config set cpus 28
  minikube -p $PROFILE config set memory 80g
  minikube -p $PROFILE config set disk-size 100g
  minikube -p $PROFILE config view

  minikube -p $PROFILE start --nodes 2 --driver='hyperkit' --insecure-registry "192.168.64.0/24,10.0.0.0/8"

  minikube -p $PROFILE addons enable ingress
  minikube -p $PROFILE addons enable ingress-dns
  minikube -p $PROFILE addons enable registry
  minikube -p $PROFILE addons enable registry-aliases
  minikube -p $PROFILE addons enable dashboard
  minikube -p $PROFILE addons enable metrics-server
  minikube -p $PROFILE addons disable registry-creds

  minikube -p $PROFILE addons list
}

argocd_setup()
{
  kubectx cluster2
  kubectl create namespace argocd
  kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
  kubectl apply -f argocd-ingress.yaml
}

argocd_show_password()
{
  kubectx cluster2
  echo "ARGOCD admin password:"
  kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
}

internal_registry_setup()
{
  # Configure docker internal registry
  kubectx cluster # Our docker engine used by MAC OS
  kubectl -n kube-system delete configmap registry-cluster
  kubectl -n kube-system create configmap registry-cluster \
                  --from-literal=registryAliases=registry.minikube \
                  --from-literal=registryServiceHost=$(minikube -p cluster2 ip) # Internal registry
  kubectl -n kube-system apply -f node-etc-hosts-update.yaml
}

are_you_sure()
{
  read -q "REPLY?Initialize Minikube?"
  echo "\n"

  if [ "$REPLY" = "n" ]; then
    exit
  fi
}

are_you_sure

set -x

# Delete all clusters
minikube -p cluster delete
minikube -p cluster2 delete

# Create cluster with 1 node
cluster1_create

# Create cluster with 2 nodes
cluster2_create

# Setup internal registry
internal_registry_setup

# Set current minikube profile
minikube profile cluster2
kubectx cluster2

# Setup Argocd
argocd_setup
argocd_show_password
