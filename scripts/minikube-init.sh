#!/usr/bin/env zsh

# https://itnext.io/goodbye-docker-desktop-hello-minikube-3649f2a1c469
# brew install docker-credential-helper

# Last stable kuberbetes version supported by RANCHER 2.5
KUBERNETES_VERSION="1.21.6"

MINIKUBE_HOME="${MINIKUBE_HOME:-${HOME}/.minikube}"
MINIKUBE_FILES=$MINIKUBE_HOME/files

cluster1_create()
{
  PROFILE=cluster

  minikube -p $PROFILE config set cpus 14
  minikube -p $PROFILE config set memory 16g
  minikube -p $PROFILE config set disk-size 100g
  minikube -p $PROFILE config view

  minikube -p $PROFILE start --kubernetes-version="v${KUBERNETES_VERSION}" \
           --nodes 1 --driver='hyperkit' --insecure-registry "192.168.64.0/24,10.0.0.0/8"

  minikube -p cluster docker-env > "$MINIKUBE_HOME"/docker-env

  source "$MINIKUBE_HOME"/docker-env
}

cluster2_create()
{
  PROFILE=cluster2

  minikube -p $PROFILE config set cpus 28
  minikube -p $PROFILE config set memory 80g
  minikube -p $PROFILE config set disk-size 100g
  minikube -p $PROFILE config view

  minikube -p $PROFILE start --kubernetes-version="v${KUBERNETES_VERSION}" \
           --nodes 2 --driver='hyperkit' --insecure-registry "192.168.64.0/24,10.0.0.0/8"

  minikube -p $PROFILE addons enable ingress
  minikube -p $PROFILE addons enable ingress-dns
  minikube -p $PROFILE addons enable registry
  minikube -p $PROFILE addons enable registry-aliases
  minikube -p $PROFILE addons enable dashboard
  minikube -p $PROFILE addons enable metrics-server
  minikube -p $PROFILE addons disable registry-creds
  minikube -p $PROFILE addons enable metallb

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

  while : ; do
    kubectl -n argocd get secret/argocd-initial-admin-secret 2> /dev/null > /dev/null && break
    sleep 20
  done
  set +x
  echo -n "ARGOCD admin password: "
  kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
  echo ""
  set -x
}

internal_registry_setup()
{
  # Configure docker internal registry
  kubectx cluster # Our docker engine used by MAC OS
  kubectl -n kube-system delete configmap registry-cluster || echo -n
  kubectl -n kube-system create configmap registry-cluster \
                  --from-literal=registryAliases=registry.minikube \
                  --from-literal=registryServiceHost=$(minikube -p cluster2 ip) # Internal registry
  kubectl -n kube-system apply -f node-etc-hosts-update.yaml
}

# https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s/
rancher_setup()
{
  helm repo add rancher-stable https://releases.rancher.com/server-charts/stable
  helm repo update
  kubectl create namespace cattle-system
  helm install rancher rancher-stable/rancher \
     --namespace cattle-system \
     --set hostname=rancher.world.xpt \
     --set replicas=2 \
     --set ingress.tls.source=secret
}

rancher_show_password()
{
  echo -n "Rancher password: "
  kubectl get secret --namespace cattle-system bootstrap-secret \
     -o go-template='{{.data.bootstrapPassword|base64decode}}{{"\n"}}'
}

copy_cert()
{
  mkdir -p "$MINIKUBE_HOME"/certs
  cp world.xpt.pem "$MINIKUBE_HOME"/certs
}

are_you_sure()
{
  read -q "REPLY?Initialize Minikube(y/n)? "
  echo "\n"

  if [ "$REPLY" = "n" ]; then
    exit
  fi
}

init()
{
  echo "Installation..."
  set -x
  set -e

  # Delete all clusters
  minikube -p cluster delete
  minikube -p cluster2 delete

  # Copy certificate
  copy_cert

  # Create cluster with 1 node
  cluster1_create

  # Create cluster with 2 nodes
  cluster2_create

  # Setup internal registry
  internal_registry_setup

  # Set current minikube profile
  minikube profile cluster2
  kubectx cluster2

  kubectl apply -f persistent-volumes.yaml

  # Setup Argocd
  argocd_setup
}

# //192.168.0.201/share /share cifs  credentials=/home/docker/.smbcredentials 0 0
# sudo mount.cifs "\\\\192.168.0.201\share" -o user=rogermm,pass=password /data/share
post_init()
{
  echo "Post installation..."
  set -x
  rancher_setup
  python3 minikube-init.py
  argocd_show_password
  rancher_show_password
}

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
